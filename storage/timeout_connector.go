package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/bldsoft/gost/log"
)

func DBConnectAsync(wg *sync.WaitGroup, connect func(), n int, sleepPeriod time.Duration) {
	wg.Add(1)
	go func() {
		time.Sleep(10 * time.Second)
		defer wg.Done()
		for i := 1; (i != n+1) || (n < 0); i++ {
			err := func() (err error) {
				defer func() {
					if _err := recover(); _err != nil {
						err = fmt.Errorf("%v", _err)
					}
				}()
				connect()
				return
			}()

			if err != nil {
				log.ErrorWithFields(log.Fields{"error": err}, "error connecting to db")
				time.Sleep(sleepPeriod)
				continue
			}

			return
		}
	}()
}
