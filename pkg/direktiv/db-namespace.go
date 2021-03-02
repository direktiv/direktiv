package direktiv

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"

	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/ent/namespace"
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

	ns, err := db.dbEnt.Namespace.
		Create().
		SetID(name).
		SetKey(key).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return ns, nil

}

func (db *dbManager) deleteNamespace(ctx context.Context, name string) error {

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

	return nil

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
