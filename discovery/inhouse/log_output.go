package inhouse

import (
	"fmt"

	"github.com/bldsoft/gost/log"
)

type logOutput struct{}

func (logOutput) Write(p []byte) (n int, err error) {
	var lvl, msg string
	if _, err := fmt.Sscanf(string(p), "[%s] %s", &lvl, &msg); err != nil {
		return 0, err
	}
	msg = "Discovery: " + msg
	n = len(p)
	switch lvl {
	case "INFO":
		log.Logger.Info(msg)
	case "WARN":
		log.Logger.Warn(msg)
	case "ERROR":
		log.Logger.Error(msg)
	case "DEBUG":
		fallthrough
	default:
		log.Logger.Debug(msg)
	}
	return
}
