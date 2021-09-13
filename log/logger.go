package log

import (
	"github.com/bldsoft/gost/utils"
	"github.com/rs/zerolog"
)

type ServiceLogger struct {
	logger zerolog.Logger

	LogFuncDuration bool
}

//Fields struct
type Fields map[string]interface{}

//WithFields creates a new logger with given fields
func (l *ServiceLogger) WithFields(fields Fields) ServiceLogger {
	return ServiceLogger{logger: l.logger.With().Fields(fields).Logger(), LogFuncDuration: l.LogFuncDuration}
}

func (l *ServiceLogger) WithPanic(fields Fields) ServiceLogger {
	return ServiceLogger{logger: l.logger.Level(zerolog.PanicLevel).With().Fields(fields).Logger()}
}

//WithFields creates a new logger with given fields
func (l *ServiceLogger) WithLevel(lvl zerolog.Level) ServiceLogger {
	return ServiceLogger{logger: l.logger.Level(lvl)}
}

// Trace logs a message at level Trace on the default logger.
func (l *ServiceLogger) Trace(msg string) {
	l.logger.Trace().Msg(msg)
}

// Tracef logs a message at level Trace on the default logger.
func (l *ServiceLogger) Tracef(format string, v ...interface{}) {
	l.logger.Trace().Msgf(format, v...)
}

// TraceWithFields logs a message at level Trace on the default logger.
func (l *ServiceLogger) TraceWithFields(fields Fields, msg string) {
	l.logger.Trace().Fields(fields).Msg(msg)
}

// TracefWithFields logs a message at level Trace on the default logger.
func (l *ServiceLogger) TracefWithFields(fields Fields, format string, v ...interface{}) {
	l.logger.Trace().Fields(fields).Msgf(format, v...)
}

