package mirror_test

import (
	"fmt"
	"io/fs"
	"sort"
	"testing"

	"github.com/direktiv/direktiv/pkg/mirror"
	"github.com/psanford/memfs"
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
