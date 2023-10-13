package config

import (
	"net"
	"strconv"
	"strings"
)

type Address string

func NewAddress(proto, host, port string) Address {
	var sb strings.Builder
	if proto != "" {
		sb.WriteString(strings.TrimSuffix(proto, "://"))
		sb.WriteString("://")
	}
	sb.WriteString(host)
	if port != "" {
		sb.WriteString(":")
		sb.WriteString(strings.TrimPrefix(port, ":"))
	}
	return Address(sb.String())
}

func (a Address) Splitted() (proto, host, port string) {
	s := string(a)
	host = s
	if i := strings.Index(s, "://"); i > 0 {
		proto = s[:i]
		host = s[i+3:]
	}
	if h, p, err := net.SplitHostPort(host); err == nil {
		host = h
		port = p
	}
	return proto, host, port
}

func (a Address) Scheme() string {
	proto, _, _ := a.Splitted()
	return proto
}

func (a Address) Host() string {
	_, host, _ := a.Splitted()
	return host
}

func (a Address) Port() string {
	_, _, port := a.Splitted()
	return port
}

func (a Address) HostPort() string {
	_, host, port := a.Splitted()
	if port == "" {
		return host
	}
	return net.JoinHostPort(host, port)
}

func (a Address) PortInt() int {
	port, _ := strconv.Atoi(a.Port())
	return port
}

func (a Address) String() string {
	return string(a)
}

type HttpAddress Address

func (a HttpAddress) Splitted() (proto, host, port string) {
	proto, host, port = Address(a).Splitted()
	if proto == "" {
		proto = "http"
	}
	return proto, host, port
}

func (a HttpAddress) Scheme() string {
	proto, _, _ := a.Splitted()
	return proto
}
func (a HttpAddress) Host() string {
	return Address(a).Host()
}

func (a HttpAddress) Port() string {
	return Address(a).Port()
}

func (a HttpAddress) HostPort() string {
	return Address(a).HostPort()
}

func (a HttpAddress) PortInt() int {
	return Address(a).PortInt()
}

func (a HttpAddress) String() string {
	return string(a)
}
