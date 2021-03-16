package direktiv

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/secrets"
	"github.com/vorteil/direktiv/pkg/secrets/ent"
	"github.com/vorteil/direktiv/pkg/secrets/ent/bucketsecret"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type secretsServer struct {
	secrets.UnimplementedSecretsServiceServer

	config *Config
	db     *ent.Client
	grpc   *grpc.Server
}

func newSecretsServer(config *Config) (*secretsServer, error) {

	db, err := ent.Open("postgres", config.SecretsAPI.DB)
	if err != nil {
		log.Errorf("can not connect to secrets db: %v", err)
		return nil, err
	}

	if err := db.Schema.Create(context.Background()); err != nil {
		log.Errorf("failed creating schema resources: %v", err)
		return nil, err
	}

	return &secretsServer{
		config: config,
		db:     db,
	}, nil

}

func getRegistries(db *dbManager, c *Config, client secrets.SecretsServiceClient, namespace string) (map[string]string, error) {

	r := make(map[string]string)

	// get default registry
	reg := c.FlowAPI.Registry

	if len(reg.Name) > 0 {
		r[reg.Name] = fmt.Sprintf("%s!%s", reg.User, reg.Token)
	}

	var d secrets.GetSecretsRequest
	d.Namespace = &namespace
	d.Stype = secrets.SecretTypes_REGISTRY.Enum()

	ss, err := client.GetSecretsWithData(context.Background(), &d)
	if err != nil {
		return r, err
	}

	// add all registries to map
	for _, s := range ss.Secrets {
		data, err := decryptData(db, namespace, s.GetData())
		if err != nil {
			return nil, err
		}
		r[s.GetName()] = string(data)
	}

	return r, nil

}

func (ss *secretsServer) start(s *WorkflowServer) error {
	return s.grpcStart(&ss.grpc, "secrets", s.config.SecretsAPI.Bind, func(srv *grpc.Server) {
		secrets.RegisterSecretsServiceServer(srv, ss)
	})
}

func (ss *secretsServer) StoreSecret(ctx context.Context, in *secrets.SecretsStoreRequest) (*empty.Empty, error) {

	var resp emptypb.Empty
	log.Debugf("store secret %v %v %v %v", in.Namespace, in.Name, len(in.Data), in.Stype)

	if in.Namespace == nil || in.Name == nil || len(in.Data) == 0 || in.Stype == nil {
		return nil, fmt.Errorf("all attributes are required")
	}

	bs, _ := ss.db.BucketSecret.
		Query().
		Where(
			bucketsecret.And(
				bucketsecret.NsEQ(in.GetNamespace()),
				bucketsecret.NameEQ(in.GetName()),
			)).
		Only(context.Background())

	if bs != nil {
		return nil, fmt.Errorf("secret already exists")
	}

	_, err := ss.db.BucketSecret.
		Create().
		SetName(in.GetName()).
		SetSecret(in.Data).
		SetType(int(in.GetStype())).
		SetNs(in.GetNamespace()).
		Save(ctx)

	return &resp, err

}

func (ss *secretsServer) RetrieveSecret(ctx context.Context, in *secrets.SecretsRetrieveRequest) (*secrets.SecretsRetrieveResponse, error) {

	var resp secrets.SecretsRetrieveResponse

	if in.Name == nil || in.Namespace == nil {
		return nil, fmt.Errorf("required attributes are missing")
	}

	bs, err := ss.db.BucketSecret.
		Query().
		Where(
			bucketsecret.And(
				bucketsecret.NsEQ(in.GetNamespace()),
				bucketsecret.NameEQ(in.GetName()),
			)).
		Only(context.Background())

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "secret '%s' not found", in.GetName())
		}
		return nil, err
	}

	resp.Data = bs.Secret
	return &resp, nil

}

func (ss *secretsServer) DeleteSecret(ctx context.Context, del *secrets.SecretsDeleteRequest) (*secrets.SecretsDeleteResponse, error) {

	var resp secrets.SecretsDeleteResponse

	if del.Name == nil || del.Namespace == nil || del.Stype == nil {
		return nil, fmt.Errorf("required attributes are missing")
	}

	c, err := ss.db.BucketSecret.
		Delete().
		Where(
			bucketsecret.And(
				bucketsecret.NsEQ(del.GetNamespace()),
				bucketsecret.NameEQ(del.GetName()),
				bucketsecret.TypeEQ(int(del.GetStype())),
			)).
		Exec(ctx)

	i := int32(c)
	resp.Count = &i

	return &resp, err

}

func (ss *secretsServer) GetSecretsWithData(ctx context.Context, in *secrets.GetSecretsRequest) (*secrets.GetSecretsDataResponse, error) {

	var (
		resp secrets.GetSecretsDataResponse
		ls   []*secrets.GetSecretsDataResponse_Secret
	)

	res, err := ss.db.BucketSecret.
		Query().
		Where(
			bucketsecret.And(
				bucketsecret.NsEQ(in.GetNamespace()),
				bucketsecret.TypeEQ(int(in.GetStype())),
			)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, bs := range res {
		ls = append(ls, &secrets.GetSecretsDataResponse_Secret{
			Name: &bs.Name,
			Data: bs.Secret,
		})
	}
	resp.Secrets = ls

	return &resp, nil

}

func (ss *secretsServer) GetSecrets(ctx context.Context, in *secrets.GetSecretsRequest) (*secrets.GetSecretsResponse, error) {

	dbs, err := ss.db.BucketSecret.
		Query().
		Where(
			bucketsecret.And(
				bucketsecret.NsEQ(in.GetNamespace()),
				bucketsecret.TypeEQ(int(in.GetStype())),
			)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	var ls []*secrets.GetSecretsResponse_Secret

	for _, s := range dbs {
		ls = append(ls, &secrets.GetSecretsResponse_Secret{
			Name: &s.Name,
		})
	}

	var resp secrets.GetSecretsResponse
	resp.Secrets = ls

	return &resp, nil
}

func (ss *secretsServer) name() string {
	return "secrets"
}

func (ss *secretsServer) stop() {

	if ss.grpc != nil {
		ss.grpc.GracefulStop()
	}

	if ss.db != nil {
		ss.db.Close()
	}

}

func decryptedDataForNS(ctx context.Context, instance *workflowLogicInstance, ns, name string) ([]byte, error) {

	var (
		dd   []byte
		resp *secrets.SecretsRetrieveResponse
	)

	resp, err := instance.engine.secretsClient.RetrieveSecret(ctx, &secrets.SecretsRetrieveRequest{
		Namespace: &ns,
		Name:      &name,
	})
	if err != nil {
		s := status.Convert(err)
		if s.Code() == codes.NotFound {
			return nil, NewUncatchableError("direktiv.secrets.notFound", "secret '%s' not found", name)
		}
		return nil, NewInternalError(err)
	}

	// decrypt data with key of namespace
	dd, err = decryptData(instance.engine.server.dbManager, ns, resp.GetData())
	if err != nil {
		return nil, NewInternalError(err)
	}

	return dd, nil

}

func decryptData(db *dbManager, ns string, data []byte) ([]byte, error) {

	namespace, err := db.getNamespace(ns)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(namespace.Key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := data[:gcm.NonceSize()]
	data = data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, data, nil)

}

func encryptData(key, data []byte) ([]byte, error) {

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil

}
