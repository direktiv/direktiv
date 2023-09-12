package functions

import (
	"encoding/json"
	"fmt"
	"strings"

	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/mitchellh/hashstructure/v2"
	hash "github.com/mitchellh/hashstructure/v2"
)

func SanitizeLabel(s string) string {
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimSuffix(s, "/")
	s = strings.ReplaceAll(s, "_", "--")
	s = strings.ReplaceAll(s, "/", "-")

	if len(s) > 63 {
		s = s[:63]
	}

	return s
}

func fixAnnotations(in map[string]string) map[string]string {
	rm := make(map[string]string)

	for k, v := range in {
		rm[k] = SanitizeLabel(v)
	}

	return rm
}

// GenerateServiceName generates a knative name based on workflow details.
func GenerateServiceName(info *igrpc.FunctionsBaseInfo /* ns, wf, n string*/) (string, string, string) {
	var name, scope, hash string

	if info.GetWorkflow() != "" {
		scope = PrefixWorkflow
		name, hash = GenerateWorkflowServiceName(info)
	} else {
		scope = PrefixNamespace

		h, err := hashstructure.Hash(fmt.Sprintf("%s-%s",
			info.GetNamespaceName(), info.GetName()),
			hashstructure.FormatV2, nil)
		if err != nil {
			panic(err)
		}
		name = fmt.Sprintf("%s-%s", PrefixNamespace, fmt.Sprintf("%d", h))
		hash = ""
	}

	return SanitizeLabel(name), scope, hash
}

// GenerateWorkflowServiceName generates a knative name based on workflow details.
func GenerateWorkflowServiceName(info *igrpc.FunctionsBaseInfo) (string, string) {
	wf := info.GetWorkflow()
	fndef := fndefFromBaseInfo(info)

	strs := []string{
		fndef.Cmd, fndef.ID, fndef.Image,
		fmt.Sprintf("%v", fndef.Size), fmt.Sprintf("%v", fndef.Type),
		fmt.Sprintf("%v", info.GetEnvs()),
	}

	def, err := json.Marshal(strs)
	if err != nil {
		panic(err)
	}

	h, err := hash.Hash(fmt.Sprintf("%s-%s-%s", info.GetNamespace(), wf, def), hash.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	name := fmt.Sprintf("%s-%d", PrefixWorkflow, h)

	return SanitizeLabel(name), fmt.Sprintf("%v", h)
}

func fndefFromBaseInfo(info *igrpc.FunctionsBaseInfo) *model.ReusableFunctionDefinition {
	size := int(info.GetSize())

	return &model.ReusableFunctionDefinition{
		Cmd:   info.GetCmd(),
		ID:    info.GetName(),
		Image: info.GetImage(),
		Size:  model.Size(size),
		Type:  model.ReusableContainerFunctionType,
	}
}

// AssembleWorkflowServiceName generates a knative name based on workflow details.
func AssembleWorkflowServiceName(hash uint64) string {
	return SanitizeLabel(fmt.Sprintf("%s-%d", PrefixWorkflow, hash))
}
