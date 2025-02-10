package local

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/direktiv/direktiv/pkg/secrets"
)

const (
	DriverName = "aws"
)

type Driver struct {
}

type Config struct {
	DriverName string
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

	config := aws.NewConfig()

	// TODO: determine what kinds of config we need/want to support
	// NOTE: have found the following options in aws library code:
	/*
		WithCredentialsChainVerboseErrors
		WithCredentials
		WithEndpoint
		WithEndpointResolver
		WithRegion
		WithDisableSSL
		WithHTTPClient
		WithMaxRetries
		WithDisableParamValidation
		WithDisableComputeChecksums
		WithLogLevel
		WithLogger
		WithS3ForcePathStyle
		WithS3Disable100Continue
		WithS3UseAccelerate
		WithS3DisableContentMD5Validation
		WithS3UseARNRegion
		WithUseDualStack
		WithUseFIPSEndpoint
		WithEC2MetadataDisableTimeoutOverride
		WithEC2MetadataEnableFallback
		WithSleepDelay
		WithEndpointDiscovery
		WithDisableEndpointHostPrefix
		WithSTSRegionalEndpoint
		WithS3UsEast1RegionalEndpoint
		WithLowerCaseHeaderMaps
		WithDisableRestProtocolURICleaning
	*/

	var err error

	src.session, err = session.NewSession() // TODO: determine what configs we need to support here
	if err != nil {
		return &secrets.BadConfigSource{
			Name:  DriverName,
			Error: fmt.Errorf("driver error: '%w'", err),
		}
	}

	src.sm = secretsmanager.New(src.session, config)

	return src
}

type Source struct {
	Config  Config
	session *session.Session
	sm      *secretsmanager.SecretsManager
}

func (s *Source) Get(ctx context.Context, path string) ([]byte, error) {
	output, err := s.sm.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &path,
	})
	if err != nil {
		return nil, err

		// TODO: figure out how to translate aws errors into our errors
	}

	var data []byte

	if output.SecretString != nil {
		data = []byte(*output.SecretString)
	} else {
		data = output.SecretBinary
	}

	return data, nil
}
