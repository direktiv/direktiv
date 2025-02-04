package gateway

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
	"github.com/pb33f/libopenapi/bundler"
	"github.com/pb33f/libopenapi/datamodel"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/index"
	"gopkg.in/yaml.v3"
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
	fmt.Printf("RESOLVE %v\n", name)

	ff := filepath.Join("/home/jens/go/src/github/jensg-st/test-flow", name)

	return os.Open(ff)
	// file, err := d.fileStore.ForNamespace(d.ns).GetFile(context.Background(), name)
	// if err != nil {
	// 	return nil, err
	// }

	// data, err := d.fileStore.ForFile(file).GetData(context.Background())
	// if err != nil {
	// 	return nil, err
	// }

	// r := bytes.NewReader(data)

	// f := &direktivOpenAPIFile{
	// 	reader: r,
	// 	file:   file,
	// 	name:   filepath.Base(name),
	// 	data:   data,
	// }
	// return f, nil
}

func (d *direktivOpenAPIFS) GetFiles() map[string]index.RolodexFile {
	fmt.Println("GETFILES")
	return make(map[string]index.RolodexFile)
}

func (f *direktivOpenAPIFile) Name() string {
	fmt.Printf("NAME %v\n", f.name)
	return f.name
}

func (f *direktivOpenAPIFile) ModTime() time.Time {
	fmt.Printf("MODETIME %v\n", f.file.UpdatedAt)
	return f.file.UpdatedAt
}

func (f *direktivOpenAPIFile) IsDir() bool {
	fmt.Printf("ISIDR %v\n", f.file.Typ == filestore.FileTypeDirectory)
	return f.file.Typ == filestore.FileTypeDirectory
}

func (f *direktivOpenAPIFile) Sys() any {
	fmt.Println("SYS")
	return ""
}

func (f *direktivOpenAPIFile) Size() int64 {
	fmt.Println("SIZE")
	return int64(len(f.data))
}

func (f *direktivOpenAPIFile) Mode() os.FileMode {
	fmt.Println("MODE")
	return 0777
}

func (f *direktivOpenAPIFile) Stat() (fs.FileInfo, error) {
	fmt.Println("STAT")
	return f, nil
}

func (f *direktivOpenAPIFile) Read(b []byte) (int, error) {
	fmt.Println("READ")

	a, c := f.reader.Read(b)
	fmt.Println(string(b))
	fmt.Println(c)
	return a, c
	// return f.reader.Read(b)
}

func (f *direktivOpenAPIFile) Close() error {
	fmt.Println("CLOSE")
	f.reader = bytes.NewReader(f.data)
	return nil
}

type openAPIDoc struct {
	doc   libopenapi.Document
	model *v3high.Document
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

	fmt.Printf("PATH !!!!!!!!!!!!!!!!!!!!!! %v\n", path)
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
			ExtractRefsSequentially: true,
			Logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})),
			BundleInlineRefs: true,
			// SkipCircularReferenceCheck:          true,
			// IgnorePolymorphicCircularReferences: true,
			// IgnoreArrayCircularReferences:       true,
		})

	if err != nil {
		return nil, err
	}

	model, errs := doc.BuildV3Model()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return &openAPIDoc{
		doc:   doc,
		model: &model.Model,
	}, nil
}

func (o *openAPIDoc) expand() (map[string]interface{}, error) {
	b, err := bundler.BundleDocument(o.model)
	if err != nil {
		return nil, err
	}

	var newDoc map[string]interface{}
	err = yaml.Unmarshal(b, &newDoc)
	return newDoc, err
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
			totalErrors = append(totalErrors, valErrs[i].Error())
		}
	}
	return totalErrors
}
