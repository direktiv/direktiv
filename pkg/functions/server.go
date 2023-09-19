package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/dlog"
	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/direktiv/direktiv/pkg/version"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	logger2 "gorm.io/gorm/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const FunctionsChannel = "fnsync"

var (
	grpcServer      *grpc.Server
	empty           emptypb.Empty
	functionsConfig config

	logger *zap.SugaredLogger
)

const (
	confFile = "/etc/direktiv/functions-config.yaml"
	port     = 5555
)

type functionsServer struct {
	igrpc.UnimplementedFunctionsServer
	dbStore core.ServicesStore

	reusableCacheLock  sync.Mutex
	reusableCache      map[string]*cacheTuple
	reusableCacheIndex map[string]*cacheTuple
}

// StartServer starts functions grpc server.
func StartServer(echan chan error) {
	var err error

	logger, err = dlog.ApplicationLogger("functions")
	if err != nil {
		echan <- err
		return
	}

	// we read first in case the watcher is not working
	logger.Infof("loading config file %s", confFile)
	readConfig(confFile, &functionsConfig)

	err = initLocks(os.Getenv(util.DBConn))
	if err != nil {
		echan <- err
		return
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  os.Getenv(util.DBConn),
		PreferSimpleProtocol: false, // disables implicit prepared statement usage
		// Conn:                 edb.DB(),
	}), &gorm.Config{
		Logger: logger2.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger2.Config{
				LogLevel: logger2.Silent,
			},
		),
	})
	if err != nil {
		echan <- fmt.Errorf("creating services store, err: %w", err)
		return
	}

	gdb, err := gormDB.DB()
	if err != nil {
		echan <- fmt.Errorf("connecting via gorm driver, err: %w", err)
		return
	}
	gdb.SetMaxIdleConns(8)
	gdb.SetMaxOpenConns(8)

	fServer := functionsServer{
		dbStore:            datastoresql.NewServicesStore(gormDB),
		reusableCache:      make(map[string]*cacheTuple),
		reusableCacheIndex: make(map[string]*cacheTuple),
	}

	err = util.GrpcStart(&grpcServer, "functions",
		fmt.Sprintf(":%d", port), func(srv *grpc.Server) {
			igrpc.RegisterFunctionsServer(srv, &fServer)
			reflection.Register(srv)
		})
	if err != nil {
		echan <- err
	}

	go fServer.reusableGC()
	go fServer.orphansGC()

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			logger.Errorf("pubsub error: %v\n", err)
			os.Exit(1)
		}
	}

	listener := pq.NewListener(os.Getenv(util.DBConn), 10*time.Second,
		time.Minute, reportProblem)
	err = listener.Listen(FunctionsChannel)
	if err != nil {
		echan <- err
		return
	}

	go func(l *pq.Listener) {
		defer func() {
			err := l.UnlistenAll()
			logger.Errorf("Failed to deregister listeners: %v.", err)
		}()

		for {
			var more bool
			var notification *pq.Notification

			notification, more = <-l.Notify
			if !more {
				logger.Errorf("database listener closed\n")
				return
			}

			if notification == nil {
				continue
			}

			var tuples []*HeartbeatTuple

			err = json.Unmarshal([]byte(notification.Extra), &tuples)
			if err != nil {
				logger.Error(fmt.Sprintf("unexpected notification listener: %v", err))
				continue
			} else {
				go fServer.heartbeat(tuples)
			}
		}
	}(listener)

	err = fServer.reconstructServices(context.Background())
	if err != nil {
		echan <- err
	}

	select {}
}

// StopServer is stopping server gracefully.
func StopServer() {
	if grpcServer != nil {
		grpcServer.GracefulStop()
	}
}

type HeartbeatTuple struct {
	NamespaceName      string
	NamespaceID        string
	WorkflowPath       string
	Revision           string
	FunctionDefinition *model.ReusableFunctionDefinition
}

