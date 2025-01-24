package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/pb33f/libopenapi/index"
)

type DirektivOpenAPIFS struct {
	fileStore filestore.FileStore
	ns        string
}

type DirektivOpenAPIFile struct {
	file   *filestore.File
	reader *bytes.Reader
	name   string
	size   int64
	data   []byte
}

func (d *DirektivOpenAPIFS) Open(name string) (fs.File, error) {
	fmt.Printf("RESOLVE %v\n", name)
	file, err := d.fileStore.ForNamespace(d.ns).GetFile(context.Background(), name)
	if err != nil {
		return nil, err
	}

	data, err := d.fileStore.ForFile(file).GetData(context.Background())
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(data)

	return &DirektivOpenAPIFile{
		reader: r,
		file:   file,
		size:   int64(len(data)),
		name:   filepath.Base(name),
		data:   data,
	}, nil
}

func (d *DirektivOpenAPIFS) GetFiles() map[string]index.RolodexFile {
	return make(map[string]index.RolodexFile)
}

func (f *DirektivOpenAPIFile) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f *DirektivOpenAPIFile) Read(b []byte) (int, error) {

	d, e := f.reader.Read(b)
	fmt.Println(string(b))
	return d, e
	// return f.reader.Read(b)
}

func (f *DirektivOpenAPIFile) Close() error {
	f.reader = bytes.NewReader(f.data)
	return nil
}

func (f *DirektivOpenAPIFile) Name() string {
	return f.name
}

func (f *DirektivOpenAPIFile) Size() int64 {
	return f.size
}

func (f *DirektivOpenAPIFile) Mode() fs.FileMode {
	return 0700
}

func (f *DirektivOpenAPIFile) ModTime() time.Time {
	return f.file.UpdatedAt
}

func (f *DirektivOpenAPIFile) IsDir() bool {
	return f.file.Typ == filestore.FileTypeDirectory
}

func (f *DirektivOpenAPIFile) Sys() any {
	return ""
}
