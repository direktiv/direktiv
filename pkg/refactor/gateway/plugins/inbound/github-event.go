package inbound

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
)

const (
	GithubWebhookPluginName = "github-event"
)

type GithubWebhookPluginConfig struct {
	Secret        string   `mapstructure:"secret"          yaml:"secret"`
	ListenForType []string `mapstructure:"listen_for_type" yaml:"listen_for_type"`
}

type GithubWebhookPlugin struct {
	config *GithubWebhookPluginConfig
}

func ConfigureGithubWebhook(config interface{}, _ string) (core.PluginInstance, error) {
	requestConvertConfig := &GithubWebhookPluginConfig{}

	err := plugins.ConvertConfig(config, requestConvertConfig)
	if err != nil {
		return nil, err
	}

	return &GithubWebhookPlugin{
		config: requestConvertConfig,
	}, nil
}

func (p *GithubWebhookPlugin) Config() interface{} {
	return p.config
}

func (p *GithubWebhookPlugin) ExecutePlugin(_ *core.ConsumerFile, w http.ResponseWriter, r *http.Request) bool {
	eventType := r.Header.Get("X-GitHub-Event")
	signature := r.Header.Get("X-Hub-Signature-256")

	body, err := io.ReadAll(r.Body)
	// Replace the body with a new reader after reading from the original
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	if err != nil {
		slog.Error("can not read the body",
			slog.String("plugin", GithubWebhookPluginName))

		return false
	}

	if (p.config.Secret != "" || signature != "") && !p.verifySignature(body, signature) {
		slog.Warn("Got Github event with wrong signature", slog.String("plugin", GithubWebhookPluginName))
		plugins.ReportError(w, http.StatusUnauthorized,
			"github-event", fmt.Errorf("signature validation failed"))

		return false
	}
	if len(p.config.ListenForType) > 0 {
		var ret bool
		for _, v := range p.config.ListenForType {
			ret = ret || v == eventType
		}

		return ret
	}

	return true
}

func (p *GithubWebhookPlugin) verifySignature(payload []byte, signature string) bool {
	digest := hmac.New(sha256.New, []byte(p.config.Secret))
	_, _ = digest.Write([]byte(string(payload)))
	sig1 := "sha256=" + hex.EncodeToString(digest.Sum(nil))

	return hmac.Equal([]byte(sig1), []byte(signature))
}

func (*GithubWebhookPlugin) Type() string {
	return RequestConvertPluginName
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		GithubWebhookPluginName,
		plugins.InboundPluginType,
		ConfigureGithubWebhook))
}
