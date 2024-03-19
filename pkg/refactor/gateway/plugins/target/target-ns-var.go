package target

import (
	"fmt"
	"io"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
)

const (
	TargetNamespaceVarPluginName = "target-namespace-var"
)

type NamespaceVarConfig struct {
	Namespace   string `mapstructure:"namespace"    yaml:"namespace"`
	Variable    string `mapstructure:"variable"     yaml:"variable"`
	ContentType string `mapstructure:"content_type" yaml:"content_type"`
}

type NamespaceVarPlugin struct {
	config *NamespaceVarConfig
}

func ConfigureNamespaceVarPlugin(config interface{}, ns string) (core.PluginInstance, error) {
	targetNamespaceVarConfig := &NamespaceVarConfig{}

	err := plugins.ConvertConfig(config, targetNamespaceVarConfig)
	if err != nil {
		return nil, err
	}

	if targetNamespaceVarConfig.Variable == "" {
		return nil, fmt.Errorf("variable required")
	}

	// set default to gateway namespace
	if targetNamespaceVarConfig.Namespace == "" {
		targetNamespaceVarConfig.Namespace = ns
	}

	// throw error if non magic namespace targets different namespace
	if targetNamespaceVarConfig.Namespace != ns && ns != core.MagicalGatewayNamespace {
		return nil, fmt.Errorf("plugin can not target different namespace")
	}

	return &NamespaceVarPlugin{
		config: targetNamespaceVarConfig,
	}, nil
}

func (tnv NamespaceVarPlugin) Config() interface{} {
	return tnv.config
}

func (tnv NamespaceVarPlugin) ExecutePlugin(_ *core.ConsumerFile,
	w http.ResponseWriter, r *http.Request,
) bool {
	// request failed if nil and response already written
	resp := doDirektivRequest(direktivNamespaceVarRequest, map[string]string{
		namespaceArg: tnv.config.Namespace,
		varArg:       tnv.config.Variable,
	}, w, r)
	if resp == nil {
		return false
	}

	// set headers from Direktiv
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))

	// overwrite content type
	if tnv.config.ContentType != "" {
		w.Header().Set("Content-Type", tnv.config.ContentType)
	}

	_, err := io.Copy(w, resp.Body)
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"can not serve variable", err)

		return false
	}
	resp.Body.Close()

	return true
}

func (tnv NamespaceVarPlugin) Type() string {
	return TargetNamespaceVarPluginName
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		TargetNamespaceVarPluginName,
		plugins.TargetPluginType,
		ConfigureNamespaceVarPlugin))
}
