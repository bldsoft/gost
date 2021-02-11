package log

import (
	"io"
	"os"
	"path"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
	//"github.com/sirupsen/logrus"
	//"github.com/snowzach/rotatefilehook"
)

//Logger is global log instance
var Logger = ServiceLogger{logger: zerolog.New(os.Stderr)}

//InitLogger initializes log
func InitLogger(logFile, logLevel string) {
	initZerolog(logFile, logLevel)
}

func initZerolog(logFile, logLevel string) {
	zerolog.TimestampFieldName = "t"
	zerolog.LevelFieldName = "l"
	zerolog.MessageFieldName = "m"

	var writers []io.Writer
	writers = append(writers, zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: "15:04:05"})

	if logFile != "" {
		writers = append(writers, newFileRotation(logFile))
	}

	multiWriter := zerolog.MultiLevelWriter(writers...)
	Logger = ServiceLogger{logger: zerolog.New(multiWriter).With().Timestamp().Logger()}

	err := SetLogLevel(logLevel)
	if err != nil {
		Logger.FatalWithFields(Fields{"error": err}, "Failed to parse LogLevel")
	}
}

//SetLogLevel sets global log level
func SetLogLevel(sLevel string) error {
	level, err := zerolog.ParseLevel(sLevel)
	if err == nil {
		zerolog.SetGlobalLevel(level)
	}
	return err
}

func newFileRotation(filename string) io.Writer {
	dir := path.Dir(filename)
	if err := os.MkdirAll(dir, 0744); err != nil {
		Logger.FatalWithFields(
			Fields{"path": dir, "error": err},
			"can't create log directory")
		return nil
	}

	return &lumberjack.Logger{
		Filename:   filename,
		MaxBackups: 3,  // files
		MaxSize:    50, // megabytes
		MaxAge:     30, // days
	}
}

/*func initLogrus(logFile, logLevel string) {

	//log to console
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:            true,
		FullTimestamp:          true,
		TimestampFormat:        "15:04:05.000",
		DisableLevelTruncation: true,
	})
	//log.SetReportCaller(true)
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Fatalf("Failed to parse LogLevel: %v", err)
	}
	logrus.SetLevel(level)

	if logFile != "" {
		rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
			Filename:   logFile,
			MaxSize:    50, // megabytes
			MaxBackups: 3,
			MaxAge:     30, //days
			Level:      level,
			Formatter:  &logrus.JSONFormatter{
				//TimestampFormat: time.RFC822,
			},
		})

		if err != nil {
			logrus.Fatalf("Failed to initialize file rotate hook: %v", err)
		}
		logrus.AddHook(rotateFileHook)
	}
}*/

/*func formatFilePath(path string) string {
    arr := strings.Split(path, "/")
    return arr[len(arr)-1]
}*/

// Trace logs a message at level Trace on the default logger.
func Trace(msg string) {
	Logger.Trace(msg)
}

// Tracef logs a message at level Trace on the default logger.
func Tracef(format string, v ...interface{}) {
	Logger.Tracef(format, v...)
}

// TraceWithFields logs a message at level Trace on the default logger.
func TraceWithFields(fields Fields, msg string) {
	Logger.TraceWithFields(fields, msg)
}

// TracefWithFields logs a message at level Trace on the default logger.
func TracefWithFields(fields Fields, format string, v ...interface{}) {
	Logger.TracefWithFields(fields, format, v...)
}

// Debug logs a message at level Debug on the default logger.
func Debug(msg string) {
	Logger.Debug(msg)
}

// Debugf logs a message at level Debug on the default logger.
func Debugf(format string, v ...interface{}) {
	Logger.Debugf(format, v...)
}

// DebugWithFields logs a message at level Debug on the default logger.
func DebugWithFields(fields Fields, msg string) {
	Logger.DebugWithFields(fields, msg)
}

// DebugfWithFields logs a message at level Debug on the default logger.
func DebugfWithFields(fields Fields, format string, v ...interface{}) {
	Logger.DebugfWithFields(fields, format, v...)
}

// DebugOrError logs a message at level Error if err is not nil or Debug otherwise on the default logger.
func DebugOrError(err error, msg string) {
	Logger.DebugOrError(err, msg)
}

// DebugOrErrorf logs a message at level Error if err is not nil or Debug otherwise on the default logger.
func DebugOrErrorf(err error, format string, v ...interface{}) {
	Logger.DebugOrErrorf(err, format, v...)
}

// DebugOrErrorWithFields logs a message at level Error if err is not nil or Debug otherwise on the default logger.
func DebugOrErrorWithFields(err error, fields Fields, msg string) {
	Logger.DebugOrErrorWithFields(err, fields, msg)
}

