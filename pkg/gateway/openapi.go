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
	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
	"github.com/pb33f/libopenapi/datamodel"
	"github.com/pb33f/libopenapi/index"
)

type direktivOpenAPIFS struct {
	fileStore filestore.FileStore
	ns        string
}

type direktivOpenAPIFile struct {
	file   *filestore.File
	reader *bytes.Reader
	name   string
	data   []byte
}

func (d *direktivOpenAPIFS) Open(name string) (fs.File, error) {
	file, err := d.fileStore.ForNamespace(d.ns).GetFile(context.Background(), name)
	if err != nil {
		return nil, err
	}

	data, err := d.fileStore.ForFile(file).GetData(context.Background())
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(data)

	f := &direktivOpenAPIFile{
		reader: r,
		file:   file,
		name:   filepath.Base(name),
		data:   data,
	}
	return f, nil
}

func (d *direktivOpenAPIFS) GetFiles() map[string]index.RolodexFile {
	return make(map[string]index.RolodexFile)
}

func (f *direktivOpenAPIFile) Name() string {
	return f.name

}

func (f *direktivOpenAPIFile) ModTime() time.Time {
	return time.Now()
}

func (f *direktivOpenAPIFile) IsDir() bool {
	return f.file.Typ == filestore.FileTypeDirectory
}

func (f *direktivOpenAPIFile) Sys() any {
	return ""
}

func (f *direktivOpenAPIFile) Size() int64 {
	return int64(len(f.data))
}

func (f *direktivOpenAPIFile) Mode() os.FileMode {
	return 0700
}

func (f *direktivOpenAPIFile) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f *direktivOpenAPIFile) Read(b []byte) (int, error) {
	return f.reader.Read(b)
}

func (f *direktivOpenAPIFile) Close() error {
	f.reader = bytes.NewReader(f.data)
	return nil
}

type openAPIDoc struct {
	doc libopenapi.Document
}

func newOpenAPIDoc(ns, path, content string, fileStore filestore.FileStore) (*openAPIDoc, error) {
	if content == "" {
		content = fmt.Sprintf(`openapi: 3.0.0
info:
   title: %s
   version: "1.0.0"
paths:
`, ns)
	}

	doc, err := libopenapi.NewDocumentWithConfiguration([]byte(content),
		&datamodel.DocumentConfiguration{
			AllowFileReferences:   true,
			AllowRemoteReferences: true,
			BasePath:              filepath.Dir(path),
			AvoidIndexBuild:       true,
			LocalFS: &direktivOpenAPIFS{
				fileStore: fileStore,
				ns:        ns,
			},
		})

	if err != nil {
		return nil, err
	}

	return &openAPIDoc{
		doc: doc,
	}, nil
}

func (o *openAPIDoc) validate() []string {
	totalErrors := make([]string, 0)

	hlval, errs := validator.NewValidator(o.doc)
	if len(errs) > 0 {
		for i := range errs {
			totalErrors = append(totalErrors, errs[i].Error())
		}
	}

	_, valErrs := hlval.ValidateDocument()
	if len(valErrs) > 0 {
		for i := range valErrs {
			fmt.Println(valErrs[i].HowToFix)
			fmt.Println(valErrs[i].Message)
			fmt.Println(valErrs[i].Reason)
			// fmt.Println(valErrs[i].)
			// fmt.Println(valErrs[i].HowToFix)

			totalErrors = append(totalErrors, valErrs[i].Error())
		}
	}
	return totalErrors
}
