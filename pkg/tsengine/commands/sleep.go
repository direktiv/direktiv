package commands

import "time"

type SleepCommand struct{}

func (c *SleepCommand) GetName() string {
	return "sleep"
}

func (c *SleepCommand) GetCommandFunction() interface{} {
	return func(duration int) {
		time.Sleep(time.Duration(duration) * time.Second)
	}
}
