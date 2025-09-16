package sched

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

func calculateCronExpr(cronExpr string, start time.Time) (time.Time, error) {
	opts := cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow

	// if the cron expression has a seconds field, add it to the options
	if len(strings.Fields(cronExpr)) == 6 {
		opts |= cron.Second
	}
	schedule, err := cron.NewParser(opts).Parse(cronExpr)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse cron string: %s, err: %w", cronExpr, err)
	}

	return schedule.Next(start), nil
}
