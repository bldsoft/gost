package inhouse

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/log"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/memberlist"
)

const (
	// udpPacketBufSize is used to buffer incoming packets during read
	// operations.
	udpPacketBufSize = 65536

	// udpRecvBufSize is a large buffer size that we attempt to set UDP
	// sockets to in order to handle a large volume of messages.
	udpRecvBufSize = 2 * 1024 * 1024
)

var wsUpgrader = websocket.Upgrader{} // use default options

// Transport is used to abstract over communicating with other peers. The packet
// interface is assumed to be best-effort and the stream interface is assumed to
// be reliable.
type Transport struct {
	packetCh chan *memberlist.Packet
	streamCh chan net.Conn

	endpointPath string

	wg          sync.WaitGroup
	udpListener *net.UDPConn

	shutdown atomic.Int32
}

func NewTransport(bindAddress config.Address) (*Transport, error) {
	t := &Transport{
		packetCh:     make(chan *memberlist.Packet, 10),
		streamCh:     make(chan net.Conn, 10),
		endpointPath: "/api/discovery/in-house",
	}

	var ok bool
	// Clean up listeners if there's an error.
	defer func() {
		if !ok {
			t.Shutdown()
		}
	}()

	// Build the UDP listener
	addr, port := bindAddress.Host(), bindAddress.PortInt()
	ip := net.ParseIP(addr)
	udpAddr := &net.UDPAddr{IP: ip, Port: port}
	var err error
	t.udpListener, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to start UDP listener on %q port %d: %w", addr, port, err)
	}
	if err := setUDPRecvBuf(t.udpListener); err != nil {
		return nil, fmt.Errorf("failed to resize UDP buffer: %w", err)
	}

	t.wg.Add(1)
	go t.udpListen(t.udpListener)

	ok = true
	return t, nil
}

func (t *Transport) udpListen(udpLn *net.UDPConn) {
	defer t.wg.Done()
	for {
		// Do a blocking read into a fresh buffer. Grab a time stamp as
		// close as possible to the I/O.
		buf := make([]byte, udpPacketBufSize)
		n, addr, err := udpLn.ReadFrom(buf)
		ts := time.Now()
		if err != nil {
			if s := t.shutdown.Load(); s == 1 {
				break
			}

			log.Logger.Errorf("Discovery: memberlist: error reading UDP packet: %v", err)
			continue
		}

		// Check the length - it needs to have at least one byte to be a
		// proper message.
		if n < 1 {
			log.Logger.Errorf("Discovery: memberlist: UDP packet too short (%d bytes) %s", len(buf), addr.String())
			continue
		}

		// Ingest the packet.
		t.packetCh <- &memberlist.Packet{
			Buf:       buf[:n],
			From:      addr,
			Timestamp: ts,
		}
	}
}

// FinalAdvertiseAddr is given the user's configured values (which
// might be empty) and returns the desired IP and port to advertise to
// the rest of the cluster.
func (t *Transport) FinalAdvertiseAddr(ip string, port int) (net.IP, int, error) {
	// If they've supplied an address, use that.
	advertiseAddr := net.ParseIP(ip)
	if advertiseAddr == nil {
		return nil, 0, fmt.Errorf("failed to parse advertise address %q", ip)
	}

	// Ensure IPv4 conversion if necessary.
	if ip4 := advertiseAddr.To4(); ip4 != nil {
		advertiseAddr = ip4
	}
	return advertiseAddr, port, nil
}

// WriteTo is a packet-oriented interface that fires off the given
// payload to the given address in a connectionless fashion. This should
// return a time stamp that's as close as possible to when the packet
// was transmitted to help make accurate RTT measurements during probes.
//
// This is similar to net.PacketConn, though we didn't want to expose
// that full set of required methods to keep assumptions about the
// underlying plumbing to a minimum. We also treat the address here as a
// string, similar to Dial, so it's network neutral, so this usually is
// in the form of "host:port".
func (t *Transport) WriteTo(b []byte, addr string) (time.Time, error) {
	a := memberlist.Address{Addr: addr, Name: ""}
	return t.WriteToAddress(b, a)
}

func (t *Transport) WriteToAddress(b []byte, a memberlist.Address) (time.Time, error) {
	addr := a.Addr

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return time.Time{}, err
	}

	// We made sure there's at least one UDP listener, so just use the
	// packet sending interface on the first one. Take the time after the
	// write call comes back, which will underestimate the time a little,
	// but help account for any delays before the write occurs.
	_, err = t.udpListener.WriteTo(b, udpAddr)
	return time.Now(), err
}

// PacketCh returns a channel that can be read to receive incoming
// packets from other peers. How this is set up for listening is left as
// an exercise for the concrete transport implementations.
func (t *Transport) PacketCh() <-chan *memberlist.Packet {
	return t.packetCh
}

// DialTimeout is used to create a connection that allows us to perform
// two-way communication with a peer. This is generally more expensive
// than packet connections so is used for more infrequent operations
// such as anti-entropy or fallback probes if the packet-oriented probe
// failed.
func (t *Transport) DialTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	dialer := websocket.Dialer{HandshakeTimeout: timeout}
	u := url.URL{Scheme: "ws", Host: addr, Path: t.endpointPath}
	c, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return c.UnderlyingConn(), nil
}

func (t *Transport) DialAddressTimeout(a memberlist.Address, timeout time.Duration) (net.Conn, error) {
	return t.DialTimeout(a.Addr, timeout)
}

// StreamCh returns a channel that can be read to handle incoming stream
// connections from other peers. How this is set up for listening is
// left as an exercise for the concrete transport implementations.
func (t *Transport) StreamCh() <-chan net.Conn {
	return t.streamCh
}

// Shutdown is called when memberlist is shutting down; this gives the
// transport a chance to clean up any listeners.
func (t *Transport) Shutdown() error {
	t.shutdown.Store(1)

	t.udpListener.Close()

	// Block until the listener has died.
	t.wg.Wait()
	return nil
}

// setUDPRecvBuf is used to resize the UDP receive window. The function
// attempts to set the read buffer to `udpRecvBuf` but backs off until
// the read buffer can be set.
func setUDPRecvBuf(c *net.UDPConn) error {
	size := udpRecvBufSize
	var err error
	for size > 0 {
		if err = c.SetReadBuffer(size); err == nil {
			return nil
		}
		size = size / 2
	}
	return err
}

func (t *Transport) Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	c, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "Discovery: ws: failed to upgrade")
		return
	}
	t.streamCh <- c.UnderlyingConn()
}

func (t *Transport) Mount(r chi.Router) {
	r.HandleFunc(t.endpointPath, t.Handler)
}
