package consumer

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/stretchr/testify/assert"
)

func TestConsumerNoUsername(t *testing.T) {
	cl := []*core.ConsumerFile{
		{},
	}

	consumerList := NewConsumerList()
	consumerList.SetConsumers(cl)
	assert.Len(t, consumerList.GetConsumers(), 0)
}

func TestConsumerWithAttributes(t *testing.T) {
	consumerList := NewConsumerList()
	cl := []*core.ConsumerFile{
		{
			Username: "user",
			Password: "pwd",
			Tags:     []string{"tag1", "tag2"},
			Groups:   []string{"group1"},
			APIKey:   "123",
		},
	}
	consumerList.SetConsumers(cl)
	assert.Len(t, consumerList.GetConsumers(), 1)

	assert.NotNil(t, consumerList.FindByUser("user"))
	assert.NotNil(t, consumerList.FindByAPIKey("123"))

	assert.Nil(t, consumerList.FindByUser("doesnotexist"))
	assert.Nil(t, consumerList.FindByAPIKey("doesnotexist"))
}

func TestConsumerDuplicate(t *testing.T) {
	consumerList := NewConsumerList()
	cl := []*core.ConsumerFile{
		{
			Username: "user",
		},
		{
			Username: "user",
		},
	}
	consumerList.SetConsumers(cl)
	assert.Len(t, consumerList.GetConsumers(), 1)
}
