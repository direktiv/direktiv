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
		client: cset,
		config: c,
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
		searchSrc[v.getId()] = v
	}

	m.lock.Unlock()

	knList, err := m.client.listServices()
	if err != nil {
		return []error{err}
	}

	//fmt.Printf("klist2:")
	//for _, v := range knList {
	//	fmt.Printf(" {%v %v} ", v.getId(), v.getValueHash())
	//}
	//fmt.Printf("\n")

	target := make([]reconcileObject, len(knList))
	for i, v := range knList {
		target[i] = v
	}

	fmt.Printf("f2: lens: %v %v\n", len(src), len(target))

	//fmt.Printf("srcs:")
	//for _, v := range src {
	//	fmt.Printf(" {%v %v} ", v.getId(), v.getValueHash())
	//}
	//fmt.Printf("\n")
	//
	//fmt.Printf("tars:")
	//for _, v := range target {
	//	fmt.Printf(" {%v %v} ", v.getId(), v.getValueHash())
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
		v.Error = ""
		if err := m.client.createService(v); err != nil {
			v.Error = err.Error()
		}
	}

	for _, id := range result.updates {
		v := searchSrc[id]
		v.Error = ""
		if err := m.client.updateService(v); err != nil {
			v.Error = err.Error()
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
			errs := m.runCycle()
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

	m.list = make([]*Config, len(list))

	copy(m.list, list)
}

func (m *Manager) SetOneService(service *Config) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for i, v := range m.list {
		if v.getId() == service.getId() {
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

	services, err := m.client.listServices()
	if err != nil {
		return nil, err
	}
	searchServices := map[string]Status{}
	for _, v := range services {
		searchServices[v.getId()] = v
	}

	result := []ConfigStatus{}

	for _, v := range cfgList {
		service, _ := searchServices[v.getId()]
		// sometimes hashes might be different (not reconciled yet).
		if service != nil && service.getValueHash() == v.getValueHash() {
			result = append(result, ConfigStatus{
				Config:     v,
				Conditions: service.getConditions(),
			})
		} else {
			result = append(result, ConfigStatus{
				Config:     v,
				Conditions: nil,
			})
		}
	}

	return result, nil
}
