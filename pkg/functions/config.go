package functions

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/util"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
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
		Small  string `yaml:"small"`
		Medium string `yaml:"medium"`
		Large  string `yaml:"large"`
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

func configWatcher() {
	cs, err := getClientSet()
	if err != nil {
		fmt.Printf("could not create fetch client: %v", err)
		return
	}

	for {
		if done, err := watchConfigChanges(cs); err != nil {
			logger.Errorf("function watcher channel failed to restart: %s", err.Error())
			return
		} else if done {
			// connection has ended
			return
		}
		logger.Debugf("function watcher channel has closed, attempting to restart")
		time.Sleep(5 * time.Second)
	}
}

func watchConfigChanges(cs *kubernetes.Clientset) (bool, error) {
	logger.Info("start watching configuration")

	ns := os.Getenv(util.DirektivNamespace)

	watcher, err := cs.CoreV1().ConfigMaps(ns).Watch(context.TODO(),
		metav1.SingleObject(metav1.ObjectMeta{Name: "direktiv-config-functions", Namespace: ns}))
	if err != nil {
		return false, fmt.Errorf("could not start watcher: %w", err)
	}

	for {
		select {
		case event := <-watcher.ResultChan():
			switch event.Type {
			case watch.Modified:
				cm := event.DeepCopy().Object.(*v1.ConfigMap)
				c := cm.Data["functions-config.yaml"]
				updateConfig([]byte(c), &functionsConfig)
			}

		case <-time.After(watcherTimeout):
			return false, nil
		}
	}
}

func updateConfig(data []byte, c *config) {
	err := yaml.Unmarshal(data, c)
	if err != nil {
		logger.Fatalf("can not unmarshal config file: %v", err)
		return
	}

	var sc subConfig
	err = kyaml.Unmarshal(data, &sc)
	if err != nil {
		logger.Fatalf("can not unmarshal config file (k8s): %v", err)
		return
	}

	c.extraVolumes = sc.ExtraVolumes
	c.extraContainers = sc.ExtraContainers
	c.knativeAffinity = sc.KnativeAffinity
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

	updateConfig(buf, &functionsConfig)
}
