package compiler

type Compiler struct{}

func New(path, typeScript string) (*Compiler, error) {
	return &Compiler{}, nil
}

func (c *Compiler) CompileFlow() (*FlowInformation, error) {
	return &FlowInformation{}, nil
}

type FlowInformation struct {
	ID         string
	Definition *Definition

	Functions map[string]Function
}

type Function struct {
	Image   string            `json:"image,omitempty"`
	Size    string            `json:"size"`
	Envs    map[string]string `json:"envs,omitempty"`
	Cmd     string            `json:"cmd,omitempty"`
	Init    []string          `json:"init,omitempty"`
	Flow    string            `json:"flow,omitempty"`
	Service string            `json:"service,omitempty"`
}

type Definition struct {
	Type  string
	Store string
	JSON  bool
	State string
	Cron  string

	Timeout string

	Event  FlowEvent
	Events []FlowEvent

	Scale []Scale
}

type FlowEvent struct {
	Type    string
	Context map[string]interface{}
}

type Scale struct {
	Min    int
	Max    int
	Cron   string
	Metric string
	Value  int
}
