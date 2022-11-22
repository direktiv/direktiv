package functions

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	kyaml "sigs.k8s.io/yaml"
)

type config struct {
	Logging      string `yaml:"logging"`
	IngressClass string `yaml:"ingress-class"`
	FlowService  string `yaml:"flow-service"`

	ServiceAccount string `yaml:"service-account"`
	Namespace      string `yaml:"namespace"`
	Sidecar        string `yaml:"sidecar"`

	MaxScale int    `yaml:"max-scale"`
	NetShape string `yaml:"net-shape"`

	Runtime              string `yaml:"runtime"`
	OpenTelemetryBackend string `yaml:"opentelemetry-backend"`

	Memory struct {
		Small  int `yaml:"small"`
		Medium int `yaml:"medium"`
		Large  int `yaml:"large"`
	} `yaml:"memory"`
	CPU struct {
		Small  float64 `yaml:"small"`
		Medium float64 `yaml:"medium"`
		Large  float64 `yaml:"large"`
	} `yaml:"cpu"`
	Disk struct {
		Small  int `yaml:"small"`
		Medium int `yaml:"medium"`
		Large  int `yaml:"large"`
	} `yaml:"disk"`
	Proxy struct {
		No    string `yaml:"no"`
		HTTPS string `yaml:"https"`
		HTTP  string `yaml:"http"`
	} `yaml:"proxy"`

	knativeAffinity v1.NodeAffinity `yaml:"-"`
	extraContainers []v1.Container  `yaml:"-"`
	extraVolumes    []v1.Volume     `yaml:"-"`

	// Deprecated
	RequestTimeout  int    `yaml:"request-timeout"`
	Concurrency     int    `yaml:"concurrency"`
	Storage         int    `yaml:"storage"`
	PodCleaner      bool   `yaml:"pod-cleaner"`
	InitPod         string `yaml:"init-pod"`
	KeepRevisions   int    `yaml:"keep-revisions"`
	MaxJobs         int    `yaml:"max-jobs"`
	RolloutDuration int    `yaml:"rollout-duration"`
	Database        string `yaml:"db"`
}

type subConfig struct {
	ExtraContainers []v1.Container  `yaml:"extraContainers"`
	ExtraVolumes    []v1.Volume     `yaml:"extraVolumes"`
	KnativeAffinity v1.NodeAffinity `yaml:"knativeAffinity"`
}

func readConfig(path string, c *config) {

	logger.Debugf("reading config %s", path)
	file, err := os.Open(path)
	if err != nil {
		logger.Fatalf("can not open config file: %v", err)
		return
	}

	fi, err := file.Stat()
	if err != nil {
		logger.Fatalf("can not stat file: %v", err)
		return
	}

	buf := make([]byte, fi.Size())
	_, err = io.ReadFull(file, buf)

	if err != nil {
		logger.Fatalf("can not read config file: %v", err)
		return
	}

	err = yaml.Unmarshal(buf, c)
	if err != nil {
		logger.Fatalf("can not unmarshal config file: %v", err)
		return
	}

	var sc subConfig
	err = kyaml.Unmarshal(buf, &sc)
	if err != nil {
		logger.Fatalf("can not unmarshal config file (k8s): %v", err)
		return
	}

	c.extraVolumes = sc.ExtraVolumes
	c.extraContainers = sc.ExtraContainers
	c.knativeAffinity = sc.KnativeAffinity

}
