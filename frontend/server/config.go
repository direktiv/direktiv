package server

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Log struct {
	API  string `yaml:"api"`
	JSON bool   `yaml:"json"`
}

type APIServer struct {
	Listen      string `yaml:"listen"`
	TLSCert     string `yaml:"tls-cert"`
	TLSKey      string `yaml:"tls-key"`
	Assets      string `yaml:"assets"`
	Backend     string `yaml:"backend"`
	BackendSkip bool   `yaml:"skipverify"`
	APIKey      string `yaml:"apikey"`
}

type Config struct {
	Log    Log       `yaml:"log"`
	Server APIServer `yaml:"server"`
}

const (
	DebugLog = "debug"
)

func ReadConfigAndPrepare(configDir string, c interface{}) error {

	viper.SetConfigName("direktiv")
	viper.AddConfigPath(configDir)
	viper.SetEnvPrefix("DIREKTIV")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("log.json", false)
	viper.SetDefault("log.api", "debug")

	viper.SetDefault("server.listen", "0.0.0.0:2304")
	viper.SetDefault("server.assets", "/html")
	viper.SetDefault("server.backend", "localhost:1604")
	viper.SetDefault("server.backendskip", true)
	viper.SetDefault("server.apikey", "")
	viper.SetDefault("server.tlscert", "")
	viper.SetDefault("server.tlskey", "")

	if configDir != "" {
		log.Info().Msgf("starting direktiv with config file direktiv.yaml in %s", configDir)
		err := viper.ReadInConfig()
		if err != nil {
			return err
		}
	} else {
		log.Info().Msgf("starting direktiv without config file")
	}

	err := viper.Unmarshal(&c)
	if err != nil {
		return err
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(os.Stderr)

	if !viper.GetBool("log.json") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true,
			TimeFormat: time.RFC3339Nano})
	}

	// set general logging level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if viper.GetString("log.api") == DebugLog {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		b, _ := json.MarshalIndent(c, "", "   ")
		log.Debug().Msgf("%s", string(b))
	}

	return nil

}
