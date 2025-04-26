package log

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bldsoft/gost/config/feature"
	"github.com/bldsoft/gost/utils"
	"github.com/bldsoft/gost/version"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"
)

type Level = zerolog.Level

type LogRecord struct {
	Service        string        `json:"service,omitempty"`
	ServiceVersion string        `json:"serviceVersion,omitempty"`
	Instance       string        `json:"instance,omitempty"`
	Timestamp      int64         `json:"timestamp,omitempty"`
	Level          zerolog.Level `json:"level,string,omitempty"`
	ReqID          string        `json:"reqID,omitempty"`
	Msg            string        `json:"msg,omitempty"`
	Fields         []byte        `json:"fields,omitempty"` // json
}

// for level parsing
func (r *LogRecord) UnmarshalJSON(data []byte) error {
	type Record struct {
		Service        string        `json:"service,omitempty"`
		ServiceVersion string        `json:"serviceVersion,omitempty"`
		Instance       string        `json:"instance,omitempty"`
		Timestamp      int64         `json:"timestamp,omitempty"`
		Level          zerolog.Level `json:"level,omitempty"`
		ReqID          string        `json:"reqID,omitempty"`
		Msg            string        `json:"msg,omitempty"`
		Fields         []byte        `json:"fields,omitempty"` // json
	}
	return json.Unmarshal(data, (*Record)(r))
}

type ExportLogWriter struct {
	cfg      LogExporterConfig
	hostname string

	exporters        []LogExporter
	exportersToggles []*feature.Bool
}

func NewExportLogWriter(cfg LogExporterConfig) *ExportLogWriter {
	return &ExportLogWriter{cfg: cfg, hostname: utils.Hostname()}
}

func (w *ExportLogWriter) Append(exporter LogExporter, isOn *feature.Bool) {
	w.exportersToggles = append(w.exportersToggles, isOn)
	w.exporters = append(w.exporters, exporter)
}

func (w *ExportLogWriter) allOff() bool {
	for _, t := range w.exportersToggles {
		if t == nil || t.Get() {
			return false
		}
	}
	return true
}

func (w *ExportLogWriter) parseRecord(p []byte) (*LogRecord, error) {
	var event map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(p))
	d.UseNumber()
	err := d.Decode(&event)
	if err != nil {
		return nil, err
	}

	rec := LogRecord{
		Service:        w.cfg.Service,
		ServiceVersion: version.LongVersion(),
		Instance:       w.hostname,
	}

	if l, ok := event[zerolog.LevelFieldName].(string); ok {
		lvl, _ := zerolog.ParseLevel(l)
		rec.Level = zerolog.Level(lvl)
		delete(event, zerolog.LevelFieldName)
	}

	if msg, ok := event[zerolog.MessageFieldName].(string); ok {
		rec.Msg = msg
		delete(event, zerolog.MessageFieldName)
	}

	if reqID, ok := event[ReqIdFieldName].(string); ok {
		rec.ReqID = reqID
		delete(event, ReqIdFieldName)
	}

	if ts, ok := event[zerolog.TimestampFieldName].(json.Number); ok {
		tt, err := ts.Int64()
		if err != nil {
			return nil, err
		}
		rec.Timestamp = tt
		delete(event, zerolog.TimestampFieldName)
	}

	if len(event) == 0 {
		return &rec, nil
	}

	rec.Fields, err = json.Marshal(event)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (w *ExportLogWriter) export(rec *LogRecord) error {
	var multiErr error
	for i, exporter := range w.exporters {
		if !exporter.IsReady() {
			continue
		}
		if t := w.exportersToggles[i]; t != nil && !t.Get() {
			continue
		}
		if _, err := exporter.Export(rec); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr
}

func (w *ExportLogWriter) Write(p []byte) (n int, err error) {
	if len(w.exporters) == 0 {
		return len(p), nil
	}

	if w.allOff() {
		return len(p), nil
	}

	rec, err := w.parseRecord(p)
	if err != nil {
		return 0, fmt.Errorf("failed to parse log record: %w", err)
	}

	err = w.export(rec)
	if err != nil {
		return 0, fmt.Errorf("failed to export log record: %w", err)
	}

	return len(p), nil
}
