// nolint
package service

import (
	"fmt"
	"io"
	"sync"
	"time"

	dClient "github.com/docker/docker/client"
	"k8s.io/client-go/rest"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

type Manager struct {
	list   []*Config
	client client

	lock *sync.Mutex
}

func NewManager(enableDocker bool) (*Manager, error) {
	if enableDocker {
		return newDockerManager()
	}

	return newKnativeManager()
}

func newKnativeManager() (*Manager, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("error cluster config: %v\n", err)
		return nil, err
	}
	cSet, err := versioned.NewForConfig(config)
	if err != nil {
		fmt.Printf("error cluster config: %v\n", err)
		return nil, err
	}

	c := &ClientConfig{
		ServiceAccount: "direktiv-functions-pod",
		Namespace:      "direktiv-services-direktiv",
		IngressClass:   "contour.ingress.networking.knative.dev",
	}

	c, err = validateConfig(c)
	if err != nil {
		return nil, fmt.Errorf("invalid client config, err: %s", err)
	}

	client := &knativeClient{
		client: cSet,
		config: c,
	}

	return newManagerFromClient(client), nil
}

func newDockerManager() (*Manager, error) {
	cli, err := dClient.NewClientWithOpts(dClient.FromEnv, dClient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	client := dockerClient{
		cli: cli,
	}
	return newManagerFromClient(&client), nil
}

func newManagerFromClient(client client) *Manager {
	return &Manager{
		list:   make([]*Config, 0, 0),
		client: client,

		lock: &sync.Mutex{},
	}
}

func (m *Manager) runCycle() []error {
	// clone the list
	src := make([]reconcileObject, len(m.list))
	for i, v := range m.list {
		src[i] = v
	}
	searchSrc := map[string]*Config{}
	for _, v := range m.list {
		searchSrc[v.getID()] = v
	}

	knList, err := m.client.listServices()
	if err != nil {
		return []error{err}
	}

	//fmt.Printf("klist2:")
	//for _, v := range knList {
	//	fmt.Printf(" {%v %v} ", v.getID(), v.getValueHash())
	//}
	//fmt.Printf("\n")

	target := make([]reconcileObject, len(knList))
	for i, v := range knList {
		target[i] = v
	}

	fmt.Printf("f2: lens: %v %v\n", len(src), len(target))

	//fmt.Printf("srcs:")
	//for _, v := range src {
	//	fmt.Printf(" {%v %v} ", v.getID(), v.getValueHash())
	//}
	//fmt.Printf("\n")
	//
	//fmt.Printf("tars:")
	//for _, v := range target {
	//	fmt.Printf(" {%v %v} ", v.getID(), v.getValueHash())
	//}
	//fmt.Printf("\n")

	result := reconcile(src, target)

	// fmt.Printf("f2: recocile: %v\n", result)

	errs := []error{}

	for _, id := range result.deletes {
		if err := m.client.deleteService(id); err != nil {
			errs = append(errs, err)
		}
	}

	for _, id := range result.creates {
		v := searchSrc[id]
		v.Error = nil
		if err := m.client.createService(v); err != nil {
			errStr := err.Error()
			v.Error = &errStr
		}
	}

	for _, id := range result.updates {
		v := searchSrc[id]
		v.Error = nil
		if err := m.client.updateService(v); err != nil {
			*v.Error = err.Error()
		}
	}

	return errs
}

func (m *Manager) Start(done <-chan struct{}, wg *sync.WaitGroup) {
	go func() {
	loop:
		for {
			select {
			case <-done:
				break loop
			default:
			}
			m.lock.Lock()
			errs := m.runCycle()
			m.lock.Unlock()
			for _, err := range errs {
				fmt.Printf("f2 error: %s\n", err)
			}
			time.Sleep(10 * time.Second)
		}

		wg.Done()
	}()
}

func (m *Manager) SetServices(list []*Config) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.list = list
}

func (m *Manager) getList(filterNamespace string, filterTyp string, filterPath string) ([]*ConfigStatus, error) {
	// clone the list
	cfgList := make([]*Config, len(m.list))
	for i, v := range m.list {
		if filterNamespace != "" && filterNamespace != v.Namespace {
			continue
		}
		if filterPath != "" && v.FilePath != filterPath {
			continue
		}
		if filterTyp != "" && v.Typ != filterTyp {
			continue
		}

		cfgList[i] = v
	}

	services, err := m.client.listServices()
	if err != nil {
		return nil, err
	}
	searchServices := map[string]Status{}
	for _, v := range services {
		searchServices[v.getID()] = v
	}

	result := []*ConfigStatus{}

	for _, v := range cfgList {
		service, _ := searchServices[v.getID()]
		// sometimes hashes might be different (not reconciled yet).
		if service != nil && service.getValueHash() == v.getValueHash() {
			result = append(result, &ConfigStatus{
				ID:           v.getID(),
				Config:       *v,
				Conditions:   service.getConditions(),
				CurrentScale: service.getCurrentScale(),
			})
		} else {
			result = append(result, &ConfigStatus{
				ID:         v.getID(),
				Config:     *v,
				Conditions: nil,
			})
		}
	}

	return result, nil
}

func (m *Manager) GetListByNamespace(namespace string) ([]*ConfigStatus, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.getList(namespace, "", "")
}

func (m *Manager) StreamLogs(namespace string, serviceID string, podNumber int) (io.ReadCloser, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	list, err := m.getList(namespace, "", "")
	if err != nil {
		return nil, err
	}

	for _, svc := range list {
		if svc.ID == serviceID {
			return m.client.streamServiceLogs(serviceID, 1)
		}
	}

	return nil, ErrNotFound
}
