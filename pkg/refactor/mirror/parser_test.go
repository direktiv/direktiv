package mirror_test

import (
	"fmt"
	"io/fs"
	"sort"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/psanford/memfs"
	"go.uber.org/zap"
)

type memSource struct {
	fs *memfs.FS
}

var _ mirror.Source = &mirror.DirectorySource{}

func newMemSource() *memSource {
	return &memSource{
		fs: memfs.New(),
	}
}

func (src *memSource) FS() fs.FS {
	return src.fs
}

func (src *memSource) Free() error {
	return nil
}

func (src *memSource) Notes() map[string]string {
	return make(map[string]string)
}

func assertTree(t *testing.T, p *mirror.Parser, paths []string) {
	expect := fmt.Sprintf("%v", paths)

	paths, err := p.ListFiles()
	if err != nil {
		t.Error(err)
		return
	}

	actual := fmt.Sprintf("%v", paths)

	if expect != actual {
		t.Errorf("assertTree failed: expected %s but got %s", expect, actual)
		t.Fail()
	}
}

func assertFilters(t *testing.T, p *mirror.Parser, filters []string) {
	expect := fmt.Sprintf("%v", filters)

	var x []string

	for k := range p.Filters {
		x = append(x, k)
	}

	sort.Strings(x)

	actual := fmt.Sprintf("%v", x)

	if expect != actual {
		t.Errorf("assertFilters failed: expected %s but got %s", expect, actual)
		t.Fail()
	}
}

func assertWorkflows(t *testing.T, p *mirror.Parser, paths []string) {
	expect := fmt.Sprintf("%v", paths)

	var x []string

	for k := range p.Workflows {
		x = append(x, k)
	}

	sort.Strings(x)

	actual := fmt.Sprintf("%v", x)

	if expect != actual {
		t.Errorf("assertWorkflows failed: expected %s but got %s", expect, actual)
		t.Fail()
	}
}

func TestParseSimple(t *testing.T) {
	log := zap.NewNop().Sugar()
	src := newMemSource()
	_ = src.fs.WriteFile(".direktivignore", []byte(``), 0o755)
	_ = src.fs.WriteFile("x.yaml", []byte(`x: 5`), 0o755)
	_ = src.fs.WriteFile("y.json", []byte(`{}`), 0o755)
	_ = src.fs.MkdirAll("a/b", 0o755)
	_ = src.fs.WriteFile("a/b/c.yaml", []byte(`direktiv_api: workflow/v1
states:
- id: a
  type: noop
`), 0o755)
	_ = src.fs.WriteFile("a/b/d.yaml", []byte(`
states:
- id: a
	type: noop
`), 0o755)
	_ = src.fs.WriteFile("a/b/e.yaml", []byte(`
states:
- id: a
  type: noop
`), 0o755)

	p, err := mirror.NewParser(log, src)
	if err != nil {
		t.Error(err)
		return
	}

	assertTree(t, p, []string{
		"a",
		"a/b",
		"a/b/d.yaml",
		"x.yaml",
		"y.json",
	})

	assertWorkflows(t, p, []string{
		"a/b/c.yaml",
		"a/b/e.yaml",
	})
}

func TestParseComplex(t *testing.T) {
	log := zap.NewNop().Sugar()
	src := newMemSource()
	_ = src.fs.WriteFile(".direktivignore", []byte(`*.csv
a/b/e.yaml
a/f/`), 0o755)
	_ = src.fs.WriteFile("filters.yaml", []byte(`direktiv_api: filters/v1
filters:
- name: alpha
`), 0o755)
	_ = src.fs.WriteFile("x.yaml", []byte(`x: 5`), 0o755)
	_ = src.fs.WriteFile("y.json", []byte(`{}`), 0o755)
	_ = src.fs.WriteFile("z.csv", []byte(`1,2,3`), 0o755)
	_ = src.fs.MkdirAll("a/b", 0o755)
	_ = src.fs.WriteFile("a/b/c.yaml", []byte(`direktiv_api: workflow/v1
states:
- id: a
  type: noop
`), 0o755)
	_ = src.fs.WriteFile("a/b/d.yaml", []byte(`
states:
- id: a
	type: noop
`), 0o755)
	_ = src.fs.WriteFile("a/b/e.yaml", []byte(`
states:
- id: a
  type: noop
`), 0o755)
	_ = src.fs.WriteFile("a/b/z.csv", []byte(`1,2,3`), 0o755)
	_ = src.fs.MkdirAll("a/b", 0o755)
	_ = src.fs.WriteFile("a/f/c.yaml", []byte(`direktiv_api: workflow/v1
states:
- id: a
  type: noop
`), 0o755)
	_ = src.fs.WriteFile("a/f/d.yaml", []byte(`
states:
- id: a
	type: noop
`), 0o755)
	_ = src.fs.WriteFile("a/f/e.yaml", []byte(`
states:
- id: a
  type: noop
`), 0o755)

	p, err := mirror.NewParser(log, src)
	if err != nil {
		t.Error(err)
		return
	}

	assertTree(t, p, []string{
		"a",
		"a/b",
		"a/b/d.yaml",
		"x.yaml",
		"y.json",
	})

	assertFilters(t, p, []string{
		"alpha",
	})

	assertWorkflows(t, p, []string{
		"a/b/c.yaml",
	})
}
