package utils

import (
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func NewMockGorm() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// filestore tables
	type File struct {
		filestore.File
		Revisions []filestore.Revision `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	}
	type Root struct {
		filestore.Root
		Files []File `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	}

	err = db.AutoMigrate(&Root{}, &File{}, &filestore.Revision{})
	if err != nil {
		return nil, err
	}

	// datastore tables
	err = db.AutoMigrate(&core.FileAttributes{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
