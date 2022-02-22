package flow

import (
	"context"
	"time"

	"github.com/direktiv/direktiv/pkg/model"
	secretsgrpc "github.com/direktiv/direktiv/pkg/secrets/grpc"
	"github.com/senseyeio/duration"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//
// README
//
// Here are the state logic implementations. If you're editing them or writing
// your own there are some things you should know.
//
// General Rules:
//
//   1. Under no circumstances should any functions here panic in production.
//	Panics here are not caught by the caller and will bring down the
//	server.
//
//   2. In all functions provided context.Context objects as an argument the
//	implementation must identify areas of logic that could run for a long
//	time and ensure that the logic can break out promptly if the context
// 	expires.

type stateTransition struct {
	NextState string
	Transform interface{}
}

type stateChild struct {
	Id   string
	Type string
}

type stateLogic interface {
	ID() string
	Type() string
	Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time
	ErrorCatchers() []model.ErrorDefinition
	Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error)
	LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild
	LogJQ() interface{}
	MetadataJQ() interface{}
}

//
// Helper functions
//

func deadlineFromString(ctx context.Context, engine *engine, im *instanceMemory, s string) time.Time {

	var t time.Time
	var d time.Duration

	d = time.Minute * 15

	if s != "" {
		dur, err := duration.ParseISO8601(s)
		if err != nil {
			engine.logToInstance(ctx, time.Now(), im.in, "Got an invalid ISO8601 timeout: %v", err)
		} else {
			now := time.Now()
			later := dur.Shift(now)
			d = later.Sub(now)
		}
	}

	t = time.Now()
	t = t.Add(d)
	t = t.Add(time.Second * 5)

	return t

}

func addSecrets(ctx context.Context, engine *engine, im *instanceMemory, m map[string]interface{}, secrets ...string) (map[string]interface{}, error) {

	var err error

	if len(secrets) > 0 {
		engine.logToInstance(ctx, time.Now(), im.in, "Decrypting secrets.")

		s := make(map[string]string)

		for _, name := range secrets {
			var dd []byte
			dd, err = getSecretsForInstance(ctx, engine, im, name)
			if err != nil {
				return nil, err
			}
			s[name] = string(dd)
		}

		m["secrets"] = s
	}

	return m, nil

}

func getSecretsForInstance(ctx context.Context, engine *engine, im *instanceMemory, name string) ([]byte, error) {

	var resp *secretsgrpc.SecretsRetrieveResponse

	namespace := im.in.Edges.Namespace.ID.String()

	resp, err := engine.secrets.client.RetrieveSecret(ctx, &secretsgrpc.SecretsRetrieveRequest{
		Namespace: &namespace,
		Name:      &name,
	})
	if err != nil {
		s := status.Convert(err)
		if s.Code() == codes.NotFound {
			return nil, NewUncatchableError("direktiv.secrets.notFound", "secret '%s' not found", name)
		}
		return nil, NewInternalError(err)
	}

	return resp.GetData(), nil

}

type multiactionTuple struct {
	ID       string
	Complete bool
	Type     string
	Attempts int
	Results  interface{}
}
