package datastore

import (
	"context"
	"time"
)

type HeartBeat struct {
	Group string
	Key   string

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type HeartBeatsStore interface {
	Set(ctx context.Context, heartBeat *HeartBeat) error
	Since(ctx context.Context, group string, secondsAgo int) ([]*HeartBeat, error)
}
