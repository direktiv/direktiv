package localsecrets

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/secrets"
)

// NOTE: the local driver behaves identically to the other drivers
// 		critically, this means that the List function won't know
// 		about a secret unless it has been queried recently. This
// 		makes this driver inappropriate for use in APIs managing
// 		the contents of the local secrets database... We could
// 		continue to rely on the existing APIs for this, or we could
// 		treat this driver specially. The choice is unclear...

const (
	DriverName = "database"
)

type Driver struct {
	SecretsStore datastore.SecretsStore
}

type Config struct {
	DriverName string
	Namespace  string
}

func (d *Driver) ConstructSource(data []byte) secrets.Source {
	src := new(Source)

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&src.Config); err != nil {
		return &secrets.BadConfigSource{
			Name:  DriverName,
			Error: err,
		}
	}

	if err := d.ValidateConfig(data); err != nil {
		return &secrets.BadConfigSource{
			Name:  DriverName,
			Error: err,
		}
	}

	src.SecretsStore = d.SecretsStore

	return src
}

func (d *Driver) RedactConfig(data []byte) ([]byte, error) {
	return data, nil // no sensitive information is stored in this config
}

func (d *Driver) ValidateConfig(data []byte) error {
	var config Config

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&config); err != nil {
		return err
	}

	if config.DriverName != DriverName {
		return fmt.Errorf("invalid driver name: '%s'", config.DriverName)
	}

	if config.Namespace == "" {
		return errors.New("missing namespace")
	}

	return nil
}

type Source struct {
	Config       Config
	SecretsStore datastore.SecretsStore
}

func (s *Source) Get(ctx context.Context, path string) ([]byte, error) {
	secret, err := s.SecretsStore.Get(ctx, s.Config.Namespace, path)
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			return nil, secrets.ErrSecretNotFound
		}

		return nil, secrets.NewJSONMarshalableError(err)
	}

	return secret.Data, nil
}
