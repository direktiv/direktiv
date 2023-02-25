package psql

import (
	"context"
	"fmt"
	"github.com/direktiv/direktiv/pkg/experimental/filesystem"
	"gorm.io/gorm"
	"path/filepath"
	"strings"
	"time"
)

type Namespace struct {
	ID   int64
	Name string

	Files []File `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	CreatedAt time.Time
	UpdatedAt time.Time

	db *gorm.DB
}

type NamespaceList []*Namespace

func (n *Namespace) CreateFile(ctx context.Context, path string, typ string, payload []byte) (filesystem.File, error) {
	path = filepath.Clean(path)
	depth := strings.Count(path, "/")
	if path == "/" {
		depth = 0
	}

	f := &File{
		Path:        path,
		Depth:       depth,
		Payload:     payload,
		IsDirectory: typ == "directory",
		NamespaceID: n.ID,
		db:          n.db,
	}
	res := n.db.WithContext(ctx).Create(f)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return f, nil
}

func (n *Namespace) GetFile(ctx context.Context, path string) (filesystem.File, error) {
	f := &File{}
	path = filepath.Clean(path)

	res := n.db.WithContext(ctx).Where("namespace_id", n.ID).Where("path = ?", path).First(f)
	if res.Error != nil {
		return nil, res.Error
	}
	f.db = n.db

	return f, nil
}

func (n *Namespace) ListPath(ctx context.Context, path string) ([]filesystem.File, error) {
	var list []File

	path = filepath.Clean(path)
	depth := strings.Count(path, "/")
	if path == "/" {
		depth = 0
	}

	// TODO: add namespace in condition.
	res := n.db.WithContext(ctx).Where("namespace_id", n.ID).Where("depth", depth+1).Where("path LIKE ?", path+"%", path).Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}
	var files []filesystem.File
	for i, _ := range list {
		list[i].db = n.db
		files = append(files, &list[i])
	}

	return files, nil
}

func (n *Namespace) GetName() string {
	return n.Name
}
