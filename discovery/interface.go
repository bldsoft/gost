package discovery

import (
	"context"
	"strconv"

	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/gost/utils"
)

var NotFound = utils.ErrObjectNotFound

type Discovery interface {
	server.AsyncRunner
	Services(ctx context.Context) ([]*ServiceInfo, error)
	ServiceByName(ctx context.Context, name string) (*ServiceInfo, error)
	SetMetadata(key, value string)
}

type NotifyingDiscovery interface {
	Subscribe(handler EventServiceHandler, handlers ...EventServiceHandler)
}

type ServiceInfo struct {
	Name      string                `json:"name"`
	Instances []ServiceInstanceInfo `json:"instances"`
}

type ServiceInstanceInfo struct {
	ServiceName string            `json:"serviceName"`
	ID          string            `json:"id"`
	Proto       string            `json:"proto"`
	Host        string            `json:"host"`
	Port        string            `json:"port"`
	Node        string            `json:"node"`
	Version     string            `json:"version"`
	Commit      string            `json:"commit"`
	Branch      string            `json:"branch"`
	Healthy     bool              `json:"healthy"`
	Meta        map[string]string `json:"-"`
}

func (i ServiceInstanceInfo) PortInt() int {
	port, _ := strconv.Atoi(i.Port)
	return port
}
