package storage

import (
	"fmt"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils/errgroup"
)

func DBConnectAsync(eg *errgroup.Group, connect func(), n int, sleepPeriod time.Duration) {
	go func() {
		eg.Go(func() error {
			var err error
			for i := 1; (i != n+1) || (n < 0); i++ {
				err = func() (err error) {
					defer func() {
						if _err := recover(); _err != nil {
							err = fmt.Errorf("%v", err)
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

				return nil
			}
			return err
		})
	}()
}
