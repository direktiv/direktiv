package direktiv

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"

	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/ent/namespace"
	secretsgrpc "github.com/vorteil/direktiv/pkg/secrets/grpc"

	log "github.com/sirupsen/logrus"
)

func (db *dbManager) getNamespace(name string) (*ent.Namespace, error) {

	ns, err := db.dbEnt.Namespace.
		Query().
		Where(namespace.IDEQ(name)).
		Only(db.ctx)

	if err != nil {
		return nil, err
	}

	return ns, nil

}

func (db *dbManager) addNamespace(ctx context.Context, name string) (*ent.Namespace, error) {

	key := make([]byte, 32)
	_, err := rand.Read(key)

	if err != nil {
		return nil, err
	}

	tx, err := db.dbEnt.Tx(db.ctx)
	if err != nil {
		return nil, err
	}

	ns, err := tx.Namespace.
		Create().
		SetID(name).
		SetKey(key).
		Save(ctx)
	if err != nil {
		return nil, rollback(tx, err)
	}

	err = kubernetesActionServiceAccount(name, true)
	if err != nil {
		return nil, rollback(tx, err)
	}

	return ns, tx.Commit()

}

func (db *dbManager) deleteNamespace(ctx context.Context, name string) error {

	// delete all workflows
	wfs, err := db.getWorkflows(ctx, name, 0, 0)
	if err != nil {
		log.Errorf("can not get workflows for namespace %s", name)
	}

	for _, w := range wfs {
		err := db.deleteWorkflow(ctx, w.ID.String())
		if err != nil {
			log.Errorf("can not delete workflow %s from namespace %s", w.Name, name)
		}
	}

	i, err := db.dbEnt.Namespace.
		Delete().
		Where(namespace.IDEQ(name)).
		Exec(ctx)

	if err != nil {
		return err
	}

	if i == 0 {
		return fmt.Errorf("namespace %s does not exist", name)
	}

	err = kubernetesActionServiceAccount(name, false)
	if err != nil {
		// we can still proceed
		log.Errorf("can not delete kubernetes service account: %v", err)
	}

	// delete secrets from secrets backend
	dr := &secretsgrpc.DeleteSecretsRequest{
		Namespace: &name,
	}
	_, err = db.secretsClient.DeleteSecrets(context.Background(), dr)

	return err

}

func (db *dbManager) getNamespaces(ctx context.Context, offset, limit int) ([]*ent.Namespace, error) {

	if limit == 0 {
		limit = math.MaxInt32
	}

	ns, err := db.dbEnt.Namespace.
		Query().
		Limit(limit).
		Offset(offset).
		Order(ent.Asc(namespace.FieldID)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return ns, nil

}
