package consul

import (
	"context"
	"sort"
	"strconv"
	"time"

	"github.com/bldsoft/gost/discovery"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/gost/utils/errgroup"
	"github.com/bldsoft/gost/version"
	"github.com/hashicorp/consul/api"
)

const (
	MetadataKeyVersion = "version"
	MetadataKeyBranch  = "branch"
	MetadataKeyCommmit = "commit"
	MetadataKeyNode    = "node"
	MetadataKeyProto   = "proto"
)

type Discovery struct {
	base         discovery.BaseDiscovery
	cfg          Config
	consulClient *api.Client
	server.AsyncRunner
}

func (d *Discovery) ApiClient() *api.Client {
	return d.consulClient
}

func (d *Discovery) SetMetadata(key string, value string) {
	d.base.SetMetadata(key, value)
}

func NewDiscovery(serviceCfg server.Config, consulCfg Config) *Discovery {
	d := &Discovery{base: discovery.NewBaseDiscovery(serviceCfg), cfg: consulCfg}
	if err := d.initClient(); err != nil {
		panic(err)
	}

	d.initMetadata()

	d.AsyncRunner = server.NewContextAsyncRunner(func(ctx context.Context) error {
		if len(serviceCfg.ServiceAddress) == 0 { // do not register in consul
			return nil
		}
		if err := d.Register(); err != nil {
			return err
		}
		d.heartBeat(ctx, d.cfg.HealthCheckTTL/3)
		return d.Deregister()
	})
	return d
}

func (d *Discovery) initMetadata() {
	d.SetMetadata(MetadataKeyVersion, version.Version)
	d.SetMetadata(MetadataKeyBranch, version.GitBranch)
	d.SetMetadata(MetadataKeyCommmit, version.GitCommit)
	d.SetMetadata(MetadataKeyNode, discovery.Hostname())
	d.SetMetadata(MetadataKeyProto, d.base.ServiceInfo.Proto)
}

func (d *Discovery) initClient() (err error) {
	d.consulClient, err = api.NewClient(&api.Config{
		Address: d.cfg.ConsulAddr.HostPort(),
		Scheme:  d.cfg.ConsulAddr.Scheme(),
		Token:   d.cfg.Token.String(),
	})
	return err
}

func (d *Discovery) Register() error {
	check := &api.AgentServiceCheck{
		TTL:     d.cfg.HealthCheckTTL.String(),
		CheckID: d.base.ServiceInfo.ID,
		Status:  api.HealthPassing,
	}
	if d.cfg.DeregisterTTL > 0 {
		check.DeregisterCriticalServiceAfter = d.cfg.DeregisterTTL.String()
	}

	reg := &api.AgentServiceRegistration{
		ID:      d.base.ServiceInfo.ID,
		Name:    d.base.ServiceInfo.ServiceName,
		Address: d.base.ServiceInfo.Host,
		Port:    d.base.ServiceInfo.PortInt(),
		Check:   check,
		Meta:    d.base.ServiceInfo.Meta,
	}

	return d.consulClient.Agent().ServiceRegister(reg)
}

func (d *Discovery) Deregister() error {
	return d.consulClient.Agent().ServiceDeregister(d.base.ServiceInfo.ID)
}

func (d *Discovery) heartBeat(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := d.consulClient.Agent().UpdateTTL(d.base.ServiceInfo.ID, "online", api.HealthPassing); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Discovery: consul health check failed")
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

func (d *Discovery) Services(ctx context.Context) ([]*discovery.ServiceInfo, error) {
	services, _, err := d.ApiClient().Catalog().Services(&api.QueryOptions{})
	if err != nil {
		return nil, err
	}

	var eg errgroup.Group
	serviceInfoC := make(chan discovery.ServiceInfo, len(services))
	for service := range services {
		service := service
		eg.Go(func() (err error) {
			nodes, _, err := d.ApiClient().Health().Service(service, "", false, &api.QueryOptions{})
			if err != nil {
				return err
			}

			serviceInfo := discovery.ServiceInfo{
				Name: service,
			}
			for _, node := range nodes {
				healthy := true
				for _, check := range node.Checks {
					if check.Status != api.HealthPassing {
						healthy = false
					}
				}

				serviceInfo.Instances = append(serviceInfo.Instances, discovery.ServiceInstanceInfo{
					Host:    node.Service.Address,
					Port:    strconv.Itoa(node.Service.Port),
					Proto:   node.Service.Meta[MetadataKeyProto],
					Node:    node.Service.Meta[MetadataKeyNode],
					Version: node.Service.Meta[MetadataKeyVersion],
					Branch:  node.Service.Meta[MetadataKeyBranch],
					Commit:  node.Service.Meta[MetadataKeyCommmit],
					Healthy: healthy,
					Meta:    node.Service.Meta,
				})
			}
			serviceInfoC <- serviceInfo
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	close(serviceInfoC)

	res := make([]*discovery.ServiceInfo, 0, len(services))
	for info := range serviceInfoC {
		info := info
		res = append(res, &info)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return res, nil
}

func (d *Discovery) ServiceByName(ctx context.Context, name string) (*discovery.ServiceInfo, error) {
	_, checkInfos, err := d.ApiClient().Agent().AgentHealthServiceByName(name)
	if err != nil {
		return nil, err
	}
	if len(checkInfos) == 0 {
		return nil, discovery.NotFound
	}

	res := &discovery.ServiceInfo{Name: name}
	for _, info := range checkInfos {
		res.Instances = append(res.Instances, discovery.ServiceInstanceInfo{
			Host:    info.Service.Address,
			Port:    strconv.Itoa(info.Service.Port),
			Node:    info.Service.Meta[MetadataKeyNode],
			Version: info.Service.Meta[MetadataKeyVersion],
			Branch:  info.Service.Meta[MetadataKeyBranch],
			Commit:  info.Service.Meta[MetadataKeyCommmit],
			Healthy: info.Checks.AggregatedStatus() == api.HealthPassing,
			Meta:    info.Service.Meta,
		})
	}
	return res, nil
}
