package utils

import "time"

func Retry(delay time.Duration, times int, f func() error) {
	for range times {
		err := f()
		if err == nil {
			return
		}
		time.Sleep(delay)
	}
}