func (fServer *functionsServer) heartbeat(tuples []*HeartbeatTuple) {
	logger.Debugf("workflow functions heartbeat received.")

	ctx := context.Background()

	for _, tuple := range tuples {
		size := int32(tuple.FunctionDefinition.Size)
		minscale := int32(0)

		wf := bytedata.ShortChecksum(tuple.WorkflowPath)

		in := &igrpc.FunctionsCreateFunctionRequest{
			Info: &igrpc.FunctionsBaseInfo{
				Name:          &tuple.FunctionDefinition.ID,
				Namespace:     &tuple.NamespaceID,
				Workflow:      &wf,
				Image:         &tuple.FunctionDefinition.Image,
				Cmd:           &tuple.FunctionDefinition.Cmd,
				Size:          &size,
				MinScale:      &minscale,
				NamespaceName: &tuple.NamespaceName,
				Path:          &tuple.WorkflowPath,
				Revision:      &tuple.Revision,
			},
		}

		name, _, _ := GenerateServiceName(in.Info)

		wfID := name // NOTE: alan, we used to use a different value here to group services by workflow, but I can't figure out why we wanted to do that and I now believe that was causing bugs. I think this means there's no longer any sense keeping two different caches (cache + cache index), but I don't want to mess with what's working right now.

		logger.Debugf("checking service %s in heartbeat", name)

		fServer.reusableCacheLock.Lock()

		ct, exists := fServer.reusableCache[wfID]
		if exists {
			ct.Add(name)
		} else {
			ct = new(cacheTuple)
			ct.Add(name)
			fServer.reusableCache[wfID] = ct
		}
		fServer.reusableCacheIndex[name] = ct
		fServer.reusableCacheLock.Unlock()

		logger.Debugf("creating workflow function in heartbeat: %s", name)

		_, err := fServer.CreateFunction(ctx, in)
		if err != nil {
			if status.Code(err) != codes.AlreadyExists {
				logger.Errorf("Failed to create workflow function in heartbeat: %v", err)
				continue
			}
		}
	}
}

func (fServer *functionsServer) reusableGC() {
	ticker := time.NewTicker(time.Minute * 5)

	for {
		<-ticker.C

		logger.Debugf("reusable heartbeat garbage collector running.")

		cutoff := time.Now().UTC().Add(time.Minute * -15)

		fServer.reusableCacheLock.Lock()

		for k, tuple := range fServer.reusableCache {
			if tuple.t.Before(cutoff) {
				go fServer.reusableFree(k)
			}
		}

		fServer.reusableCacheLock.Unlock()
	}
}

type cacheTuple struct {
	t     time.Time
	names []string
}

func (ct *cacheTuple) Add(name string) {
	ct.t = time.Now().UTC()

	sort.Strings(ct.names)

	idx := sort.SearchStrings(ct.names, name)

	if idx < len(ct.names) && ct.names[idx] == name {
		return
	}

	ct.names = append(ct.names, name)
}

func (fServer *functionsServer) reusableFree(k string) {
	fServer.reusableCacheLock.Lock()

	x, exists := fServer.reusableCache[k]

	if exists {
		delete(fServer.reusableCache, k)
		for _, name := range x.names {
			delete(fServer.reusableCacheIndex, name)
		}
	}

	fServer.reusableCacheLock.Unlock()

	if !exists {
		return
	}

	ctx := context.Background()

	logger.Debugf("reusable heartbeat garbage collector purging workflow functions: %s", k)

	for i := range x.names {
		name := x.names[i]

		in := &igrpc.FunctionsGetFunctionRequest{
			ServiceName: &name,
		}

		logger.Debugf("reusable heartbeat garbage collector purging workflow function: %s", name)

		_, err := fServer.DeleteFunction(ctx, in)
		if err != nil {
			logger.Errorf("reusable heartbeat garbage collector failed to purge workflow function: %v", err)
			continue
		}
	}
}

func (fServer *functionsServer) orphansGC() {
	ticker := time.NewTicker(time.Minute * 2)

	for {
		<-ticker.C

		logger.Debugf("reusable orphans garbage collector running.")

		ctx := context.Background()

		filtered := map[string]string{
			"direktiv.io/scope": "workflow",
		}

		cs, err := fetchServiceAPI()
		if err != nil {
			err = fmt.Errorf("error getting clientset for knative: %w", err)
			logger.Errorf("reusable orphans garbage collector failed to list workflow functions: %v", err)
			continue
		}

		lo := metav1.ListOptions{LabelSelector: labels.Set(filtered).String()}
		l, err := cs.ServingV1().Services(functionsConfig.Namespace).List(context.Background(), lo)
		if err != nil {
			logger.Errorf("reusable orphans garbage collector failed to list workflow functions: %v", err)
			continue
		}

		for i := range l.Items {
			item := l.Items[i]

			fServer.reusableCacheLock.Lock()
			_, exists := fServer.reusableCacheIndex[item.Name]
			fServer.reusableCacheLock.Unlock()

			if !exists {
				if !item.CreationTimestamp.Time.Before(time.Now().UTC().Add(time.Minute * -60)) {
					continue
				}
				logger.Debugf("Reusable orphans garbage collector deleting detected orphan function: %s", item.Name)

				_, err := fServer.DeleteFunction(ctx, &igrpc.FunctionsGetFunctionRequest{
					ServiceName: &item.Name,
				})
				if err != nil {
					logger.Errorf("Reusable orphans garbage collector failed to purge orphaned function: %v", err)
					continue
				}
			}
		}
	}
}

func (is *functionsServer) Build(ctx context.Context, in *emptypb.Empty) (*igrpc.FunctionsBuildResponse, error) {
	var resp igrpc.FunctionsBuildResponse
	resp.Build = version.Version
	return &resp, nil
}
