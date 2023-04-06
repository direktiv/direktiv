package handler

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/dlog"
	"github.com/direktiv/direktiv/pkg/secrets/ent"
	entc "github.com/direktiv/direktiv/pkg/secrets/ent"
	"github.com/direktiv/direktiv/pkg/secrets/ent/namespacesecret"
	_ "github.com/lib/pq" // lib needs to imported for postgres
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

/* #nosec */
const (
	secretsConn = "DIREKTIV_SECRETS_DB"
	secretsKey  = "DIREKTIV_SECRETS_KEY"
)

type dbHandler struct {
	db  *entc.Client
	key string
}

var logger *zap.SugaredLogger

func setupDB() (SecretsHandler, error) {
	var err error

	dbEnv := os.Getenv(secretsConn)
	keyEnv := os.Getenv(secretsKey)

	logger, err = dlog.ApplicationLogger("secrets-db")
	if err != nil {
		return nil, err
	}

	if keyEnv == "" || dbEnv == "" {
		return nil, fmt.Errorf("DB and Key have to be set")
	}

	if len(keyEnv) != 32 {
		return nil, fmt.Errorf("key needs to be 32 characters")
	}

	db, err := ent.Open("postgres", dbEnv)
	if err != nil {
		logger.Errorf("can not connect to secrets db: %v", err)
		return nil, err
	}

	if err := db.Schema.Create(context.Background()); err != nil {
		logger.Errorf("failed creating schema resources: %v", err)
		return nil, err
	}

	return &dbHandler{
		db:  db,
		key: keyEnv,
	}, err
}

func (db *dbHandler) AddSecret(namespace, name string, secret []byte, ignoreError bool) error {
	logger.Infof("adding secret %s", name)

	bs, _ := db.db.NamespaceSecret.
		Query().
		Where(
			namespacesecret.And(
				namespacesecret.NsEQ(namespace),
				namespacesecret.NameEQ(name),
			)).
		Only(context.Background())

	if bs != nil && ignoreError {
		return nil
	}
	if bs != nil {
		return fmt.Errorf("secret already exists")
	}

	var d []byte
	var err error
	if !strings.HasSuffix(name, "/") { // dont excrypt when folder cause  folder have empty data
		d, err = encryptData([]byte(db.key), secret)
		if err != nil {
			return fmt.Errorf("error encrypting data: %w", err)
		}
	}

	_, err = db.db.NamespaceSecret.
		Create().
		SetName(name).
		SetSecret(d).
		SetNs(namespace).
		Save(context.Background())

	if err != nil {
		return err
	}
	return err
}

func (db *dbHandler) UpdateSecret(namespace, name string, secret []byte) error {
	logger.Infof("adding secret %s", name)

	bs, err := db.db.NamespaceSecret.
		Query().
		Where(
			namespacesecret.And(
				namespacesecret.NsEQ(namespace),
				namespacesecret.NameEQ(name),
			)).
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return status.Errorf(codes.NotFound, "secret '%s' not found", name)
		}
		return err
	}

	var d []byte
	d, err = encryptData([]byte(db.key), secret)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	_, err = db.db.NamespaceSecret.
		UpdateOne(bs).
		SetSecret(d).
		Save(context.Background())
	if err != nil {
		return fmt.Errorf("failed to store encrypted secret: %w", err)
	}

	return err
}

func (db *dbHandler) GetSecret(namespace, name string) ([]byte, error) {
	bs, err := db.db.NamespaceSecret.
		Query().
		Where(
			namespacesecret.And(
				namespacesecret.NsEQ(namespace),
				namespacesecret.NameEQ(name),
			)).
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "secret '%s' not found", name)
		}
		return nil, err
	}

	return decryptData([]byte(db.key), bs.Secret)
}

func (db *dbHandler) GetSecrets(namespace string, name string) ([]string, error) {
	var names []string
	name = strings.TrimPrefix(name, "/")

	dbs, err := db.db.NamespaceSecret.
		Query().
		Where(
			namespacesecret.And(
				namespacesecret.NsEQ(namespace),
				namespacesecret.NameHasPrefix(name),
			)).
		All(context.Background())
	if err != nil {
		return nil, err
	}

	pathPatternFolders := name + "*/"
	pathPatternFiles := name + "*"

	for _, s := range dbs {
		matchesPathPatternFolders, _ := filepath.Match(pathPatternFolders, s.Name)
		matchesPathPatternFiles, _ := filepath.Match(pathPatternFiles, s.Name)

		if matchesPathPatternFiles || matchesPathPatternFolders {
			names = append(names, s.Name)
		}
	}

	return names, nil
}

func (db *dbHandler) SearchForName(namespace string, name string) ([]string, error) {
	var names []string
	name = strings.TrimPrefix(name, "/")

	dbs, err := db.db.NamespaceSecret.
		Query().
		Where(
			namespacesecret.And(
				namespacesecret.NsEQ(namespace),
				namespacesecret.NameContains(name),
			)).
		All(context.Background())
	if err != nil {
		return nil, err
	}

	for _, s := range dbs {
		names = append(names, s.Name)
	}

	return names, nil
}

func (db *dbHandler) RemoveSecret(namespace, name string) error {
	// check if secret is already existing
	_, err := db.GetSecret(namespace, name)
	if err != nil {
		return err
	}

	_, err = db.db.NamespaceSecret.
		Delete().
		Where(
			namespacesecret.And(
				namespacesecret.NsEQ(namespace),
				namespacesecret.NameEQ(name),
			)).
		Exec(context.Background())

	return err
}

func (db *dbHandler) RemoveFolder(namespace, name string) error {
	if !strings.HasSuffix(name, "/") {
		return status.Errorf(codes.InvalidArgument, "secrets requested, expected folder")
	}
	_, err := db.db.NamespaceSecret.
		Delete().
		Where(
			namespacesecret.And(
				namespacesecret.NsEQ(namespace),
				namespacesecret.NameHasPrefix(name),
			)).
		Exec(context.Background())
	return err
}

func (db *dbHandler) RemoveNamespaceSecrets(namespace string) error {
	_, err := db.db.NamespaceSecret.
		Delete().
		Where(
			namespacesecret.And(
				namespacesecret.NsEQ(namespace),
			)).
		Exec(context.Background())

	return err
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

func decryptData(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
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
