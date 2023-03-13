package flow

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"go.uber.org/zap"
)

func (f *flow) init() {

	var logger *zap.SugaredLogger

	defer shutdown()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	conf, err := util.ReadConfig("/etc/direktiv/flow-config.yaml")
	if err != nil {
		os.Exit(1)
	}

	err = Run(ctx, logger, conf)
	if err != nil {
		os.Exit(1)
	}
	/* load test data */
}

func (f *flow) TestServerInit(t *testing.T) {
	args := grpc.CreateNamespaceRequest{Name: "test-453"}
	res, _ := f.engine.flow.CreateNamespace(context.Background(), &args)
	if res.GetNamespace().Name != "test-453" {
		t.Errorf("excepted %s; got %s", "test-453", res.Namespace.Name)
	}

}

func shutdown() {
	// just in case, stop DNS server
	pv, err := os.ReadFile("/proc/version")
	if err == nil {
		// this is a direktiv machine, so we press poweroff
		if strings.Contains(string(pv), "#direktiv") {

			log.Printf("direktiv machine, powering off")

			if err := exec.Command("/sbin/poweroff").Run(); err != nil {
				fmt.Println("error shutting down:", err)
			}

		}
	}
}
