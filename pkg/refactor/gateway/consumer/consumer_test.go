package consumer

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/stretchr/testify/assert"
)

func TestConsumerNoUsername(t *testing.T) {
	cl := []*core.Consumer{
		{},
	}
	SetConsumer(cl)
	assert.Len(t, GetConsumers(), 0)
}

func TestConsumerWithAttributes(t *testing.T) {
	cl := []*core.Consumer{
		{
			Username: "user",
			Password: "pwd",
			Tags:     []string{"tag1", "tag2"},
			Groups:   []string{"group1"},
			APIKey:   "123",
		},
	}
	SetConsumer(cl)
	assert.Len(t, GetConsumers(), 1)

	assert.NotNil(t, FindByUser("user"))
	assert.NotNil(t, FindByAPIKey("123"))

	assert.Nil(t, FindByUser("doesnotexist"))
	assert.Nil(t, FindByAPIKey("doesnotexist"))
}

func TestConsumerDuplicate(t *testing.T) {

	cl := []*core.Consumer{
		{
			Username: "user",
		},
		{
			Username: "user",
		},
	}
	SetConsumer(cl)
	assert.Len(t, GetConsumers(), 1)

}
