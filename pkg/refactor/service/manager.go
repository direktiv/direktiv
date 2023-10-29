package service

import (
	"fmt"
	"io"
	"slices"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	dClient "github.com/docker/docker/client"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

// manager struct implements core.ServiceManager by wrapping runtimeClient. manager implementation manages
// services in the system in a declarative manner. This implementation spans up a goroutine (via Start())
// to reconcile the services in list param versus what is running in the runtime.
type manager struct {
	// this list maintains all the service configurations that need to be running.
	list []*core.ServiceConfig

	// the underlying service runtime used to create/schedule the services.
	runtimeClient runtimeClient

	logger *zap.SugaredLogger
	lock   *sync.Mutex
}

func NewManager(c *core.Config, logger *zap.SugaredLogger, enableDocker bool) (core.ServiceManager, error) {
	logger = logger.With("module", "service manager")
	if enableDocker {
		cli, err := dClient.NewClientWithOpts(dClient.FromEnv, dClient.WithAPIVersionNegotiation())
		if err != nil {
			return nil, fmt.Errorf("creating docker client: %w", err)
		}

		client := &dockerClient{
			cli: cli,
		}
		err = client.cleanAll()
		if err != nil {
			return nil, fmt.Errorf("cleaning docker client: %w", err)
		}

		return &manager{
			list:          make([]*core.ServiceConfig, 0),
			runtimeClient: client,

			logger: logger,
			lock:   &sync.Mutex{},
		}, nil
	}

	return newKnativeManager(c, logger)
}

func newKnativeManager(c *core.Config, logger *zap.SugaredLogger) (*manager, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	knativeCli, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k8sCli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// TODO: remove dev code.
	c.KnativeServiceAccount = "direktiv-functions-pod"
	c.KnativeNamespace = "direktiv-services-direktiv"
	c.KnativeIngressClass = "contour.ingress.networking.knative.dev"
	c.KnativeMaxScale = 5
	c.KnativeSidecar = "localhost:5000/direktiv"

	client := &knativeClient{
		config:     c,
		knativeCli: knativeCli,
		k8sCli:     k8sCli,
	}

	return &manager{
		list:          make([]*core.ServiceConfig, 0),
		runtimeClient: client,

		logger: logger,
		lock:   &sync.Mutex{},
	}, nil
}

func (m *manager) runCycle() []error {
	// clone the list
	src := make([]reconcileObject, len(m.list))
	for i, v := range m.list {
		src[i] = v
	}

	searchSrc := map[string]*core.ServiceConfig{}
	for _, v := range m.list {
		searchSrc[v.GetID()] = v
	}

	knList, err := m.runtimeClient.listServices()
	if err != nil {
		return []error{err}
	}

	target := make([]reconcileObject, len(knList))
	for i, v := range knList {
		target[i] = v
	}

	m.logger.Debugw("reconcile length", "src", len(src), "target", len(target))

	result := reconcile(src, target)

	errs := []error{}
	for _, id := range result.deletes {
		if err := m.runtimeClient.deleteService(id); err != nil {
			errs = append(errs, fmt.Errorf("delete service id: %s %w", id, err))
		}
	}

	for _, id := range result.creates {
		v := searchSrc[id]
		v.Error = nil
		// v is passed un-cloned.
		if err := m.runtimeClient.createService(v); err != nil {
			errStr := err.Error()
			v.Error = &errStr
		}
	}

	for _, id := range result.updates {
		v := searchSrc[id]
		v.Error = nil
		// v is passed un-cloned.
		if err := m.runtimeClient.updateService(v); err != nil {
			errStr := err.Error()
			v.Error = &errStr
		}
	}

	return errs
}

func (m *manager) Start(done <-chan struct{}, wg *sync.WaitGroup) {
	const cycleTime = 2 * time.Second

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
				m.logger.Errorw("run cycle", "error", err)
			}
			time.Sleep(cycleTime)
		}

		wg.Done()
	}()
}

func (m *manager) SetServices(list []*core.ServiceConfig) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.list = slices.Clone(list)
	for i := range m.list {
		cp := *m.list[i]
		cp.SetDefaults()
		m.list[i] = &cp
	}
}

func (m *manager) getList(filterNamespace string, filterTyp string, filterPath string, filterName string) ([]*core.ServiceStatus, error) {
	// clone the list
	cfgList := make([]*core.ServiceConfig, 0)
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
		if filterName != "" && v.Name != filterName {
			continue
		}

		cfgList = append(cfgList, m.list[i])
	}

	services, err := m.runtimeClient.listServices()
	if err != nil {
		return nil, err
	}
	searchServices := map[string]status{}
	for _, v := range services {
		searchServices[v.GetID()] = v
	}

	result := []*core.ServiceStatus{}

	for _, v := range cfgList {
		service := searchServices[v.GetID()]
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

// nolint:unparam
func (m *manager) getOne(namespace string, serviceID string) (*core.ServiceStatus, error) {
	list, err := m.getList(namespace, "", "", "")
	if err != nil {
		return nil, err
	}
	for _, svc := range list {
		if svc.ID == serviceID {
			cp := *svc

			return &cp, nil
		}
	}

	return nil, core.ErrNotFound
}

func (m *manager) GeAll(namespace string) ([]*core.ServiceStatus, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.getList(namespace, "", "", "")
}

func (m *manager) GetPods(namespace string, serviceID string) (any, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// check if serviceID exists.
	_, err := m.getOne(namespace, serviceID)
	if err != nil {
		return nil, err
	}

	pods, err := m.runtimeClient.listServicePods(serviceID)
	if err != nil {
		return nil, err
	}

	return pods, nil
}

func (m *manager) StreamLogs(namespace string, serviceID string, podID string) (io.ReadCloser, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// check if serviceID exists.
	_, err := m.getOne(namespace, serviceID)
	if err != nil {
		return nil, err
	}

	return m.runtimeClient.streamServiceLogs(serviceID, podID)
}

func (m *manager) Kill(namespace string, serviceID string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// check if serviceID exists.
	_, err := m.getOne(namespace, serviceID)
	if err != nil {
		return err
	}

	return m.runtimeClient.killService(serviceID)
}
