package inhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bldsoft/gost/discovery"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/memberlist"
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slices"
)

const joinRetryInterval = 10 * time.Second

type instanceKey struct {
	serviceName string
	instanceID  string
}

type Discovery struct {
	discovery.BaseDiscovery
	server.AsyncRunner

	cfg  Config
	list *memberlist.Memberlist

	services                  map[string]*discovery.ServiceInfo
	instanceIDToDownTimestamp map[instanceKey]time.Time
	servicesMtx               sync.RWMutex

	transport *Transport
}

func NewDiscovery(serviceCfg server.Config, cfg Config) *Discovery {
	d := &Discovery{
		cfg:                       cfg,
		BaseDiscovery:             discovery.NewBaseDiscovery(serviceCfg),
		services:                  make(map[string]*discovery.ServiceInfo),
		instanceIDToDownTimestamp: make(map[instanceKey]time.Time),
	}

	d.AsyncRunner = server.NewContextAsyncRunner(d.run)

	if d.cfg.Embedded {
		var err error
		d.transport, err = NewTransport(serviceCfg.ServiceBindAddress)
		if err != nil {
			panic(err)
		}
	}

	return d
}

func (d *Discovery) memberlistConfig() (*memberlist.Config, error) {
	memberlistCfg := memberlist.DefaultLANConfig()
	memberlistCfg.LogOutput = logOutput{}
	memberlistCfg.Name = d.ServiceInfo.ID
	memberlistCfg.BindAddr = d.cfg.BindAddress.Host()
	memberlistCfg.BindPort = d.cfg.BindAddress.PortInt()
	memberlistCfg.AdvertiseAddr = d.ServiceInfo.Address.Host()
	memberlistCfg.AdvertisePort = d.ServiceInfo.Address.PortInt()
	memberlistCfg.SecretKey = []byte(d.cfg.SecretKey.String())
	if d.transport != nil {
		memberlistCfg.Transport = d.transport
	}
	memberlistCfg.Delegate = d
	memberlistCfg.Events = d
	return memberlistCfg, nil
}

func (d *Discovery) run(ctx context.Context) error {
	var err error
	cfg, err := d.memberlistConfig()
	if err != nil {
		return err
	}

	if d.transport != nil {
		go d.transport.Run()
	}

	d.list, err = memberlist.Create(cfg)
	if err != nil {
		return fmt.Errorf("failed to create memberlist: %w", err)
	}
	d.join(ctx, true, d.cfg.ClusterMembers...)

	checkExpiredInterval := min(d.cfg.DeregisterServiceAfter/2, 5*time.Minute)
	t := time.NewTicker(checkExpiredInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			d.deregisterExpired()
		case <-ctx.Done():
			return nil
		}
	}
}

func (d *Discovery) join(ctx context.Context, logError bool, members ...string) {
	if len(members) == 0 {
		return
	}

	log.Logger.InfoWithFields(log.Fields{"members": members}, "Discovery: joining")

	t := time.NewTicker(joinRetryInterval)
	defer t.Stop()
	for {
		_, err := d.list.Join(members)
		if err == nil {
			break
		}
		if logError {
			log.Errorf("Discovery: failed to join cluster: %s", strings.TrimSpace(err.Error()))
		}

		select {
		case <-t.C:
		case <-ctx.Done():
			return
		}

	}
	log.Logger.InfoWithFields(log.Fields{"members": members}, "Discovery: joined")
}

func (d *Discovery) instanceKey(si *discovery.ServiceInstanceInfo) instanceKey {
	return instanceKey{si.ServiceName, si.ID}
}

func (d *Discovery) TriggerEventCtx(ctx context.Context, eventType discovery.EventType, instance discovery.ServiceInstanceInfo) {
	switch eventType {
	case discovery.EventTypeDown:
		d.instanceIDToDownTimestamp[d.instanceKey(&instance)] = time.Now()
	case discovery.EventTypeDiscovered:
		fallthrough
	case discovery.EventTypeUp:
		fallthrough
	case discovery.EventTypeRemoved:
		delete(d.instanceIDToDownTimestamp, d.instanceKey(&instance))
	}
	d.BaseDiscovery.TriggerEventCtx(ctx, eventType, instance)
}

func (d *Discovery) TriggerEvent(eventType discovery.EventType, instance discovery.ServiceInstanceInfo) {
	d.TriggerEventCtx(context.Background(), eventType, instance)
}

func (d *Discovery) deregisterExpired() {
	d.servicesMtx.Lock()
	defer d.servicesMtx.Unlock()
	for instanceKey, ts := range d.instanceIDToDownTimestamp {
		if time.Since(ts) > d.cfg.DeregisterServiceAfter {
			service := d.services[instanceKey.serviceName]

			i := slices.IndexFunc(service.Instances, func(sii discovery.ServiceInstanceInfo) bool {
				return sii.ID == instanceKey.instanceID
			})
			instance := service.Instances[i]

			service.Instances = slices.Delete(service.Instances, i, i+1)
			if len(service.Instances) == 0 {
				delete(d.services, instanceKey.serviceName)
			}

			d.TriggerEvent(discovery.EventTypeRemoved, instance)
		}
	}
}

