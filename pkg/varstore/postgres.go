package varstore

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"io/ioutil"
	"strings"
)

type postgres struct {
	db *sql.DB
}

func (pg *postgres) connect(database string) error {

	var err error

	pg.db, err = sql.Open("postgres", database)
	if err != nil {
		return err
	}

	return nil

}

func NewPostgresVarStorage(database string) (VarStorage, error) {

	pg := new(postgres)

	err := pg.connect(database)
	if err != nil {
		return nil, err
	}

	err = pg.init()
	if err != nil {
		return nil, err
	}

	return pg, nil

}

func (pg *postgres) init() error {

	tx, err := pg.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`create table if not exists variables (
		id serial primary key,
		scope text,
		size bigint,
		key text,
		val bytea
	)`)
	if err != nil {
		return err
	}

	return tx.Commit()

}

func scopeString(scope ...string) string {
	return strings.Join(scope, ".")
}

func (pg *postgres) Close() error {
	return nil
}

type varInfo struct {
	key  string
	size int64
}

func (vi *varInfo) Key() string {
	return vi.key
}

func (vi *varInfo) Size() int64 {
	return vi.size
}

func (pg *postgres) List(ctx context.Context, scope ...string) ([]VarInfo, error) {

	rows, err := pg.db.QueryContext(ctx,
		`SELECT key, size FROM variables WHERE scope = $1 ORDER BY key ASC`,
		scopeString(scope...),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var vis []VarInfo

	for rows.Next() {

		var key string
		var size int64

		err = rows.Scan(&key, &size)
		if err != nil {
			return nil, err
		}

		vis = append(vis, &varInfo{
			key:  key,
			size: size,
		})

	}

	return vis, nil

}

type varReader struct {
	size int64
	io.Reader
	io.Closer
}

func (vr *varReader) Size() int64 {
	return vr.size
}

func (pg *postgres) Retrieve(ctx context.Context, key string, scope ...string) (VarReader, error) {

	vr := new(varReader)

	row := pg.db.QueryRowContext(ctx,
		`SELECT size, val FROM variables WHERE scope = $1 AND key = $2`,
		scopeString(scope...),
		key,
	)

	var size int64
	var data []byte

	err := row.Scan(&size, &data)
	if err != nil {
		if err == sql.ErrNoRows {
			data = make([]byte, 0)
			size = 0
		} else {
			return nil, err
		}
	}

	vr.size = size
	buf := bytes.NewReader(data)
	rc := ioutil.NopCloser(buf)
	vr.Reader = rc
	vr.Closer = rc

	return vr, nil

}

type varWriter struct {
	closed bool
	pg     *postgres
	ctx    context.Context
	key    string
	scope  []string
	buf    *bytes.Buffer
	io.Writer
}

func (vw *varWriter) Close() error {

	if vw.closed {
		return nil
	}
	vw.closed = true

	tx, err := vw.pg.db.BeginTx(vw.ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(vw.ctx,
		`SELECT COUNT(*) FROM variables WHERE scope = $1 AND key = $2`,
		scopeString(vw.scope...),
		vw.key,
	)

	var k int
	err = row.Scan(&k)
	if err != nil {
		return err
	}

	if k == 0 {
		_, err = tx.ExecContext(vw.ctx,
			`INSERT INTO variables (scope, size, key, val) values($1, $2, $3, $4)`,
			scopeString(vw.scope...),
			vw.buf.Len(),
			vw.key,
			vw.buf.Bytes(),
		)
	} else {
		_, err = vw.pg.db.ExecContext(vw.ctx,
			`UPDATE variables SET size = $1, val = $2 WHERE scope = $3 AND key = $4`,
			vw.buf.Len(),
			vw.buf.Bytes(),
			scopeString(vw.scope...),
			vw.key,
		)
	}
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil

}

func (pg *postgres) Store(ctx context.Context, key string, scope ...string) (io.WriteCloser, error) {

	vw := new(varWriter)
	vw.ctx = ctx
	vw.pg = pg
	vw.key = key
	vw.scope = scope
	vw.buf = new(bytes.Buffer)
	vw.Writer = vw.buf

	return vw, nil

}

func (pg *postgres) Delete(ctx context.Context, key string, scope ...string) error {

	_, err := pg.db.ExecContext(ctx,
		`DELETE FROM variables WHERE scope = $1 AND key = $2`,
		scopeString(scope...),
		key,
	)
	if err != nil {
		return err
	}

	return nil

}

func (pg *postgres) DeleteAllInScope(ctx context.Context, scope ...string) error {

	_, err := pg.db.ExecContext(ctx,
		`DELETE FROM variables WHERE (scope LIKE $1)`,
		scopeString(scope...)+"%",
	)
	if err != nil {
		return err
	}

	return nil

}
