package main

import (
	gost "github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/examples/notes/pkg/config"
	"github.com/bldsoft/gost/examples/notes/pkg/microservice"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/gost/version"
)

func main() {
	var cfg config.Config
	gost.ReadConfig(&cfg, "")
	log.InitLogger(&cfg.Log, cfg.Server.LogExporterConfig())

	log.Infof("%s v%s", config.ServiceName, version.Version)
	log.Logf("ENVIRONMENT:\n***\n%s***", gost.FormatEnv(&cfg))

	server.NewServer(cfg.Server, microservice.New(cfg)).
		UseDefaultMiddlewares().
		WithHealthProbes(true).
		Start()
}
