// nolint
package function

import (
	"fmt"
	"strconv"

	"github.com/mitchellh/hashstructure/v2"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

type FunctionDefination struct {
	Namespace string
	Name      string

	ServicePath  string
	WorkflowPath string

	Image string
	CMD   string
	Size  string
	Scale int
}

func (c *FunctionDefination) id() string {
	str := fmt.Sprintf("%s-%s-%s-%s", c.Namespace, c.Name, c.ServicePath, c.WorkflowPath)
	v, err := hashstructure.Hash(str, hashstructure.FormatV2, nil)
	if err != nil {
		panic("unexpected hashstructure.Hash error: " + err.Error())
	}

	return fmt.Sprintf("obj-%d-obj", v)
}

func (c *FunctionDefination) hash() string {
	str := fmt.Sprintf("%s-%s-%s-%d", c.Image, c.CMD, c.Size, c.Scale)
	v, err := hashstructure.Hash(str, hashstructure.FormatV2, nil)
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
	Config *FunctionDefination
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
