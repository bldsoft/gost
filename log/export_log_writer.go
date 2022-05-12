package log

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"
)

type LogExporterConfig struct {
	Instanse string `mapstructure:"SERVICE_NAME"`
}

type LogRecord struct {
	Instanse  string
	Timestamp int64
	Level     zerolog.Level
	ReqID     string
	Msg       []byte
}

type LogExporter interface {
	WriteLogRecord(r LogRecord) error
}

type ExportLogWriter struct {
	cfg       LogExporterConfig
	exporters []LogExporter
}

func NewExportLogWriter(cfg LogExporterConfig) *ExportLogWriter {
	return &ExportLogWriter{cfg: cfg}
}

func (w *ExportLogWriter) Append(exporter LogExporter) {
	w.exporters = append(w.exporters, exporter)
}

func (w *ExportLogWriter) Write(p []byte) (n int, err error) {
	if len(w.exporters) == 0 {
		return len(p), nil
	}

	var event map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(p))
	d.UseNumber()
	err = d.Decode(&event)

	rec := LogRecord{
		Instanse: w.cfg.Instanse,
	}

	if l, ok := event[zerolog.LevelFieldName].(string); ok {
		lvl, _ := zerolog.ParseLevel(l)
		rec.Level = zerolog.Level(lvl)
		delete(event, zerolog.LevelFieldName)
	}

	if reqID, ok := event[ReqIdFieldName].(string); ok {
		rec.ReqID = reqID
		delete(event, ReqIdFieldName)
	}

	if ts, ok := event[zerolog.TimestampFieldName].(string); ok {
		tt, err := time.Parse(zerolog.TimeFieldFormat, ts)
		if err != nil {
			return 0, err
		}
		rec.Timestamp = tt.Unix()
		delete(event, zerolog.TimestampFieldName)
	}

	rec.Msg, err = json.Marshal(event)
	if err != nil {
		return 0, err
	}

	var multiErr error
	for _, exporter := range w.exporters {
		if err := exporter.WriteLogRecord(rec); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}

	if multiErr != nil {
		return 0, multiErr
	}

	return len(p), nil
}
