package direktiv

import (
	"io/ioutil"
	"os"

	"github.com/vorteil/direktiv/pkg/util"
	"gopkg.in/yaml.v2"
)

const (
	flowBind    = "0.0.0.0:7777"
	ingressBind = "0.0.0.0:6666"
)

// Config is the configuration for workflow and runner server
type Config struct {
	FunctionsProtocol string `yaml:"functions-protocol"`

	Database struct {
		DB string
	}

	InstanceLogging struct {
		Driver string
	}

	VariablesStorage struct {
		Driver string
	}
}

// ReadConfig reads the configuration file and overwrites with environment variables if set
func ReadConfig(file string) (*Config, error) {

	c := new(Config)

	appLog.Debugf("reading config %s", file)

	/* #nosec */
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}

	// settig additional config values
	c.Database.DB = os.Getenv(util.DBConn)

	// at the moment there is just one implementation
	c.InstanceLogging.Driver = "database"
	c.VariablesStorage.Driver = "database"

	return c, nil

}
