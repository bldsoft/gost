package discovery

import (
	"context"

	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/gost/utils"
)

var NotFound = utils.ErrObjectNotFound

type Discovery interface {
	server.AsyncRunner
	Services(ctx context.Context) ([]*ServiceInfo, error)
	ServiceByName(ctx context.Context, name string) (*ServiceInfo, error)
}

type ServiceInfo struct {
	Name      string                `json:"name"`
	Instances []ServiceInstanceInfo `json:"instances"`
}

type ServiceInstanceInfo struct {
	ID      string            `json:"id"`
	Proto   string            `json:"proto"`
	Address string            `json:"address"`
	Port    int               `json:"port"`
	Node    string            `json:"node"`
	Version string            `json:"version"`
	Commit  string            `json:"commit"`
	Branch  string            `json:"branch"`
	Healthy bool              `json:"healthy"`
	Meta    map[string]string `json:"-"`
}
