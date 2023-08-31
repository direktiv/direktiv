package main

import (
	"os"

	"github.com/direktiv/direktiv-ui/frontend/server"
	"github.com/rs/zerolog/log"
)

func main() {

	var configDir string
	if len(os.Args) > 1 {
		configDir = os.Args[1]
	}

	var config *server.Config

	err := server.ReadConfigAndPrepare(configDir, &config)
	if err != nil {
		log.Fatal().Msgf("can not read config file: %s", err.Error())
	}

	rm, err := server.NewRouteManagerAPI(config)
	if err != nil {
		log.Fatal().Err(err).Msg("configuring route manager failed")
	}

	server := server.NewServer(config, rm)
	server.Start()

}
