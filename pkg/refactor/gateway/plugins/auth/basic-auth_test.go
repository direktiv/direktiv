package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/consumer"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
)

func TestExecutePlugin(t *testing.T) {

	// prepare consumer
	cl := []*core.Consumer{
		{
			Username: "demo",
			Password: "hello",
			Tags:     []string{"tag1", "tag2"},
		},
	}
	consumer.SetConsumer(cl)

	p, _ := plugins.GetPluginFromRegistry(basicAuthPluginName)

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/dummy", nil)

	p.ExecutePlugin(context.Background(), w, r)
	// httptest.NewRecorder()

	// p.ExecutePlugin()
	// t.Log(p)
	// consumerList []*core.Consumer

}
