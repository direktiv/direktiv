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

	"github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/varstore"

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
	Run: func(cmd *cobra.Command, args []string) {

		defer func() {
			// just in case, stop DNS server
			pv, err := ioutil.ReadFile("/proc/version")
			if err == nil {
				// this is a vorteil machine, so we press poweroff
				if strings.Contains(string(pv), "#vorteil") {
					log.Printf("vorteil machine, powering off")

					if err := exec.Command("/sbin/poweroff").Run(); err != nil {
						fmt.Println("error shutting down:", err)
					}

				}
			}

		}()

		l, err := dlog.ApplicationLogger("flow")
		if err != nil {
			log.Fatalf("can not get logger: %v", err)
		}
		l.Info("starting direktiv flow component")

		c, err := direktiv.ReadConfig(configFile)
		if err != nil {
			log.Fatalf("Failed to initialize server: %v", err)
		}

		server, err := direktiv.NewWorkflowServer(c)
		if err != nil {
			log.Fatalf("failed to create server: %v", err)
		}

		var logger dlog.Log

		switch c.InstanceLogging.Driver {
		case "database":
			l.Info("creating logger type database")
			dl, err := db.NewLogger(c.Database.DB)
			if err != nil {
				log.Fatalf(err.Error())
			}
			defer dl.CloseConnection()
			logger = dl
		default:
			l.Info("creating logger type default")
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
			l.Error(errors.New("unsupported variables storage driver"))
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

		l.Infof("server stopped\n")

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
