package sqlitewrapper

import (
	"database/sql"
	"database/sql/driver"

	"github.com/pkg/errors"
	"modernc.org/sqlite"
)

// Use only for testing
// https://github.com/ent/ent/discussions/1667
type sqliteDriver struct {
	*sqlite.Driver
}

func (d sqliteDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.Driver.Open(name)
	if err != nil {
		return conn, err
	}
	c := conn.(interface {
		Exec(stmt string, args []driver.Value) (driver.Result, error)
	})
	if _, err := c.Exec("PRAGMA foreign_keys = on;", nil); err != nil {
		conn.Close()
		return nil, errors.Wrap(err, "failed to enable foreign keys")
	}
	return conn, nil
}

func init() {
	sql.Register("sqlite3", sqliteDriver{Driver: &sqlite.Driver{}})
}
