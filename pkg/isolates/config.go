package isolates

import (
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type configReader struct {
}

type config struct {
	ServiceAccount  string `yaml:"service-account"`
	Namespace       string `yaml:"namespace"`
	SidecarDb       string `yaml:"sidecar-db"`
	Sidecar         string `yaml:"sidecar"`
	InitPod         string `yaml:"init-pod"`
	KeepRevisions   int    `yaml:"keep-revisions"`
	MaxScale        int    `yaml:"max-scale"`
	NetShape        string `yaml:"net-shape"`
	RolloutDuration int    `yaml:"rollout-duration"`
	Concurrency     int    `yaml:"concurrency"`
	Storage         int    `yaml:"storage"`
	Runtime         string `yaml:"runtime"`
	Memory          struct {
		Small  int `yaml:"small"`
		Medium int `yaml:"medium"`
		Large  int `yaml:"large"`
	} `yaml:"memory"`
	CPU struct {
		Small  float64 `yaml:"small"`
		Medium float64 `yaml:"medium"`
		Large  float64 `yaml:"large"`
	} `yaml:"cpu"`
	Proxy struct {
		No    string `yaml:"no"`
		HTTPS string `yaml:"https"`
		HTTP  string `yaml:"http"`
	} `yaml:"proxy"`
	AdditionalContainers []struct {
		Name    string `yaml:"name"`
		Image   string `yaml:"image"`
		Command string `yaml:"command"`
	} `yaml:"additionalContainers"`
}

func newConfigReader() *configReader {
	return &configReader{}
}

func readAndSet(path string, target interface{}) {
	file, err := os.Open(path)
	if err != nil {
		log.Errorf("can not open config file: %v", err)
		return
	}

	fi, err := file.Stat()
	if err != nil {
		log.Errorf("can not stat file: %v", err)
		return
	}

	buf := make([]byte, fi.Size())
	_, err = io.ReadFull(file, buf)

	if err != nil {
		log.Errorf("can not read config file: %v", err)
		return
	}

	err = yaml.Unmarshal(buf, target)
	if err != nil {
		log.Errorf("can not unmarshal config file: %v", err)
		return
	}

}

func (cr *configReader) readConfig(path string, target interface{}) {

	// initial read
	readAndSet(path, target)

	// for some reason watcher did not work on k3s
	// that might be good enough for now
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		readAndSet(path, target)
	}

}
