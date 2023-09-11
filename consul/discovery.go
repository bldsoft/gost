package consul

import (
	"context"
	"os"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/gost/version"
	"github.com/hashicorp/consul/api"
)

const (
	MetadataKeyVersion = "version"
	MetadataKeyBranch  = "branch"
	MetadataKeyCommmit = "commit"
	MetadataKeyNode    = "node"
)

type Discovery struct {
	cfg          Config
	consulClient *api.Client
	server.AsyncRunner

	serviceMeta map[string]string
}

func (d *Discovery) ApiClient() *api.Client {
	return d.consulClient
}

func (d *Discovery) SetMetadata(name string, value string) {
	d.serviceMeta[name] = value
}

func NewDiscovery(cfg Config) *Discovery {
	d := &Discovery{cfg: cfg, serviceMeta: make(map[string]string)}
	if err := d.initClient(); err != nil {
		panic(err)
	}

	d.initMetadata()

	d.AsyncRunner = server.NewContextAsyncRunner(func(ctx context.Context) error {
		if len(cfg.ServiceAddr) == 0 { // do not register in consul
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

	hostname, err := os.Hostname()
	if err != nil {
		log.Error("Failed to get hostname")
	}
	d.SetMetadata(MetadataKeyNode, hostname)
}

func (d *Discovery) initClient() (err error) {
	d.consulClient, err = api.NewClient(&api.Config{
		Address: d.cfg.ConsulAddr,
		Scheme:  d.cfg.ConsulScheme,
		Token:   d.cfg.Token.String(),
	})
	return err
}

func (d *Discovery) Register() error {
	check := &api.AgentServiceCheck{
		TTL:     d.cfg.HealthCheckTTL.String(),
		CheckID: d.cfg.checkID(),
		Status:  api.HealthPassing,
	}
	if d.cfg.DeregisterTTL > 0 {
		check.DeregisterCriticalServiceAfter = d.cfg.DeregisterTTL.String()
	}

	reg := &api.AgentServiceRegistration{
		ID:      d.cfg.ServiceID,
		Name:    d.cfg.Cluster,
		Address: d.cfg.ServiceAddr,
		Port:    d.cfg.ServicePort,
		Check:   check,
		Meta:    d.serviceMeta,
	}

	return d.consulClient.Agent().ServiceRegister(reg)
}

func (d *Discovery) Deregister() error {
	return d.consulClient.Agent().ServiceDeregister(d.cfg.ServiceID)
}

func (d *Discovery) heartBeat(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := d.consulClient.Agent().UpdateTTL(d.cfg.checkID(), "online", api.HealthPassing); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Consul health check failed")
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}
