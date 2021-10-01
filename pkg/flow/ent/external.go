package ent

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"
)

// DB exports the underlying DB driver
func (c *Client) DB() *sql.DB {
	return c.driver.(*entsql.Driver).DB()
}
