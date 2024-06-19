package provider

import (
	"context"
	"log/slog"

	"github.com/direktiv/direktiv/pkg/tsengine/compiler"
)

type SecretBuilder struct {
	secrets   map[string]string
	baseFS    string
	provider  SecretProvider
	namespace string
}

func NewSecretBuilder(provider SecretProvider, namespace string, fi compiler.FlowInformation, baseFS string) *SecretBuilder {
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

func (b *SecretBuilder) Build(ctx context.Context) map[string]string {
	for name := range b.secrets {
		slog.Debug("fetching secret", slog.String("secret", name))
		data, err := b.provider.GetSecret(ctx, b.namespace, name)
		if err != nil {
			slog.Error("fetching secret failed", slog.String("secret", name))

			continue
		}
		b.secrets[name] = string(data)
	}
	return b.secrets
}
