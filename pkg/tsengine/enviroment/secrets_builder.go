package enviroment

import (
	"context"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/compiler"
)

type SecretBuilder struct {
	secrets   map[string]string
	baseFS    string
	provider  SecretProvider
	namespace string
}

func NewSecretBuilder(fi compiler.FlowInformation, baseFS string, namespace string, provider SecretProvider) *SecretBuilder {
	secrets := make(map[string]string)
	for _, s := range fi.Secrets {
		secrets[s.Name] = ""
	}
	return &SecretBuilder{
		baseFS:    baseFS,
		secrets:   secrets,
		provider:  provider,
		namespace: namespace,
	}
}

func (b *SecretBuilder) Build() map[string]string {
	for name, _ := range b.secrets {
		slog.Debug("fetching secret", slog.String("secret", name))
		data, err := b.provider.Get(context.Background(), b.namespace, name)
		if err != nil {
			slog.Error("fetching secret failed", slog.String("secret", name))

			continue
		}
		b.secrets[name] = string(data)
	}
	return b.secrets
}
