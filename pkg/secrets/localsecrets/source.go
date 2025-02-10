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

	if src.Config.DriverName != DriverName {
		return &secrets.BadConfigSource{
			Name:  DriverName,
			Error: fmt.Errorf("invalid driver name: '%s'", src.Config.DriverName),
		}
	}

	if src.Config.Namespace == "" {
		return &secrets.BadConfigSource{
			Name:  DriverName,
			Error: errors.New("missing namespace"),
		}
	}

	src.SecretsStore = d.SecretsStore

	return src
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
