package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/vorteil/direktiv/pkg/varstore"

	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/vorteil/direktiv/pkg/direktiv"
	"github.com/vorteil/direktiv/pkg/dlog"
	"github.com/vorteil/direktiv/pkg/dlog/db"
	"github.com/vorteil/direktiv/pkg/dlog/dummy"
	_ "github.com/vorteil/direktiv/pkg/util"
)

var (
	debug      bool
	configFile string
)

var rootCmd = &cobra.Command{
	Use:   "direktiv",
	Short: "direktiv is a serverless, container workflow engine.",
	PreRun: func(cmd *cobra.Command, args []string) {
		if debug || os.Getenv("DIREKTIV_DEBUG") == "true" {
			logrus.SetLevel(logrus.DebugLevel)
			formatter := runtime.Formatter{ChildFormatter: &logrus.TextFormatter{
				FullTimestamp: true,
			}}
			formatter.Line = true
			logrus.SetFormatter(&formatter)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {

		defer func() {
			// just in case, stop DNS server
			pv, err := ioutil.ReadFile("/proc/version")
			if err == nil {
				// this is a vorteil machine, so we press poweroff
				if strings.Contains(string(pv), "#vorteil") {
					logrus.Printf("vorteil machine, powering off")

					if err := exec.Command("/sbin/poweroff").Run(); err != nil {
						fmt.Println("error shutting down:", err)
					}

				}
			}

		}()

		c, err := direktiv.ReadConfig(configFile)
		if err != nil {
			logrus.Errorf("Failed to initialize server: %v", err)
			os.Exit(1)
		}

		server, err := direktiv.NewWorkflowServer(c)
		if err != nil {
			log.Fatalf("failed to create server: %v", err)
		}

		var logger dlog.Log

		switch c.InstanceLogging.Driver {
		case "database":
			logrus.Info("creating logger type database")
			l, err := db.NewLogger(c.Database.DB)
			if err != nil {
				logrus.Error(err)
				os.Exit(1)
			}
			defer l.CloseConnection()
			logger = l
		default:
			logrus.Info("creating logger type default")
			logger, _ = dummy.NewLogger()
		}

		server.SetInstanceLogger(logger)

		var vstore varstore.VarStorage

		switch c.VariablesStorage.Driver {
		case "":
			fallthrough
		case "database":
			vstore, err = varstore.NewPostgresVarStorage(c.Database.DB)
			if err != nil {
				logrus.Error(err)
				os.Exit(1)
			}
			defer vstore.Close()
		default:
			logrus.Error(errors.New("unsupported variables storage driver"))
			os.Exit(1)
		}

		server.SetVariableStorage(vstore)

		go func() {
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
			<-sig
			server.Stop()
			<-sig
			server.Kill()
		}()

		go func() {
			err := server.Run()
			if err != nil {
				log.Fatalf("unable to start server: %v", err)
			}
		}()

		<-server.Lifeline()

		log.Printf("server stopped\n")

		return

	},
}

func main() {

	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "enabled debug output")
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "configuration file to use")

	err := rootCmd.Execute()
	if err != nil {
		logrus.Errorf("%v", err)
		os.Exit(1)
	}

}
