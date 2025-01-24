package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/pb33f/libopenapi/index"
	"gopkg.in/yaml.v3"
)

type DirektivOpenAPIFS struct {
	fileStore filestore.FileStore
	ns        string
	files     map[string]index.RolodexFile
}

type DirektivOpenAPIFile struct {
	file   *filestore.File
	reader *bytes.Reader
	name   string
	size   int64
	data   []byte
}

type DirektivRoloFile struct {
}

func (dr *DirektivRoloFile) GetContent() string {
	fmt.Println("GET CONTENT")
	return ""
}

func (dr *DirektivRoloFile) GetFileExtension() index.FileExtension {
	fmt.Println("GET ESXT")
	return index.YAML
}

func (dr *DirektivRoloFile) GetFullPath() string {
	fmt.Println("GET FULPATH")
	return ""
}

func (dr *DirektivRoloFile) GetErrors() []error {
	fmt.Println("GET ERRORS")
	return make([]error, 0)
}

func (dr *DirektivRoloFile) GetContentAsYAMLNode() (*yaml.Node, error) {
	fmt.Println("AS YAML")
	return &yaml.Node{}, nil
}

func (dr *DirektivRoloFile) GetIndex() *index.SpecIndex {
	fmt.Println("GET INDEX")
	return nil
}

func (dr *DirektivRoloFile) Name() string {
	fmt.Println("GET NAME")
	return "nil"
}

func (dr *DirektivRoloFile) ModTime() time.Time {
	return time.Now()
}

func (dr *DirektivRoloFile) IsDir() bool {
	return false
}

func (dr *DirektivRoloFile) Sys() any {
	return nil
}

func (dr *DirektivRoloFile) Size() int64 {
	fmt.Println("GET SIOZE")
	return 0
}

func (dr *DirektivRoloFile) Mode() os.FileMode {
	return 0700
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

	lf := &DirektivRoloFile{}

	d.files[name] = lf

	return &DirektivOpenAPIFile{
		reader: r,
		file:   file,
		size:   int64(len(data)),
		name:   filepath.Base(name),
		data:   data,
	}, nil
}

func (d *DirektivOpenAPIFS) GetFiles() map[string]index.RolodexFile {
	fmt.Println("FILES!!!!!")
	fmt.Println(d.files)
	return d.files
}

func (f *DirektivOpenAPIFile) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f *DirektivOpenAPIFile) Read(b []byte) (int, error) {
	return f.reader.Read(b)
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
