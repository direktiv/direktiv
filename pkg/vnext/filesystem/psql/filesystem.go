package psql

import (
	"context"
	"fmt"
	"github.com/direktiv/direktiv/pkg/vnext/filesystem"
	"gorm.io/gorm"
)

var _ filesystem.Namespace = &Namespace{}

type sqlFilesystem struct {
	db *gorm.DB
}

func NewSqlFilesystem(db *gorm.DB) filesystem.Filesystem {
	return &sqlFilesystem{
		db: db,
	}
}

func (s sqlFilesystem) CreateNamespace(ctx context.Context, namespace string) (filesystem.Namespace, error) {
	n := &Namespace{Name: namespace}
	res := s.db.WithContext(ctx).Create(n)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}
	n.db = s.db

	return n, nil
}

func (s sqlFilesystem) GetNamespace(ctx context.Context, namespace string) (filesystem.Namespace, error) {
	n := &Namespace{}
	res := s.db.WithContext(ctx).Where("name = ?", namespace).First(n)
	if res.Error != nil {
		return nil, res.Error
	}
	n.db = s.db

	return n, nil
}

func (s sqlFilesystem) GetAllNamespaces(ctx context.Context) ([]filesystem.Namespace, error) {
	var list []Namespace
	res := s.db.WithContext(ctx).Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	var ns []filesystem.Namespace
	for i, _ := range list {
		list[i].db = s.db
		ns = append(ns, &list[i])
	}

	return ns, nil
}

func (s sqlFilesystem) DeleteNamespace(ctx context.Context, namespace string) error {
	res := s.db.WithContext(ctx).Where("name = ?", namespace).Delete(&Namespace{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpedted gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}
	return nil
}

var _ filesystem.Filesystem = &sqlFilesystem{}
