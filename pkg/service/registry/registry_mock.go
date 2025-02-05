package registry

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"sync"

	"github.com/direktiv/direktiv/pkg/core"
)

type mockedManager struct {
	lock *sync.Mutex
	list map[string][]*core.Registry
}

func (c *mockedManager) ListRegistries(namespace string) ([]*core.Registry, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	list, ok := c.list[namespace]
	if !ok {
		return []*core.Registry{}, nil
	}

	l := slices.Clone(list)
	for i := range l {
		n := &core.Registry{}
		*n = *l[i]
		l[i] = n
	}

	return l, nil
}

func (c *mockedManager) DeleteRegistry(namespace string, id string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	list, ok := c.list[namespace]
	if !ok {
		return core.ErrNotFound
	}

	cp := []*core.Registry{}
	for i := range list {
		if list[i].ID != id {
			cp = append(cp, list[i])
		}
	}

	if len(list) == len(cp) {
		return core.ErrNotFound
	}
	c.list[namespace] = cp

	return nil
}

func (c *mockedManager) DeleteNamespace(namespace string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, ok := c.list[namespace]
	if !ok {
		return core.ErrNotFound
	}
	delete(c.list, namespace)

	return nil
}

func (c *mockedManager) StoreRegistry(registry *core.Registry) (*core.Registry, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	str := fmt.Sprintf("%s-%s", registry.Namespace, registry.URL)
	sh := sha256.Sum256([]byte(str))
	id := fmt.Sprintf("secret-%x", sh[:10])

	registry.ID = id
	registry.Password = ""

	cp := *registry

	list, ok := c.list[registry.Namespace]
	if !ok {
		list = make([]*core.Registry, 0)
	}
	list = append(list, &cp)
	c.list[registry.Namespace] = list

	return registry, nil
}

func (c *mockedManager) TestLogin(registry *core.Registry) error {
	return testLogin(registry)
}

var _ core.RegistryManager = &mockedManager{}