func (d *Discovery) addService(node *memberlist.Node, withLock bool) {
	meta, err := d.parseMeta(node)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if withLock {
		d.servicesMtx.Lock()
		defer d.servicesMtx.Unlock()
	}
	serviceInfo, ok := d.services[meta.ServiceName]
	if !ok {
		serviceInfo = &discovery.ServiceInfo{Name: meta.ServiceName}
		d.services[meta.ServiceName] = serviceInfo
	}
	meta.Healthy = true
	i := slices.IndexFunc(serviceInfo.Instances, func(si discovery.ServiceInstanceInfo) bool {
		return si.ID == meta.ID
	})
	if i >= 0 {
		switch {
		case !serviceInfo.Instances[i].Healthy && meta.Healthy:
			d.TriggerEvent(discovery.EventTypeUp, *meta)
		case serviceInfo.Instances[i].Healthy && !meta.Healthy:
			d.TriggerEvent(discovery.EventTypeDown, *meta)
		}

		serviceInfo.Instances[i] = *meta
		log.Logger.InfoWithFields(log.Fields{"service": meta}, "Discovery: service updated")
	} else {
		serviceInfo.Instances = append(serviceInfo.Instances, *meta)

		d.TriggerEvent(discovery.EventTypeDiscovered, *meta)
		d.TriggerEvent(discovery.EventTypeUp, *meta)

		log.Logger.InfoWithFields(log.Fields{"service": meta}, "Discovery: new service added")
	}
}

func (d *Discovery) Stop(ctx context.Context) error {
	if deadline, ok := ctx.Deadline(); ok {
		if timeout := time.Until(deadline); timeout > 0 {
			err := d.list.Leave(timeout)
			log.Logger.InfoOrError(err, "Discovery: leaving from the cluster")
		}
	}
	return d.list.Shutdown()
}

func (d *Discovery) Services(ctx context.Context) ([]*discovery.ServiceInfo, error) {
	d.servicesMtx.RLock()
	defer d.servicesMtx.RUnlock()
	res := make([]*discovery.ServiceInfo, 0, len(d.services))
	for _, s := range d.services {
		res = append(res, s)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return res, nil
}

func (d *Discovery) ServiceByName(ctx context.Context, name string) (*discovery.ServiceInfo, error) {
	d.servicesMtx.RLock()
	defer d.servicesMtx.RUnlock()
	s, ok := d.services[name]
	if !ok {
		return nil, discovery.NotFound
	}
	return s, nil
}

func (d *Discovery) Mount(r chi.Router) {
	if d.transport != nil {
		d.transport.Mount(r)
	}
}

// Delegates

// Limit argument is currently ignored.
// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (d *Discovery) NodeMeta(limit int) []byte {
	res, err := json.Marshal(d.BaseDiscovery.ServiceInfo)
	if err != nil {
		log.Error("Discovery: failed to encode service info: %w")
		return nil
	}
	return res
}

func (d *Discovery) parseMeta(node *memberlist.Node) (*discovery.ServiceInstanceInfo, error) {
	var meta discovery.ServiceInstanceInfo

	if err := json.Unmarshal(node.Meta, &meta); err != nil {
		return nil, fmt.Errorf("Discovery: failed to decode service info: %w", err)
	}
	return &meta, nil
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed
func (d *Discovery) NotifyMsg([]byte) {

}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit. Care should be taken that this method does not block,
// since doing so would block the entire UDP packet receive loop.
func (d *Discovery) GetBroadcasts(overhead, limit int) [][]byte {
	return nil
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (d *Discovery) LocalState(join bool) []byte {
	return nil
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (d *Discovery) MergeRemoteState(buf []byte, join bool) {

}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (d *Discovery) NotifyJoin(node *memberlist.Node) {
	d.addService(node, true)
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (d *Discovery) NotifyLeave(node *memberlist.Node) {
	serviceInfo, err := d.parseMeta(node)
	if err != nil {
		log.Error(err.Error())
		return
	}

	d.servicesMtx.RLock()
	defer d.servicesMtx.RUnlock()
	instances := d.services[serviceInfo.ServiceName].Instances
	for i := range instances {
		if instances[i].ID == serviceInfo.ID {
			instances[i].Healthy = false
			break
		}
	}

	gracefullyStopped := node.State == memberlist.StateLeft
	if gracefullyStopped {
		log.InfoWithFields(log.Fields{"service": serviceInfo}, "Discovery: service is gracefully stopped")
	} else {
		log.ErrorWithFields(log.Fields{"service": serviceInfo}, "Discovery: service is down")
	}

	// an attempt to include the service in the cluster again,
	// in case it has empty ClusterMembers
	if addr := serviceInfo.Address.String(); slices.Contains(d.cfg.ClusterMembers, addr) {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), d.cfg.DeregisterServiceAfter)
			defer cancel()
			d.join(ctx, !gracefullyStopped, addr)
		}()
	}

	d.TriggerEvent(discovery.EventTypeDown, *serviceInfo)
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (d *Discovery) NotifyUpdate(node *memberlist.Node) {
	d.addService(node, true)
}

var _ discovery.NotifyingDiscovery = &Discovery{}
