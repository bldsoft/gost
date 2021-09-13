package utils

import "time"

func TimeTrack(f func()) (d time.Duration) {
	start := time.Now()
	defer func() { d = time.Since(start) }()
	f()
	return d
}
