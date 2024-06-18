package commands

import "time"

func Sleep(duration int) {
	time.Sleep(time.Duration(duration) * time.Second)
}
