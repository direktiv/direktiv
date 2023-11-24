package target

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
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

func ConfigureNamespaceVarPlugin(config interface{}, ns string) (plugins.PluginInstance, error) {
	targetNamespaceVarConfig := &NamespaceVarConfig{}

	err := plugins.ConvertConfig(config, targetNamespaceVarConfig)
	if err != nil {
		return nil, err
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

func (tnv NamespaceVarPlugin) ExecutePlugin(_ *spec.ConsumerFile,
	w http.ResponseWriter, r *http.Request,
) bool {
	url, err := createURLNamespaceVar(tnv.config.Namespace, tnv.config.Variable)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not create url", err)

		return false
	}

	client := http.Client{}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url.String(), nil)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not create request", err)

		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not serve variable", err)

		return false
	}

	// set headers from Direktiv
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))

	// overwrite content type
	if tnv.config.ContentType != "" {
		w.Header().Set("Content-Type", tnv.config.ContentType)
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not serve variable", err)

		return false
	}
	resp.Body.Close()

	return true
}

func createURLNamespaceVar(ns, v string) (*url.URL, error) {
	urlString := fmt.Sprintf("http://localhost:%s/api/namespaces/%s/vars/%s",
		os.Getenv("DIREKTIV_API_V1_PORT"), ns, v)

	return url.Parse(urlString)
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		TargetNamespaceVarPluginName,
		plugins.TargetPluginType,
		ConfigureNamespaceVarPlugin))
}
