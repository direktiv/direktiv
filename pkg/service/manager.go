package service

import (
	"fmt"
	"io"
	"log/slog"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/service/reconcile"
	"github.com/direktiv/direktiv/pkg/telemetry"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

// manager struct implements core.ServiceManager by wrapping runtimeClient. manager implementation manages
// services in the system in a declarative manner. This implementation spans up a goroutine (via Start())
// to Run the services in list param versus what is running in the runtime.
type manager struct {
	cfg *core.Config
	// this list maintains all the service configurations that need to be running.
	list []*core.ServiceFileData

	// the underlying service runtime used to create/schedule the services.
	runtimeClient runtimeClient

	lock *sync.Mutex

	servicesListHasBeenSet bool // NOTE: set to true the first time SetServices is called, and used to prevent any reconciles before that has happened.
}

func NewManager(c *core.Config) (core.ServiceManager, error) {
	return newKnativeManager(c)
}

func newKnativeManager(c *core.Config) (*manager, error) {
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
		list:          make([]*core.ServiceFileData, 0),
		runtimeClient: client,

		lock: &sync.Mutex{},
	}, nil
}

func (m *manager) runCycle() []error {
	if !m.servicesListHasBeenSet {
		return nil
	}
	// clone the list
	src := make([]reconcile.Item, len(m.list))
	for i, v := range m.list {
		src[i] = v
	}

	searchSrc := map[string]*core.ServiceFileData{}
	for _, v := range m.list {
		searchSrc[v.GetID()] = v
	}

	knList, err := m.runtimeClient.listServices()
	if err != nil {
		return []error{err}
	}

	target := make([]reconcile.Item, len(knList))
	for i, v := range knList {
		target[i] = v
	}

	result := reconcile.Calculate(src, target)

	errs := []error{}
	for _, id := range result.Deletes {
		slog.Debug(fmt.Sprintf("deleting service with id %s", id))
		if err := m.runtimeClient.deleteService(id); err != nil {
			slog.Error(fmt.Sprintf("failed to delete service with id %s", id), "error", err)
			errs = append(errs, fmt.Errorf("delete service id: %s %w", id, err))
		}
	}

	for _, id := range result.Creates {
		v := searchSrc[id]
		v.Error = nil
		// v is passed un-cloned.
		slog.Debug(fmt.Sprintf("creating service %s with id %s, type %s", v.FilePath, v.ID, v.Typ))
		if err := m.runtimeClient.createService(v); err != nil {
			slog.Error(fmt.Sprintf("failed to create service %s with id %s, type %s", v.FilePath, v.ID, v.Typ), "error", err)
			errs = append(errs, fmt.Errorf("create service id: %s %w", id, err))
			errStr := err.Error()
			v.Error = &errStr
		}
	}

	for _, id := range result.Updates {
		v := searchSrc[id]
		v.Error = nil
		// v is passed un-cloned.
		slog.Debug(fmt.Sprintf("updating service %s with ID %s, type %s", v.FilePath, v.ID, v.Typ))
		if err := m.runtimeClient.updateService(v); err != nil {
			slog.Error(fmt.Sprintf("failed to update service %s with ID %s, type %s", v.FilePath, v.ID, v.Typ), "error", err)
			errs = append(errs, fmt.Errorf("update service id: %s %w", id, err))
			errStr := err.Error()
			v.Error = &errStr
		}
	}

	return errs
}

func (m *manager) Run(circuit *core.Circuit) error {
	cycleTime := m.cfg.GetFunctionsReconcileInterval()
	for {
		if circuit.IsDone() {
			return nil
		}
		m.lock.Lock()
		errs := m.runCycle()
		m.lock.Unlock()
		for _, err := range errs {
			slog.Error("run cycle", "err", err)
		}

		time.Sleep(cycleTime)
	}
}

func (m *manager) SetServices(list []*core.ServiceFileData) {
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

func (m *manager) getList(filterNamespace string, filterTyp string, filterPath string, filterName string) ([]*core.ServiceFileData, error) {
	// clone the list
	sList := make([]*core.ServiceFileData, 0)
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

	// Populate id and conditions fields.
	for _, v := range sList {
		service := searchServices[v.GetID()]
		// sometimes hashes might be different (not reconciled yet).
		if service != nil && service.GetValueHash() == v.GetValueHash() {
			v.ID = v.GetID()
			v.Conditions = service.GetConditions()
		} else {
			v.ID = v.GetID()
			v.Conditions = make([]any, 0)
		}
	}

	sort.Slice(sList, func(i, j int) bool {
		return sList[i].FilePath < sList[j].FilePath
	})

	return sList, nil
}

// nolint:unparam
func (m *manager) getOne(namespace string, serviceID string) (*core.ServiceFileData, error) {
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

func (m *manager) GeAll(namespace string) ([]*core.ServiceFileData, error) {
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

func (m *manager) setServiceDefaults(sv *core.ServiceFileData) {
	// empty size string defaults to medium
	if sv.Size == "" {
		telemetry.LogNamespace(telemetry.LogLevelWarn, sv.Namespace,
			fmt.Sprintf("empty service size for %s, defaulting to medium", sv.FilePath))
		sv.Size = "medium"
	}
	if sv.Scale > m.cfg.KnativeMaxScale {
		telemetry.LogNamespace(telemetry.LogLevelWarn, sv.Namespace,
			fmt.Sprintf("service_scale for %s is bigger than allowed max_scale, defaulting to max_scale %s",
				sv.FilePath, m.cfg.KnativeMaxScale))
		sv.Scale = m.cfg.KnativeMaxScale
	}
	if len(sv.Envs) == 0 {
		sv.Envs = make([]core.EnvironmentVariable, 0)
	}
}