// DebugOrErrorfWithFields logs a message at level Error if err is not nil or Debug otherwise on the default logger.
func DebugOrErrorfWithFields(err error, fields Fields, format string, v ...interface{}) {
	Logger.DebugOrErrorfWithFields(err, fields, format, v...)
}

// Info logs a message at level Info on the default logger.
func Info(msg string) {
	Logger.Info(msg)
}

// Infof logs a message at level Info on the default logger.
func Infof(format string, v ...interface{}) {
	Logger.Infof(format, v...)
}

// InfoWithFields logs a message at level Info on the default logger.
func InfoWithFields(fields Fields, msg string) {
	Logger.InfoWithFields(fields, msg)
}

// InfofWithFields logs a message at level Info on the default logger.
func InfofWithFields(fields Fields, format string, v ...interface{}) {
	Logger.InfofWithFields(fields, format, v...)
}

// InfoOrError logs a message at level Error if err is not nil or Info otherwise on the default logger.
func InfoOrError(err error, msg string) {
	Logger.InfoOrError(err, msg)
}

// InfoOrErrorf logs a message at level Error if err is not nil or Info otherwise on the default logger.
func InfoOrErrorf(err error, format string, v ...interface{}) {
	Logger.InfoOrErrorf(err, format, v...)
}

// InfoOrErrorWithFields logs a message at level Error if err is not nil or Info otherwise on the default logger.
func InfoOrErrorWithFields(err error, fields Fields, msg string) {
	Logger.InfoOrErrorWithFields(err, fields, msg)
}

// InfoOrErrorfWithFields logs a message at level Error if err is not nil or Info otherwise on the default logger.
func InfoOrErrorfWithFields(err error, fields Fields, format string, v ...interface{}) {
	Logger.InfoOrErrorfWithFields(err, fields, format, v...)
}

// Warn logs a message at level Warn on the default logger.
func Warn(msg string) {
	Logger.Warn(msg)
}

// Warnf logs a message at level Warn on the default logger.
func Warnf(format string, v ...interface{}) {
	Logger.Warnf(format, v...)
}

// WarnWithFields logs a message at level Warn on the default logger.
func WarnWithFields(fields Fields, msg string) {
	Logger.WarnWithFields(fields, msg)
}

// WarnfWithFields logs a message at level Warn on the default logger.
func WarnfWithFields(fields Fields, format string, v ...interface{}) {
	Logger.WarnfWithFields(fields, format, v...)
}

// Error logs a message at level Error on the default logger.
func Error(msg string) {
	Logger.Error(msg)
}

// Errorf logs a message at level Error on the default logger.
func Errorf(format string, v ...interface{}) {
	Logger.Errorf(format, v...)
}

// ErrorWithFields logs a message at level Error on the default logger.
func ErrorWithFields(fields Fields, msg string) {
	Logger.ErrorWithFields(fields, msg)
}

// ErrorfWithFields logs a message at level Error on the default logger.
func ErrorfWithFields(fields Fields, format string, v ...interface{}) {
	Logger.ErrorfWithFields(fields, format, v...)
}

// Fatal logs a message at level Fatal on the default logger.
func Fatal(msg string) {
	Logger.Fatal(msg)
}

// Fatalf logs a message at level Fatal on the default logger.
func Fatalf(format string, v ...interface{}) {
	Logger.Fatalf(format, v...)
}

// FatalWithFields logs a message at level Fatal on the default logger.
func FatalWithFields(fields Fields, msg string) {
	Logger.FatalWithFields(fields, msg)
}

// FatalfWithFields logs a message at level Fatal on the default logger.
func FatalfWithFields(fields Fields, format string, v ...interface{}) {
	Logger.FatalfWithFields(fields, format, v...)
}

// Panic logs a message at level Panic on the default logger.
func Panic(msg string) {
	Logger.Panic(msg)
}

// Panicf logs a message at level Panic on the default logger.
func Panicf(format string, v ...interface{}) {
	Logger.Panicf(format, v...)
}

// PanicWithFields logs a message at level Panic on the default logger.
func PanicWithFields(fields Fields, msg string) {
	Logger.PanicWithFields(fields, msg)
}

// PanicfWithFields logs a message at level Panic on the default logger.
func PanicfWithFields(fields Fields, format string, v ...interface{}) {
	Logger.PanicfWithFields(fields, format, v...)
}

// Log logs a message at any level on the default logger.
func Log(msg string) {
	Logger.Log(msg)
}

// Logf logs a message at any level on the default logger.
func Logf(format string, v ...interface{}) {
	Logger.Logf(format, v...)
}

// LogWithFields logs a message at any level on the default logger.
func LogWithFields(fields Fields, msg string) {
	Logger.LogWithFields(fields, msg)
}

// LogfWithFields logs a message at any level on the default logger.
func LogfWithFields(fields Fields, format string, v ...interface{}) {
	Logger.LogfWithFields(fields, format, v...)
}
