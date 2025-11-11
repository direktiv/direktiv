package runtime_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/direktiv/direktiv/internal/engine/runtime"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestHttpRequest(t *testing.T) {
	// Create container request
	req := testcontainers.ContainerRequest{
		Image:        "mendhak/http-https-echo:latest",
		ExposedPorts: []string{"8080/tcp", "8443/tcp"},
		WaitingFor:   wait.ForListeningPort("8080/tcp"), // wait until port is ready
	}

	// Start the container
	container, err := testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer container.Terminate(t.Context())

	mappedPort, err := container.MappedPort(t.Context(), "8080")
	if err != nil {
		log.Fatal(err)
	}

	script := `
		function start() {
			log(now().format("2006-01-02"))
			var r = fetchSync("` + fmt.Sprintf("http://localhost:%s", mappedPort.Port()) + `", {
				 method: "GET",
				 headers: {
					"Content-Type": "application/json",
					"Header1": "whatever",
					"X-Custom-Header": "customValue",
				},
				body: JSON.stringify({
					key: "value",
				}),
				params: {
					"param1": "p1",
					"param2": "v2",
				},
				skipTls: true,
				username: "admin",
				password: "password"
			})

			log(r.ok)

			for (const key in r.headers) {
   				 const value = r.headers[key];
    				log("key", key);
					for (const item of value) {
	  					log("value", item);	
					}
					
			}
			log(r.url)
			log(r.text())
			log(r.json()["path"])
			return "Hello"
		}
	`

	err = runtime.ExecScript(context.Background(), &runtime.Script{
		InstID:   uuid.New(),
		Text:     script,
		Mappings: "",
		Input:    "{}",
		Fn:       "start",
	}, runtime.NoOnFinish, runtime.NoOnTransition, runtime.NoOnAction)
	require.NoError(t, err)
}
