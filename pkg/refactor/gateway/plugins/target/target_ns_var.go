package target

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"net/url"
// 	"os"

// 	"github.com/direktiv/direktiv/pkg/refactor/core"
// 	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
// 	"github.com/mitchellh/mapstructure"
// 	"github.com/pkg/errors"
// )

// const (
// 	TargetNamespaceVarPluginName = "target-namespace-var"
// )

// type TargetNamespaceVarConfig struct {
// 	Namespace string `yaml:"namespace"`
// 	Variable  string `yaml:"variable"`
// }

// // TargetFlowVarPlugin returns a namespace variable
// type TargetNamespaceVarPlugin struct {
// 	config *TargetNamespaceVarConfig
// }

// func (tnv TargetNamespaceVarPlugin) Configure(config interface{}) (plugins.PluginInstance, error) {
// 	targetNamespaceVarConfig := &TargetNamespaceVarConfig{}

// 	if config != nil {
// 		err := mapstructure.Decode(config, &targetNamespaceVarConfig)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "configuration for target-ns-var invalid")
// 		}
// 	}

// 	// set default to gateway namespace
// 	if targetNamespaceVarConfig.Namespace == "" {
// 		targetNamespaceVarConfig.Namespace = core.MagicalGatewayNamespace
// 	}

// 	return &TargetNamespaceVarPlugin{
// 		config: targetNamespaceVarConfig,
// 	}, nil
// }

// func (tnv TargetNamespaceVarPlugin) Config() interface{} {
// 	return tnv.config
// }

// func (tnv TargetNamespaceVarPlugin) Name() string {
// 	return TargetNamespaceVarPluginName
// }

// func (tnv TargetNamespaceVarPlugin) Type() plugins.PluginType {
// 	return plugins.InboundPluginType
// }

// func (tnv TargetNamespaceVarPlugin) ExecutePlugin(ctx context.Context, c *core.Consumer,
// 	w http.ResponseWriter, r *http.Request) bool {

// 	url, err := createURLNamespaceVar(tnv.config.Namespace, tnv.config.Variable)
// 	if err != nil {
// 		plugins.ReportError(w, http.StatusInternalServerError,
// 			"can not create url", err)
// 		return false
// 	}

// 	client := http.Client{}

// 	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
// 	if err != nil {
// 		plugins.ReportError(w, http.StatusInternalServerError,
// 			"can not create request", err)
// 		return false
// 	}

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		plugins.ReportError(w, http.StatusInternalServerError,
// 			"can not serve variable", err)
// 		return false
// 	}

// 	// set headers from Direktiv
// 	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
// 	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))

// 	_, err = io.Copy(w, resp.Body)
// 	if err != nil {
// 		plugins.ReportError(w, http.StatusInternalServerError,
// 			"can not serve variable", err)
// 		return false
// 	}
// 	resp.Body.Close()

// 	return true
// }

// //nolint:gochecknoinits
// func init() {
// 	plugins.AddPluginToRegistry(TargetNamespaceVarPlugin{})
// }

// func createURLNamespaceVar(ns, v string) (*url.URL, error) {
// 	urlString := fmt.Sprintf("http://localhost:%s/api/namespaces/%s/vars/%s",
// 		os.Getenv("DIREKTIV_API_V1_PORT"), ns, v)

// 	return url.Parse(urlString)
// }
