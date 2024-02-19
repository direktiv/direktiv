package service

import (
	"fmt"
	"io"
	"slices"
	"sort"
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
	cfg *core.Config
	// this list maintains all the service configurations that need to be running.
	list []*core.ServiceFileExtra

	// the underlying service runtime used to create/schedule the services.
	runtimeClient runtimeClient

	logger *zap.SugaredLogger
	lock   *sync.Mutex

	servicesListHasBeenSet bool // NOTE: set to true the first time SetServices is called, and used to prevent any reconciles before that has happened.
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
			cfg:           c,
			list:          make([]*core.ServiceFileExtra, 0),
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

	client := &knativeClient{
		config:     c,
		knativeCli: knativeCli,
		k8sCli:     k8sCli,
	}

	return &manager{
		cfg:           c,
		list:          make([]*core.ServiceFileExtra, 0),
		runtimeClient: client,

		logger: logger,
		lock:   &sync.Mutex{},
	}, nil
}

func (m *manager) runCycle() []error {
	if !m.servicesListHasBeenSet {
		return nil
	}

	// clone the list
	src := make([]reconcileObject, len(m.list))
	for i, v := range m.list {
		src[i] = v
	}

	searchSrc := map[string]*core.ServiceFileExtra{}
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
			errs = append(errs, fmt.Errorf("create service id: %s %w", id, err))
			errStr := err.Error()
			v.Error = &errStr
		}
	}

	for _, id := range result.updates {
		v := searchSrc[id]
		v.Error = nil
		// v is passed un-cloned.
		if err := m.runtimeClient.updateService(v); err != nil {
			errs = append(errs, fmt.Errorf("update service id: %s %w", id, err))
			errStr := err.Error()
			v.Error = &errStr
		}
	}

	return errs
}

func (m *manager) Start(done <-chan struct{}, wg *sync.WaitGroup) {
	cycleTime := m.cfg.GetFunctionsReconcileInterval()

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

func (m *manager) SetServices(list []*core.ServiceFileExtra) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.servicesListHasBeenSet = true

	m.list = slices.Clone(list)
	for i := range m.list {
		cp := *m.list[i]
		m.setServiceDefaults(&cp)
		m.list[i] = &cp
	}
}

type serviceList []*core.ServiceStatus

func (x serviceList) Len() int {
	return len(x)
}

func (x serviceList) Less(i, j int) bool {
	return x[i].FilePath < x[j].FilePath
}

func (x serviceList) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (m *manager) getList(filterNamespace string, filterTyp string, filterPath string, filterName string) ([]*core.ServiceStatus, error) {
	// clone the list
	sList := make([]*core.ServiceFileExtra, 0)
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

		sList = append(sList, m.list[i])
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

	for _, v := range sList {
		service := searchServices[v.GetID()]
		// sometimes hashes might be different (not reconciled yet).
		if service != nil && service.GetValueHash() == v.GetValueHash() {
			result = append(result, &core.ServiceStatus{
				ID:               v.GetID(),
				ServiceFileExtra: *v,
				Conditions:       service.GetConditions(),
			})
		} else {
			result = append(result, &core.ServiceStatus{
				ID:               v.GetID(),
				ServiceFileExtra: *v,
				Conditions:       make([]any, 0),
			})
		}
	}

	sort.Sort(serviceList(result))

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

func (m *manager) Rebuild(namespace string, serviceID string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// check if serviceID exists.
	_, err := m.getOne(namespace, serviceID)
	if err != nil {
		return err
	}

	return m.runtimeClient.rebuildService(serviceID)
}

func (m *manager) setServiceDefaults(sv *core.ServiceFileExtra) {
	// empty size string defaults to medium
	if sv.Size == "" {
		m.logger.Warnw("empty service size, defaulting to medium", "service_file", sv.FilePath)
		sv.Size = "medium"
	}
	if sv.Scale > m.cfg.KnativeMaxScale {
		m.logger.Warnw("service_scale is bigger than allowed max_scale, defaulting to max_scale",
			"service_scale", sv.Scale,
			"max_scale", m.cfg.KnativeMaxScale,
			"service_file", sv.FilePath)
		sv.Scale = m.cfg.KnativeMaxScale
	}
	if len(sv.Envs) == 0 {
		sv.Envs = make([]core.EnvironmentVariable, 0)
	}
}
