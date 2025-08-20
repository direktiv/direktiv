package utils

import "time"

func Retry(delay time.Duration, times int, f func() error) {
	for i := 0; i < times; i++ {
		err := f()
		if err == nil {
			return
		}
		time.Sleep(delay)
	}
}
