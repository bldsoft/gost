package microservice

import (
	"sync"
	"time"

	"github.com/bldsoft/gost/auth"
	gost "github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/discovery"
	"github.com/bldsoft/gost/discovery/common"
	"github.com/bldsoft/gost/discovery/inhouse"
	"github.com/bldsoft/gost/examples/notes/pkg/config"
	"github.com/bldsoft/gost/examples/notes/pkg/controller"
	"github.com/bldsoft/gost/examples/notes/pkg/controller/rest"
	"github.com/bldsoft/gost/examples/notes/pkg/repository"
	"github.com/bldsoft/gost/examples/notes/pkg/service"
	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/gost/storage"
)

const (
	BasePath     = "/" + config.ServiceName
	APIKeyHeader = "X-API-Key"
)

var (
	_ service.NoteRepository           = (*repository.NoteRepository)(nil)
	_ controller.NoteService           = (*service.NoteService)(nil)
	_ server.InitializableMicroservice = (*Microservice)(nil)
)

type Microservice struct {
	cfg       config.Config
	db        *mongo.Storage
	discovery discovery.Discovery
	purger    server.AsyncRunner
	notes     *rest.NoteController
}

func New(cfg config.Config) *Microservice {
	db := mongo.NewStorage(cfg.Mongo)
	addMigrations(db)

	notesRepository := repository.NewNoteRepository(db)
	notesService := service.NewNoteService(notesRepository)

	return &Microservice{
		cfg:       cfg,
		db:        db,
		discovery: common.NewDiscovery(cfg.Server, cfg.Discovery),
		purger:    service.NewPurger(notesRepository, cfg.PurgeInterval, cfg.PurgeAfter),
		notes:     rest.NewNoteController(notesService),
	}
}

func (m *Microservice) Initialize() {
	var wg sync.WaitGroup
	storage.DBConnectAsync(&wg, m.db.Connect, -1, time.Second)
	wg.Wait()
}

func (m *Microservice) BuildRoutes(r server.Router) {
	if d, ok := m.discovery.(*inhouse.Discovery); ok {
		d.Mount(r)
	}

	r.Route(BasePath, func(r server.Router) {
		r.Get("/ping", gost.GetPingHandler)
		r.Get("/version", gost.GetVersionHandler)
		r.With(auth.ApiKeyMiddleware(APIKeyHeader, m.cfg.AdminAPIKey.String())).
			Get("/env", gost.GetEnvHandler(&m.cfg, nil))

		r.Route("/api/v1/notes", m.notes.Mount)
	})
}

// The chain stops in reverse order: the purger and discovery stop before the database disconnects.
func (m *Microservice) GetAsyncRunners() []server.AsyncRunner {
	return []server.AsyncRunner{
		server.NewAsyncJobChain(
			server.NewAsyncJob(nil, m.db.Disconnect),
			server.NewAsyncJobGroup(m.purger, m.discovery),
		),
	}
}
