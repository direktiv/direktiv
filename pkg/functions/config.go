package functions

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

type configReader struct {
}

type config struct {
	RequestTimeout     int    `yaml:"request-timeout"`
	ServiceAccount     string `yaml:"service-account"`
	Namespace          string `yaml:"namespace"`
	SidecarDb          string `yaml:"sidecar-db"`
	Sidecar            string `yaml:"sidecar"`
	PodCleaner         bool   `yaml:"pod-cleaner"`
	InitPod            string `yaml:"init-pod"`
	InitPodCertificate string `yaml:"init-pod-certificate"`
	KeepRevisions      int    `yaml:"keep-revisions"`
	MaxJobs            int    `yaml:"max-jobs"`
	MaxScale           int    `yaml:"max-scale"`
	NetShape           string `yaml:"net-shape"`
	Database           string `yaml:"db"`
	RolloutDuration    int    `yaml:"rollout-duration"`
	Concurrency        int    `yaml:"concurrency"`
	Storage            int    `yaml:"storage"`
	Runtime            string `yaml:"runtime"`
	PodSecret          string `yaml:"pod-secret"`
	Memory             struct {
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
	GrpcConfig           string         `yaml:"grpc-config"`
	AdditionalContainers []v1.Container `yaml:"additionalContainers"`
}

func newConfigReader() *configReader {
	return &configReader{}
}

func readAndSet(path string, target interface{}) {

	logger.Debugf("reading config %s", path)
	file, err := os.Open(path)
	if err != nil {
		logger.Errorf("can not open config file: %v", err)
		return
	}

	fi, err := file.Stat()
	if err != nil {
		logger.Errorf("can not stat file: %v", err)
		return
	}

	buf := make([]byte, fi.Size())
	_, err = io.ReadFull(file, buf)

	if err != nil {
		logger.Errorf("can not read config file: %v", err)
		return
	}

	err = yaml.Unmarshal(buf, target)
	if err != nil {
		logger.Errorf("can not unmarshal config file: %v", err)
		return
	}

}

func (cr *configReader) readConfig(path string, target interface{}) {
	readAndSet(path, target)
}
