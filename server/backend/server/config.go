package server

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Log struct {
	API  string `yaml:"api"`
	JSON bool   `yaml:"json"`
}

type APIServer struct {
	Listen  string `yaml:"listen"`
	TLSCert string `yaml:"tls-cert"`
	TLSKey  string `yaml:"tls-key"`
	Assets  string `yaml:"assets"`
	Backend string `yaml:"backend"`
}

type OIDCConfig struct {
	Provider     string `yaml:"provider"`
	ClientID     string `yaml:"client-id"`
	Redirect     string `yaml:"redirect"`
	SkipVerify   bool   `yaml:"skip-verify"`
	AdminGroup   string `yaml:"admin-group"`
	GroupScope   string `yaml:"group-scope"`
	CookieSecret string `yaml:"cookiesecret"`
}

type Config struct {
	Log    Log        `yaml:"log"`
	Server APIServer  `yaml:"server"`
	OIDC   OIDCConfig `yaml:"oidc"`
}

const (
	DebugLog = "debug"
)

func readConfigFile(path string) (*Config, error) {
	var conf Config

	f, err := os.ReadFile(path)
	if err != nil {

		return &conf, err
	}

	err = yaml.Unmarshal(f, &conf)

	return &conf, err
}

func ReadConfigAndPrepare(configDir string) (*Config, error) {

	viper.SetConfigName("direktiv")
	viper.AddConfigPath(configDir)
	viper.SetEnvPrefix("DIREKTIV")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("log.json", false)
	viper.SetDefault("log.api", "debug")

	viper.SetDefault("server.listen", "0.0.0.0:2304")
	viper.SetDefault("server.assets", "/direktiv/html")
	viper.SetDefault("server.backend", "0.0.0.0:1604")

	if configDir != "" {
		log.Info().Msgf("starting direktiv with config file direktiv.yaml in %s", configDir)
		err := viper.ReadInConfig()
		if err != nil {
			return nil, err
		}
	} else {
		log.Info().Msgf("starting direktiv without config file")
	}

	var c Config

	err := viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	fmt.Printf(">>> %+v --> %v\n\n\n\n", c, viper.GetBool("log.json"))

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(os.Stderr)

	if !viper.GetBool("log.json") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true,
			TimeFormat: time.RFC3339Nano})
	}

	// conf, err := readConfigFile(configFile)
	// if err != nil {
	// 	log.Error().
	// 		Err(err).
	// 		Msgf("can not read config file %s", configFile)
	// 	return nil, err
	// }

	// set general logging level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if viper.GetString("log.api") == DebugLog {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	return nil, nil

}