// Debug logs a message at level Debug on the default logger.
func (l *ServiceLogger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a message at level Debug on the default logger.
func (l *ServiceLogger) Debugf(format string, v ...interface{}) {
	l.logger.Debug().Msgf(format, v...)
}

// DebugWithFields logs a message at level Debug on the default logger.
func (l *ServiceLogger) DebugWithFields(fields Fields, msg string) {
	l.logger.Debug().Fields(fields).Msg(msg)
}

// DebugfWithFields logs a message at level Debug on the default logger.
func (l *ServiceLogger) DebugfWithFields(fields Fields, format string, v ...interface{}) {
	l.logger.Debug().Fields(fields).Msgf(format, v...)
}

// DebugOrError logs a message at level Error if err is not nil or Debug otherwise on the default logger.
func (l *ServiceLogger) DebugOrError(err error, msg string) {
	if err == nil {
		l.logger.Debug().Msg(msg)
	} else {
		l.logger.Err(err).Msg(msg)
	}
}

// DebugOrErrorf logs a message at level Error if err is not nil or Debug otherwise on the default logger.
func (l *ServiceLogger) DebugOrErrorf(err error, format string, v ...interface{}) {
	if err == nil {
		l.logger.Debug().Msgf(format, v...)
	} else {
		l.logger.Err(err).Msgf(format, v...)
	}
}

// DebugOrErrorWithFields logs a message at level Error if err is not nil or Debug otherwise on the default logger.
func (l *ServiceLogger) DebugOrErrorWithFields(err error, fields Fields, msg string) {
	if err == nil {
		l.logger.Debug().Fields(fields).Msg(msg)
	} else {
		l.logger.Err(err).Fields(fields).Msg(msg)
	}
}

// DebugOrErrorfWithFields logs a message at level Error if err is not nil or Debug otherwise on the default logger.
func (l *ServiceLogger) DebugOrErrorfWithFields(err error, fields Fields, format string, v ...interface{}) {
	if err == nil {
		l.logger.Debug().Fields(fields).Msgf(format, v...)
	} else {
		l.logger.Err(err).Fields(fields).Msgf(format, v...)
	}
}

// Info logs a message at level Info on the default logger.
func (l *ServiceLogger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a message at level Info on the default logger.
func (l *ServiceLogger) Infof(format string, v ...interface{}) {
	l.logger.Info().Msgf(format, v...)
}

// InfoWithFields logs a message at level Info on the default logger.
func (l *ServiceLogger) InfoWithFields(fields Fields, msg string) {
	l.logger.Info().Fields(fields).Msg(msg)
}

// InfofWithFields logs a message at level Info on the default logger.
func (l *ServiceLogger) InfofWithFields(fields Fields, format string, v ...interface{}) {
	l.logger.Info().Fields(fields).Msgf(format, v...)
}

// InfoOrError logs a message at level Error if err is not nil or Info otherwise on the default logger.
func (l *ServiceLogger) InfoOrError(err error, msg string) {
	l.logger.Err(err).Msg(msg)
}

// InfoOrErrorf logs a message at level Error if err is not nil or Info otherwise on the default logger.
func (l *ServiceLogger) InfoOrErrorf(err error, format string, v ...interface{}) {
	l.logger.Err(err).Msgf(format, v...)
}

// InfoOrErrorWithFields logs a message at level Error if err is not nil or Info otherwise on the default logger.
func (l *ServiceLogger) InfoOrErrorWithFields(err error, fields Fields, msg string) {
	l.logger.Err(err).Fields(fields).Msg(msg)
}

// InfoOrErrorfWithFields logs a message at level Error if err is not nil or Info otherwise on the default logger.
func (l *ServiceLogger) InfoOrErrorfWithFields(err error, fields Fields, format string, v ...interface{}) {
	l.logger.Err(err).Fields(fields).Msgf(format, v...)
}

// Warn logs a message at level Warn on the default logger.
func (l *ServiceLogger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a message at level Warn on the default logger.
func (l *ServiceLogger) Warnf(format string, v ...interface{}) {
	l.logger.Warn().Msgf(format, v...)
}

// WarnWithFields logs a message at level Warn on the default logger.
func (l *ServiceLogger) WarnWithFields(fields Fields, msg string) {
	l.logger.Warn().Fields(fields).Msg(msg)
}

// WarnfWithFields logs a message at level Warn on the default logger.
func (l *ServiceLogger) WarnfWithFields(fields Fields, format string, v ...interface{}) {
	l.logger.Warn().Fields(fields).Msgf(format, v...)
}

// Error logs a message at level Error on the default logger.
func (l *ServiceLogger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a message at level Error on the default logger.
func (l *ServiceLogger) Errorf(format string, v ...interface{}) {
	l.logger.Error().Msgf(format, v...)
}

// ErrorWithFields logs a message at level Error on the default logger.
func (l *ServiceLogger) ErrorWithFields(fields Fields, msg string) {
	l.logger.Error().Fields(fields).Msg(msg)
}

// ErrorfWithFields logs a message at level Error on the default logger.
func (l *ServiceLogger) ErrorfWithFields(fields Fields, format string, v ...interface{}) {
	l.logger.Error().Fields(fields).Msgf(format, v...)
}

// Fatal logs a message at level Fatal on the default logger.
func (l *ServiceLogger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a message at level Fatal on the default logger.
func (l *ServiceLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatal().Msgf(format, v...)
}

// FatalWithFields logs a message at level Fatal on the default logger.
func (l *ServiceLogger) FatalWithFields(fields Fields, msg string) {
	l.logger.Fatal().Fields(fields).Msg(msg)
}

// FatalfWithFields logs a message at level Fatal on the default logger.
func (l *ServiceLogger) FatalfWithFields(fields Fields, format string, v ...interface{}) {
	l.logger.Fatal().Fields(fields).Msgf(format, v...)
}

// Panic logs a message at level Panic on the default logger.
func (l *ServiceLogger) Panic(msg string) {
	l.logger.Panic().Msg(msg)
}

// Panicf logs a message at level Panic on the default logger.
func (l *ServiceLogger) Panicf(format string, v ...interface{}) {
	l.logger.Panic().Msgf(format, v...)
}

// PanicWithFields logs a message at level Panic on the default logger.
func (l *ServiceLogger) PanicWithFields(fields Fields, msg string) {
	l.logger.Panic().Fields(fields).Msg(msg)
}

// PanicfWithFields logs a message at level Panic on the default logger.
func (l *ServiceLogger) PanicfWithFields(fields Fields, format string, v ...interface{}) {
	l.logger.Panic().Fields(fields).Msgf(format, v...)
}

// Panic logs a message at level Panic on the default logger.
func (l *ServiceLogger) Log(msg string) {
	l.logger.Log().Msg(msg)
}

// Panicf logs a message at level Panic on the default logger.
func (l *ServiceLogger) Logf(format string, v ...interface{}) {
	l.logger.Log().Msgf(format, v...)
}

// PanicWithFields logs a message at level Panic on the default logger.
func (l *ServiceLogger) LogWithFields(fields Fields, msg string) {
	l.logger.Log().Fields(fields).Msg(msg)
}

// PanicfWithFields logs a message at level Panic on the default logger.
func (l *ServiceLogger) LogfWithFields(fields Fields, format string, v ...interface{}) {
	l.logger.Log().Fields(fields).Msgf(format, v...)
}

// WithFuncDuration runs f and returns logger with field with its execution time
func (l *ServiceLogger) WithFuncDuration(f func()) ServiceLogger {
	d := utils.TimeTrack(f)
	return l.WithFields(Fields{"time_ms": d})
}

// WithOptFuncDuration is same as WithFuncDuration if ServiceLogger.LogFuncFuration is true, and returns a copy of the current logger otherwise
func (l *ServiceLogger) WithOptFuncDuration(f func()) ServiceLogger {
	if l.LogFuncDuration {
		return l.WithFuncDuration(f)
	}
	f()
	return *l
}
