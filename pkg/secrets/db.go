package secrets

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	_ "github.com/lib/pq"
	toml "github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/secrets/ent"
	entc "github.com/vorteil/direktiv/pkg/secrets/ent"
	"github.com/vorteil/direktiv/pkg/secrets/ent/namespacesecret"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type dbHandler struct {
	db  *entc.Client
	key string
}

func (s *Server) setupDB() error {

	// DB, KEY
	cfgFile := os.Getenv(configFile)

	if cfgFile == "" {
		return fmt.Errorf("configuration file not provided")
	}

	b, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return err
	}

	c := new(dbConfig)

	err = toml.Unmarshal(b, c)
	if err != nil {
		return err
	}

	if c.DB == "" || c.Key == "" {
		return fmt.Errorf("DB and Key have to be set")
	}

	if len(c.Key) != 32 {
		return fmt.Errorf("key needs to be 32 characters")
	}

	db, err := ent.Open("postgres", c.DB)
	if err != nil {
		log.Errorf("can not connect to secrets db: %v", err)
		return err
	}

	if err := db.Schema.Create(context.Background()); err != nil {
		log.Errorf("failed creating schema resources: %v", err)
		return err
	}
	s.handler = &dbHandler{
		db:  db,
		key: c.Key,
	}

	log.Infof("secrets configured")

	return nil
}

func (db *dbHandler) AddSecret(namespace, name string, secret []byte) error {

	log.Infof("adding secret %s", name)

	bs, _ := db.db.NamespaceSecret.
		Query().
		Where(
			namespacesecret.And(
				namespacesecret.NsEQ(namespace),
				namespacesecret.NameEQ(name),
			)).
		Only(context.Background())

	if bs != nil {
		return fmt.Errorf("secret already exists")
	}

	d, err := encryptData([]byte(db.key), secret)
	if bs != nil {
		return fmt.Errorf("error encrypting data: %v", err)
	}

	_, err = db.db.NamespaceSecret.
		Create().
		SetName(name).
		SetSecret(d).
		SetNs(namespace).
		Save(context.Background())

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

func (db *dbHandler) GetSecrets(namespace string) ([]string, error) {

	var names []string

	dbs, err := db.db.NamespaceSecret.
		Query().
		Where(
			namespacesecret.And(
				namespacesecret.NsEQ(namespace),
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

	_, err := db.db.NamespaceSecret.
		Delete().
		Where(
			namespacesecret.And(
				namespacesecret.NsEQ(namespace),
				namespacesecret.NameEQ(name),
			)).
		Exec(context.Background())

	return err

}

func (db *dbHandler) RemoveSecrets(namespace string) error {

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
