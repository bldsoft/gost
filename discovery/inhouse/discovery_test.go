package inhouse

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/discovery"
	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/memberlist"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubMemberList struct {
	leaveTimeout time.Duration
	left         bool
	shutdown     bool
}

func (s *stubMemberList) Leave(timeout time.Duration) error {
	s.left = true
	s.leaveTimeout = timeout
	return nil
}

func (s *stubMemberList) Shutdown() error {
	s.shutdown = true
	return nil
}

func (s *stubMemberList) UpdateNode(time.Duration) error { return nil }
func (s *stubMemberList) LocalNode() *memberlist.Node    { return nil }
func (s *stubMemberList) Join([]string) (int, error)     { return 0, nil }

func TestMemberlistConfigClearsUnspecifiedAdvertise(t *testing.T) {
	d := NewDiscovery(server.Config{
		ServiceName:        "svc",
		ServiceInstance:    "svc-1",
		ServiceAddress:     "0.0.0.0:3000",
		ServiceBindAddress: "127.0.0.1:0",
	}, Config{
		Embedded:               false,
		BindAddress:            "127.0.0.1:0",
		SecretKey:              "0123456789abcdef",
		DeregisterServiceAfter: time.Hour,
	})

	cfg, err := d.memberlistConfig()
	require.NoError(t, err)
	assert.Equal(t, "", cfg.AdvertiseAddr)
	assert.Equal(t, cfg.ProbeInterval, cfg.DeadNodeReclaimTime)
}

func TestFinalAdvertiseAddrResolvesUnspecified(t *testing.T) {
	tr, err := NewTransport("127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = tr.Shutdown() })

	ip, port, err := tr.FinalAdvertiseAddr("0.0.0.0", 3000)
	require.NoError(t, err)
	require.NotNil(t, ip)
	assert.False(t, ip.IsUnspecified())
	assert.Equal(t, 3000, port)
	assert.True(t, ip.Equal(net.ParseIP("127.0.0.1")), "expected bind IP 127.0.0.1, got %v", ip)
}

func TestFinalAdvertiseAddrEmptyUsesBindOrPrivateIP(t *testing.T) {
	tr, err := NewTransport("0.0.0.0:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = tr.Shutdown() })

	ip, port, err := tr.FinalAdvertiseAddr("", 3001)
	if err != nil {
		assert.Contains(t, err.Error(), "private IP")
		return
	}
	require.NotNil(t, ip)
	assert.False(t, ip.IsUnspecified())
	assert.Equal(t, 3001, port)
}

func TestNotifyLeaveUnknownServiceNoPanic(t *testing.T) {
	d := &Discovery{
		services:                  make(map[string]*discovery.ServiceInfo),
		instanceIDToDownTimestamp: make(map[instanceKey]time.Time),
	}

	meta, err := json.Marshal(discovery.ServiceInstanceInfo{
		ServiceName: "unknown",
		ID:          "id-1",
		Address:     "127.0.0.1:3000",
	})
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		d.NotifyLeave(&memberlist.Node{Meta: meta, State: memberlist.StateLeft})
	})
}

func TestNotifyLeaveUnlocksBeforeEvent(t *testing.T) {
	d := &Discovery{
		BaseDiscovery:             discovery.NewBaseDiscovery(server.Config{ServiceName: "svc", ServiceInstance: "svc-1"}),
		services:                  make(map[string]*discovery.ServiceInfo),
		instanceIDToDownTimestamp: make(map[instanceKey]time.Time),
	}
	inst := discovery.ServiceInstanceInfo{
		ServiceName: "svc",
		ID:          "id-1",
		Address:     "127.0.0.1:3000",
		Healthy:     true,
	}
	d.services["svc"] = &discovery.ServiceInfo{Name: "svc", Instances: []discovery.ServiceInstanceInfo{inst}}

	var wg sync.WaitGroup
	wg.Add(1)
	d.Subscribe(discovery.NewEventHandler(func(ctx context.Context, _ discovery.ServiceInstanceInfo) {
		defer wg.Done()
		_, err := d.Services(ctx)
		assert.NoError(t, err)
	}).EventType(discovery.EventTypeDown))

	meta, err := json.Marshal(inst)
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		d.NotifyLeave(&memberlist.Node{Meta: meta, State: memberlist.StateDead})
		close(done)
	}()

	select {
	case <-done:
		wg.Wait()
	case <-time.After(2 * time.Second):
		t.Fatal("NotifyLeave deadlocked holding lock across TriggerEvent")
	}

	d.servicesMtx.RLock()
	defer d.servicesMtx.RUnlock()
	assert.False(t, d.services["svc"].Instances[0].Healthy)
}

func TestParseMetaRejectsEmptyIdentity(t *testing.T) {
	d := &Discovery{}
	meta, err := json.Marshal(discovery.ServiceInstanceInfo{Address: "127.0.0.1:1"})
	require.NoError(t, err)
	_, err = d.parseMeta(&memberlist.Node{Meta: meta})
	require.Error(t, err)
}

func TestNodeMetaRespectsLimit(t *testing.T) {
	d := &Discovery{
		BaseDiscovery: discovery.NewBaseDiscovery(server.Config{
			ServiceName:     "svc",
			ServiceInstance: "svc-1",
			ServiceAddress:  config.Address("127.0.0.1:3000"),
		}),
	}
	ok := d.NodeMeta(512)
	require.NotNil(t, ok)

	d.ServiceInfo.Meta = map[string]string{"k": string(make([]byte, 600))}
	assert.Nil(t, d.NodeMeta(512))
	assert.NotNil(t, d.NodeMeta(len(ok)+4000))
}

func TestStopLeavesAndCancelsRunner(t *testing.T) {
	d := NewDiscovery(server.Config{
		ServiceName:        "svc",
		ServiceInstance:    "svc-1",
		ServiceAddress:     "127.0.0.1:3000",
		ServiceBindAddress: "127.0.0.1:0",
	}, Config{
		Embedded:               false,
		BindAddress:            "127.0.0.1:0",
		SecretKey:              "0123456789abcdef",
		DeregisterServiceAfter: time.Hour,
	})

	stub := &stubMemberList{}
	d.list = stub

	stopped := make(chan struct{})
	d.AsyncRunner = server.NewAsyncJob(nil, func(ctx context.Context) error {
		close(stopped)
		return nil
	})

	require.NoError(t, d.Stop(context.Background()))
	assert.True(t, stub.left)
	assert.Equal(t, defaultLeaveTimeout, stub.leaveTimeout)
	assert.True(t, stub.shutdown)

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("expected embedded AsyncRunner.Stop to be called")
	}
}

func TestStopNilListCancelsRunner(t *testing.T) {
	d := NewDiscovery(server.Config{
		ServiceName:        "svc",
		ServiceInstance:    "svc-1",
		ServiceAddress:     "127.0.0.1:3000",
		ServiceBindAddress: "127.0.0.1:0",
	}, Config{
		Embedded:               false,
		BindAddress:            "127.0.0.1:0",
		SecretKey:              "0123456789abcdef",
		DeregisterServiceAfter: time.Hour,
	})

	stopped := make(chan struct{})
	d.AsyncRunner = server.NewAsyncJob(nil, func(ctx context.Context) error {
		close(stopped)
		return nil
	})

	require.NoError(t, d.Stop(context.Background()))
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("expected embedded AsyncRunner.Stop to be called")
	}
}
