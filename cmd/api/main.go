package main

import (
	"flag"
	"log"
	"os"

	"github.com/pelletier/go-toml"
	"github.com/vorteil/direktiv/pkg/api"
)

var cfgPath string
var cfg *api.Config

func init() {

	flag.StringVar(&cfgPath, "c", "conf.toml", "points to api server configuration file")
	flag.Parse()

	if _, e := os.Stat(cfgPath); e != nil {
		log.Fatalf("failed to locate config file at '%s': %s", cfgPath, e.Error())
	}

	cfg = new(api.Config)
	r, err := os.Open(cfgPath)
	if err != nil {
		log.Fatalf("failed to open config file: %s", err.Error())
	}
	defer r.Close()

	dec := toml.NewDecoder(r)
	err = dec.Decode(cfg)
	if err != nil {
		log.Fatalf("failed to parse config file contents: %s", err.Error())
	}

}

func main() {

	s, err := api.NewServer(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = s.Start()
	if err != nil {
		log.Fatal(err.Error())
	}

}
