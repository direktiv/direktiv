package function2

import (
	"fmt"
	"strconv"

	"github.com/mitchellh/hashstructure/v2"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

type FunctionConfig struct {
	Namespace string

	Name string

	ServicePath  string
	WorkflowPath string

	Config struct {
		CMD   string
		Image string
	}
}

func (c *FunctionConfig) id() string {
	name := fmt.Sprintf("%s-%s-%s-%s", c.Namespace, c.Name, c.ServicePath, c.WorkflowPath)
	v, err := hashstructure.Hash(name, hashstructure.FormatV2, nil)
	if err != nil {
		panic("unexpected hashstructure.Hash error: " + err.Error())
	}

	return fmt.Sprintf("obj-%d-obj", v)
}

func (c *FunctionConfig) hash() string {
	v, err := hashstructure.Hash(c.Config, hashstructure.FormatV2, nil)
	if err != nil {
		panic("unexpected hashstructure.Hash error: " + err.Error())
	}

	return strconv.Itoa(int(v))
}

type FunctionStatus interface {
	data() any
	id() string
	hash() string
}

type Function struct {
	Config *FunctionConfig
	Status FunctionStatus
}

type K8sFunctionStatus struct {
	*servingv1.Service
}

func (r *K8sFunctionStatus) data() any {
	return r.Service
}

func (r *K8sFunctionStatus) id() string {
	return r.Name
}

func (r *K8sFunctionStatus) hash() string {
	return r.Annotations["direktiv.io/input_hash"]
}

var _ FunctionStatus = &K8sFunctionStatus{}
