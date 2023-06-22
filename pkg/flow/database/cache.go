package database

import (
	"encoding/json"

	"github.com/google/uuid"
)

const (
	PubsubNotifyFunction = "cache"
)

type Notifier interface {
	PublishToCluster(string)
}

type notification struct {
	Operation string
	ID        uuid.UUID
	Recursive bool
}

func (n *notification) Marshal() string {
	data, err := json.Marshal(n)
	if err != nil {
		panic(err)
	}
	return string(data)
}
