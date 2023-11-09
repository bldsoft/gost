package inhouse

import (
	"regexp"

	"github.com/bldsoft/gost/log"
)

var memberlistLogRE = regexp.MustCompile(`^.*\[(.+)\] (.*)`)

type logOutput struct{}

func (logOutput) Write(p []byte) (n int, err error) {
	const submatchesCount = 3
	submatches := memberlistLogRE.FindSubmatch(p)
	if len(submatches) != submatchesCount {
		log.Errorf("Discovery: failed to parse message: %s", p)
		return 0, nil
	}
	lvl, msg := string(submatches[1]), string(submatches[2])

	n = len(p)
	msg = "Discovery: " + msg
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
