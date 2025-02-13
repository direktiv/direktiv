package flow

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/secrets"
	"github.com/direktiv/direktiv/pkg/secrets/localsecrets"
	"github.com/direktiv/direktiv/pkg/secrets/natscache"
	"github.com/nats-io/nats.go/jetstream"
)

func loadSource(rev []byte) (*model.Workflow, error) {
	workflow := new(model.Workflow)

	err := workflow.Load(rev)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}

// TODO: alan, should probably remove placeholder logic...
func (flow *flow) placeholdSecrets(ctx context.Context, tx *database.DB, ns string, file *filestore.File) error {
	data, err := tx.FileStore().ForFile(file).GetData(ctx)
	if err != nil {
		return err
	}

	wf, err := loadSource(data)
	if err != nil {
		return err
	}

	secretRefs := wf.GetSecretReferences()

	for _, secretRef := range secretRefs {
		_, err = tx.DataStore().Secrets().Get(ctx, ns, secretRef)
		if errors.Is(err, datastore.ErrNotFound) {
			err = tx.DataStore().Secrets().Set(ctx, &datastore.Secret{
				Namespace: ns,
				Name:      secretRef,
				Data:      nil,
			})
			if err != nil {
				continue
			}
		} else if err != nil {
			continue
		}
	}

	return nil
}

func (srv *server) initializeSecrets() error {
	slog.Debug("initializing secrets")

	// register drivers
	if err := secrets.RegisterDriver(localsecrets.DriverName, &localsecrets.Driver{
		SecretsStore: srv.db.DataStore().Secrets(),
	}); err != nil {
		return err
	}

	// set cache factory
	if srv.nats != nil {
		slog.Info("Configuring NATS for secrets cache.")

		secrets.SetDefaultCacheFactory(func(namespace string) (secrets.Cache, error) {
			js, err := jetstream.New(srv.nats)
			if err != nil {
				return nil, err
			}

			return natscache.New(js, namespace)
		})
	} else {
		slog.Warn("Using in-memory secrets cache because NATS is not configured.")
	}

	// set config getter
	secrets.SetDefaultConfigGetter(func(namespace string) (*secrets.Config, error) {
		var config *secrets.Config

		secretsConfigs, err := srv.db.DataStore().SecretsConfigs().Get(context.Background(), namespace)
		if err != nil {
			if !errors.Is(err, datastore.ErrNotFound) {
				return nil, err
			}

			config = DefaultSecretsConfig(namespace)
		} else {
			if err := json.Unmarshal(secretsConfigs.Configuration, &config); err != nil {
				return nil, err
			}
		}

		return config, nil
	})

	return nil
}

func DefaultSecretsConfig(namespace string) *secrets.Config {
	// initialize sensible default here
	confData, _ := json.Marshal(localsecrets.Config{
		DriverName: localsecrets.DriverName,
		Namespace:  namespace,
	})

	return &secrets.Config{
		DefaultSource: "local",
		RetryTime:     time.Second,
		SourceConfigs: []secrets.SourceConfig{
			{
				Name:   "local",
				Driver: localsecrets.DriverName,
				Data:   confData,
			},
		},
	}
}
