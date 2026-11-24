package discovery

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/bldsoft/gost/config"
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
	Discovery
	Subscribe(handler *EventHandler, handlers ...*EventHandler)
}

type ServiceInfo struct {
	Name      string                `json:"name"`
	Instances []ServiceInstanceInfo `json:"instances"`
}

type ServiceInstanceInfo struct {
	ServiceName string            `json:"serviceName"`
	ID          string            `json:"id"`
	Address     config.Address    `json:"address"`
	Node        string            `json:"node"`
	Version     string            `json:"version"`
	Commit      string            `json:"commit"`
	Branch      string            `json:"branch"`
	Healthy     bool              `json:"healthy"`
	Meta        map[string]string `json:"meta"`
}

// don't expose the meta field to the logger cause some services might have sensitive data inside
func (si ServiceInstanceInfo) MarshalZerologObject(e *zerolog.Event) {
	e.Str("serviceName", si.ServiceName).
		Str("id", si.ID).
		Str("address", si.Address.String()).
		Str("node", si.Node).
		Str("version", si.Version).
		Str("commit", si.Commit).
		Str("branch", si.Branch).
		Bool("healthy", si.Healthy)
}
