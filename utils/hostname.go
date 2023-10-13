package utils

import (
	"os"
)

func HostnameOrErr(allowPanic ...bool) (string, error) {
	return os.Hostname()
}

func Hostname() string {
	hostname, err := HostnameOrErr()
	if err != nil {
		panic(err)
	}
	return hostname
}
