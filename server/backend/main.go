package main

import (
	"os"

	"github.com/direktiv/direktiv-ui/server/backend/server"
	"github.com/rs/zerolog/log"
)

func main() {

	// logJson := flag.Bool("json", true, "enable json logging")
	// configFile := flag.String("config", "config.yaml", "path to config file")
	// flag.Parse()

	var configDir string
	if len(os.Args) > 1 {
		configDir = os.Args[1]
	}

	_, err := server.ReadConfigAndPrepare(configDir)
	if err != nil {
		log.Fatal().Msgf("can not read config file: %s", err.Error())
	}

	// fmt.Printf("jjj %v", config)

	// rm, err := server.NewRouteManagerAPI(config)
	// if err != nil {
	// 	log.Fatal().Msgf("can not init server: %s", err.Error())
	// }

	// server := server.NewServer(config, rm)

	// license, err := ee.LicenseCheck()
	// if err != nil {
	// 	log.Error().Err(err).Msg("could not validate license")
	// }

	// log.Info().Msgf("licensed for %s (%d workflows)", license.Company, license.Workflows)

	// eeConfig := &ee.Config{
	// 	Provider: config.OIDC.Provider,
	// 	ClientID: config.OIDC.ClientID,
	// 	// Secret:       config.OIDC.Secret,
	// 	SkipVerify: config.OIDC.SkipVerify,
	// 	// CookieSecret: config.OIDC.CookieSecret,
	// 	Redirect:   config.OIDC.Redirect,
	// 	AdminGroup: config.OIDC.AdminGroup,
	// 	GroupScope: config.OIDC.GroupScope,
	// 	Host:       config.Server.Host,
	// 	// APIKey:       config.Server.APIKey,
	// }

	// rm, err := ee.NewRouteManagerEE(eeConfig)
	// if err != nil {
	// 	log.Fatal().Msgf("can not init server: %s", err.Error())
	// }

	// server := server.NewServer(config, rm)

	// server.Start()

}
