// nolint
package service

import (
	"fmt"
	"io"
	"slices"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	dClient "github.com/docker/docker/client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

type Manager struct {
	list   []*core.ServiceConfig
	client client

	lock *sync.Mutex
}

func NewManager(c *core.Config, enableDocker bool) (*Manager, error) {
	if enableDocker {
		return newDockerManager()
	}

	return newKnativeManager(c)
}

func newKnativeManager(c *core.Config) (*Manager, error) {
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

	k8sSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error cluster config: %v\n", err)
		return nil, err
	}

	// TODO: remove dev code.
	c.KnativeServiceAccount = "direktiv-functions-pod"
	c.KnativeNamespace = "direktiv-services-direktiv"
	c.KnativeIngressClass = "contour.ingress.networking.knative.dev"
	c.KnativeMaxScale = 5

	client := &knativeClient{
		client: cSet,
		config: c,
		k8sSet: k8sSet,
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
		list:   make([]*core.ServiceConfig, 0, 0),
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
	searchSrc := map[string]*core.ServiceConfig{}
	for _, v := range m.list {
		searchSrc[v.GetID()] = v
	}

	knList, err := m.client.listServices()
	if err != nil {
		return []error{err}
	}

	//fmt.Printf("klist2:")
	//for _, v := range knList {
	//	fmt.Printf(" {%v %v} ", v.GetID(), v.GetValueHash())
	//}
	//fmt.Printf("\n")

	target := make([]reconcileObject, len(knList))
	for i, v := range knList {
		target[i] = v
	}

	fmt.Printf("f2: lens: %v %v\n", len(src), len(target))

	//fmt.Printf("srcs:")
	//for _, v := range src {
	//	fmt.Printf(" {%v %v} ", v.GetID(), v.GetValueHash())
	//}
	//fmt.Printf("\n")
	//
	//fmt.Printf("tars:")
	//for _, v := range target {
	//	fmt.Printf(" {%v %v} ", v.GetID(), v.GetValueHash())
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
			time.Sleep(2 * time.Second)
		}

		wg.Done()
	}()
}

func (m *Manager) SetServices(list []*core.ServiceConfig) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.list = slices.Clone(list)
	for i, _ := range m.list {
		cp := *m.list[i]
		m.list[i] = &cp
	}
}

func (m *Manager) getList(filterNamespace string, filterTyp string, filterPath string) ([]*core.ServiceStatus, error) {
	// clone the list
	cfgList := make([]*core.ServiceConfig, len(m.list))
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
	searchServices := map[string]status{}
	for _, v := range services {
		searchServices[v.GetID()] = v
	}

	result := []*core.ServiceStatus{}

	for _, v := range cfgList {
		service, _ := searchServices[v.GetID()]
		// sometimes hashes might be different (not reconciled yet).
		if service != nil && service.GetValueHash() == v.GetValueHash() {
			result = append(result, &core.ServiceStatus{
				ID:            v.GetID(),
				ServiceConfig: *v,
				Conditions:    service.GetConditions(),
			})
		} else {
			result = append(result, &core.ServiceStatus{
				ID:            v.GetID(),
				ServiceConfig: *v,
				Conditions:    nil,
			})
		}
	}

	return result, nil
}

func (m *Manager) getOne(namespace string, serviceID string) (*core.ServiceStatus, error) {
	list, err := m.getList(namespace, "", "")
	if err != nil {
		return nil, err
	}
	for _, svc := range list {
		if svc.ID == serviceID {
			return svc, nil
		}
	}

	return nil, core.ErrNotFound
}

func (m *Manager) GeAll(namespace string) ([]*core.ServiceStatus, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.getList(namespace, "", "")
}

func (m *Manager) GetPods(namespace string, serviceID string) (any, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	list, err := m.getList(namespace, "", "")
	if err != nil {
		return nil, err
	}

	for _, svc := range list {
		if svc.ID == serviceID {
			pods, err := m.client.listServicePods(serviceID)
			if err != nil {
				return nil, err
			}
			return pods, nil
		}
	}

	return nil, core.ErrNotFound
}

func (m *Manager) StreamLogs(namespace string, serviceID string, podID string) (io.ReadCloser, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// check if serviceID exists.
	_, err := m.getOne(namespace, serviceID)
	if err != nil {
		return nil, err
	}

	return m.client.streamServiceLogs(serviceID, podID)
}

func (m *Manager) Kill(namespace string, serviceID string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// check if serviceID exists.
	_, err := m.getOne(namespace, serviceID)
	if err != nil {
		return err
	}

	return m.client.killService(serviceID)
}
