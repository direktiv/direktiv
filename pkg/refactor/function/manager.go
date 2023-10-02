// nolint
package function

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/client-go/rest"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

type Manager struct {
	list   []*Config
	client client

	lock *sync.Mutex
}

func NewManagerFromK8s() (*Manager, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("error cluster config: %v\n", err)
		return nil, err
	}
	cset, err := versioned.NewForConfig(config)
	if err != nil {
		fmt.Printf("error cluster config: %v\n", err)
		return nil, err
	}

	client := &knClient{
		client: cset,
		config: &ClientConfig{
			ServiceAccount: "direktiv-functions-pod",
			Namespace:      "direktiv-services-direktiv",
		},
	}

	return &Manager{
		list:   make([]*Config, 0, 0),
		client: client,

		lock: &sync.Mutex{},
	}, nil
}

func (m *Manager) runCycle() []error {
	m.lock.Lock()
	// clone the list
	src := make([]reconcileObject, len(m.list))
	for i, v := range m.list {
		src[i] = v
	}
	searchSrc := map[string]*Config{}
	for _, v := range m.list {
		searchSrc[v.id()] = v
	}

	m.lock.Unlock()

	knList, err := m.client.listServices()
	if err != nil {
		return []error{err}
	}

	//fmt.Printf("klist2:")
	//for _, v := range knList {
	//	fmt.Printf(" {%v %v} ", v.id(), v.hash())
	//}
	//fmt.Printf("\n")

	target := make([]reconcileObject, len(knList))
	for i, v := range knList {
		target[i] = v
	}

	fmt.Printf("f2: lens: %v %v\n", len(src), len(target))

	//fmt.Printf("srcs:")
	//for _, v := range src {
	//	fmt.Printf(" {%v %v} ", v.id(), v.hash())
	//}
	//fmt.Printf("\n")
	//
	//fmt.Printf("tars:")
	//for _, v := range target {
	//	fmt.Printf(" {%v %v} ", v.id(), v.hash())
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
		if err := m.client.createService(v); err != nil {
			errs = append(errs, err)
		}
	}

	for _, id := range result.updates {
		v := searchSrc[id]
		if err := m.client.createService(v); err != nil {
			errs = append(errs, err)
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
			time.Sleep(10 * time.Second)
			errs := m.runCycle()
			for _, err := range errs {
				fmt.Printf("f2 error: %s\n", err)
			}
		}

		wg.Done()
	}()
}

func (m *Manager) SetServices(list []*Config) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.list = make([]*Config, len(list))

	copy(m.list, list)
}

func (m *Manager) SetOneService(service *Config) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for i, v := range m.list {
		if v.id() == service.id() {
			m.list[i] = service

			return
		}
	}

	m.list = append(m.list, service)
}

func (m *Manager) GetList() ([]ConfigStatus, error) {
	m.lock.Lock()
	// clone the list
	cfgList := make([]*Config, len(m.list))
	for i, v := range m.list {
		cfgList[i] = v
	}
	m.lock.Unlock()

	statusList, err := m.client.listServices()
	if err != nil {
		return nil, err
	}
	searchStatus := map[string]Status{}
	for _, v := range statusList {
		searchStatus[v.id()] = v
	}

	result := []ConfigStatus{}

	for _, v := range cfgList {
		status, _ := searchStatus[v.id()]
		result = append(result, ConfigStatus{
			Config: v,
			Status: status,
		})
	}

	return result, nil
}
