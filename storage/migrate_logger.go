package storage

import "github.com/bldsoft/gost/log"

type MigrateLogger struct {
}

func (l MigrateLogger) Printf(format string, v ...any) {
	log.Debugf("Migrations: "+format, v...)
}

func (l MigrateLogger) Verbose() bool {
	return true
}
