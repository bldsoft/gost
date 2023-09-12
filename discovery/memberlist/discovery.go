package memberlist

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bldsoft/gost/discovery"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils"
	"github.com/hashicorp/memberlist"
	"golang.org/x/exp/slices"
)

var NotFound = utils.ErrObjectNotFound

type serviceInfo struct {
	Service string `json:"service"`
	discovery.ServiceInstanceInfo
}

type Discovery struct {
	cfg  Config
	list *memberlist.Memberlist

	serviceInfo serviceInfo

	services    map[string]*discovery.ServiceInfo
	servicesMtx sync.RWMutex
}

func NewDiscovery(cfg Config) *Discovery {
	return &Discovery{
		cfg: cfg,
		serviceInfo: serviceInfo{
			Service:             cfg.ServiceName,
			ServiceInstanceInfo: cfg.ServiceInstanceInfo(),
		},
		services: make(map[string]*discovery.ServiceInfo),
	}
}

func (d *Discovery) memberlistConfig() *memberlist.Config {
	memberlistCfg := memberlist.DefaultLocalConfig()
	memberlistCfg.LogOutput = logOutput{}
	memberlistCfg.Name = d.cfg.ServiceID
	memberlistCfg.BindAddr = d.cfg.MemberlistHost
	memberlistCfg.BindPort = d.cfg.MemberlistPort
	memberlistCfg.Delegate = d
	memberlistCfg.Events = d
	return memberlistCfg
}

func (d *Discovery) Run() error {
	var err error
	d.list, err = memberlist.Create(d.memberlistConfig())
	if err != nil {
		return fmt.Errorf("failed to create memberlist: %w", err)
	}
	if len(d.cfg.ClusterMembers) == 0 {
		return nil
	}
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		// Join an existing cluster by specifying at least one known member.
		if _, err = d.list.Join(d.cfg.ClusterMembers); err != nil {
			log.Errorf("Failed to join memberlist cluster: %s", strings.TrimSpace(err.Error()))
		} else {
			break
		}
		<-t.C
	}
	return nil
}

func (d *Discovery) addService(node *memberlist.Node, withLock bool) {
	meta, err := d.parseMeta(node)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	if withLock {
		d.servicesMtx.Lock()
		defer d.servicesMtx.Unlock()
	}
	serviceInfo, ok := d.services[meta.Service]
	if !ok {
		serviceInfo = &discovery.ServiceInfo{Name: meta.Service}
		d.services[meta.Service] = serviceInfo
	}
	meta.ServiceInstanceInfo.Healthy = true
	i := slices.IndexFunc(serviceInfo.Instances, func(si discovery.ServiceInstanceInfo) bool {
		return si.ID == meta.ID
	})
	if i >= 0 {
		serviceInfo.Instances[i] = meta.ServiceInstanceInfo
	} else {
		serviceInfo.Instances = append(serviceInfo.Instances, meta.ServiceInstanceInfo)
	}
}

func (d *Discovery) Stop(ctx context.Context) error {
	return nil
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

// Delegates

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (d *Discovery) NodeMeta(limit int) []byte {
	var buf bytes.Buffer
	buf.Grow(limit)
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(d.serviceInfo); err != nil {
		log.Error("memberlist: failed to encode service info: %w")
		return nil
	}
	return buf.Bytes()
}

func (d *Discovery) parseMeta(node *memberlist.Node) (*serviceInfo, error) {
	var meta serviceInfo
	decoder := gob.NewDecoder(bytes.NewReader(node.Meta))
	err := decoder.Decode(&meta)
	if err != nil {
		return nil, fmt.Errorf("memberlist: failed to decode service info: %w", err)
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
	meta, err := d.parseMeta(node)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	d.servicesMtx.RLock()
	defer d.servicesMtx.RUnlock()
	instaces := d.services[meta.Service].Instances
	for i := range instaces {
		if instaces[i].ID == meta.ServiceInstanceInfo.ID {
			instaces[i].Healthy = false
			break
		}
	}
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (d *Discovery) NotifyUpdate(node *memberlist.Node) {

}
