package database

import (
	_ "embed"
)

//go:embed db_schema.sql
var Schema string
