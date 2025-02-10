package secrets

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type Controller interface {
	// NOTE: controllers are very error tolerant. incorrectly configured
	// 		drivers, missing secrets, malformed references etc will not
	// 		return errors. Errors are only returned by underlying tech
	// 		such as NATS. In our mocked testing environment, errors are
	//		impossible.
	List(ctx context.Context) (List, error)
	Lookup(ctx context.Context, refs []SecretRef) (List, error)
	Delete() error
}

var ErrKeyExists = errors.New("key exists")

// NOTE: NATS can act as a distributed cache. It's not really made for this purpose
// 		and we might run into performance problems as a result. If that happens, we
// 		need to make a version of the Cache interface that caches locally in memory.

type Cache interface {
	List(ctx context.Context) (List, error)
	Insert(ctx context.Context, secret Secret) error // NOTE: this function must throw an ErrKeyExists error if it clashes with an existing secret.
	Delete() error
}

type Config struct {
	DefaultSource string         `json:"defaultSource"`
	RetryTime     time.Duration  `json:"retryTime"`
	SourceConfigs []SourceConfig `json:"sourceConfigs"`
}

func NewController(config *Config, cache Cache) Controller {
	// NOTE: for now the new controller returns a simple ephemeral controller.
	// 		this will remain true until we become concerned about performance
	// 		and need to consider caching.
	return NewEphemeralController(config, cache)
}

type ephemeralController struct {
	config *Config
	cache  Cache
}

func NewEphemeralController(config *Config, cache Cache) Controller {
	return &ephemeralController{
		config: config,
		cache:  cache,
	}
}

func (c *ephemeralController) List(ctx context.Context) (List, error) {
	return c.cache.List(ctx)
}

func (c *ephemeralController) lookup(ctx context.Context, refs []SecretRef) (List, []SecretRef, error) {
	fullList, err := c.List(ctx)
	if err != nil {
		return nil, nil, err
	}
	shortList := make(List, 0)
	notFound := make([]SecretRef, 0)

	for refIdx, ref := range refs {
		foundIdx := -1

		if ref.Source == "" || ref.Source == DefaultSourceString {
			ref.Source = c.config.DefaultSource
			refs[refIdx].Source = ref.Source
		}

		if ref.Path == "" {
			ref.Path = ref.Name
			refs[refIdx].Path = ref.Path
		}

		for entryIdx, entry := range fullList {
			if ref.Path != entry.Path {
				continue
			}

			if ref.Source != entry.Source {
				continue
			}

			foundIdx = entryIdx

			break
		}

		if foundIdx == -1 {
			notFound = append(notFound, refs[refIdx])
			shortList = append(shortList, Secret{
				Path:   ref.Path,
				Source: ref.Source,
				Data:   make([]byte, 0),
				Error:  errors.New("unfetcheds"),
			}) // add virtual item in case we have to return without looking it up
		} else {
			shortList = append(shortList, fullList[foundIdx])
		}
	}

	return shortList, notFound, nil
}

func (c *ephemeralController) touchRef(ctx context.Context, ref SecretRef) {
	src := c.mux(ref.Source)
	secretData, err := src.Get(ctx, ref.Path)
	secret := Secret{
		Path:   ref.Path,
		Source: ref.Source,
		Data:   secretData,
		Error:  err,
	}

	if err := c.cache.Insert(ctx, secret); err != nil {
		if errors.Is(err, ErrKeyExists) {
			return
		}

		// TODO: errors here are most likely to be unimportant clashes, we should check
		// 		to be sure to avoid needlessly polluting the logs.
		slog.Error("secret cache insert error", "err", err)
	}
}

func (c *ephemeralController) touch(ctx context.Context, refs ...SecretRef) {
	// NOTE: this function must return promptly
	for _, ref := range refs {
		go c.touchRef(ctx, ref)
	}
}

func (c *ephemeralController) Lookup(ctx context.Context, refs []SecretRef) (List, error) {
	for {
		shortList, notFound, err := c.lookup(ctx, refs)
		if err != nil {
			return nil, err
		}

		if len(notFound) == 0 {
			return shortList, nil
		}

		c.touch(ctx, notFound...)

		select {
		case <-ctx.Done():
			return shortList, nil
		case <-time.After(c.config.RetryTime):
			// NOTE: if we get here, context hasn't expired and we have not-found refs,
			//		so we should try and get them.
		}
	}
}

func (c *ephemeralController) mux(sourceName string) Source {
	if sourceName == DefaultSourceString {
		sourceName = c.config.DefaultSource
	}

	for _, conf := range c.config.SourceConfigs {
		if conf.Name != sourceName {
			continue
		}

		d, err := GetDriver(conf.Driver)
		if err != nil {
			return &NullDriverSource{
				Name: conf.Driver,
			}
		}

		s := d.ConstructSource(conf.Data)

		return s
	}

	return &NullSource{
		Name: sourceName,
	}
}

func (c *ephemeralController) Delete() error {
	return c.cache.Delete()
}

// TODO: enable NATS
// TODO: design apis next
// TODO: store/load source settings from database
//
// TODO: implement apis:
//		invalidate cache
//		define new source (don't forget multiple in same request ???)
//		rename source ???
//		change defaultSource
// 		delete source
//		list sources
//		list secrets in source ???
// TODO: implement source drivers
//				cyberark // STALLED: need an account.
//				azure // STALLED: need Jens' Azure code.
//				aws // STALLED: need an IAM account.
// TODO: documentation
// TODO: exported symbols documentation comments
// TODO: more extensive unit testing
// TODO: jest e2e tests
// TODO: figure out how to automate testing against third party services

/*

APIs:

List/Get sources on namespace
Delete a source
Create a source
	AWS
	Azure
	CyberArk

List secrets in namespace cache
Invalidate a cached secret
Force lookup (as a test only; do not return)

Get the namespace secrets config
Update the namespace secrets config

*/
