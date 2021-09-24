package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
	"github.com/vorteil/direktiv/pkg/functions/ent"
	"github.com/vorteil/direktiv/pkg/model"

	"github.com/vorteil/direktiv/pkg/dlog"
	igrpc "github.com/vorteil/direktiv/pkg/functions/grpc"
	"github.com/vorteil/direktiv/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
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
	igrpc.UnimplementedFunctionsServiceServer
	db *ent.Client

	reusableCacheLock  sync.Mutex
	reusableCache      map[string]*cacheTuple
	reusableCacheIndex map[string]*cacheTuple
}

// StartServer starts functions grpc server
func StartServer(echan chan error) {

	var err error

	logger, err = dlog.ApplicationLogger("functions")
	if err != nil {
		echan <- err
		return
	}

	go runPodRequestLimiter()

	err = initKubernetesLock()
	if err != nil {
		echan <- err
		return
	}

	cr := newConfigReader()

	logger.Infof("loading config file %s", confFile)
	cr.readConfig(confFile, &functionsConfig)

	// Setup database
	db, err := ent.Open("postgres", os.Getenv(util.DBConn))
	if err != nil {
		logger.Errorf("failed to connect database client: %w", err)
		echan <- fmt.Errorf("failed to connect database client: %w", err)
	}

	// Run the auto migration tool.
	if err := db.Schema.Create(context.Background()); err != nil {
		logger.Errorf("failed to auto migrate database: %v", err)
		echan <- fmt.Errorf("failed to auto migrate database: %v", err)
	}

	fServer := functionsServer{
		db:                 db,
		reusableCache:      make(map[string]*cacheTuple),
		reusableCacheIndex: make(map[string]*cacheTuple),
	}

	err = util.GrpcStart(&grpcServer, "functions",
		fmt.Sprintf(":%d", port), func(srv *grpc.Server) {
			igrpc.RegisterFunctionsServiceServer(srv, &fServer)
			reflection.Register(srv)
		})
	if err != nil {
		echan <- err
	}

	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", functionsConfig.RedisBackend)
		},
	}

	conn := pool.Get()

	_, err = conn.Do("PING")
	if err != nil {
		echan <- fmt.Errorf("can't connect to redis, got error:\n%v", err)
	}

	go fServer.reusableGC()
	go fServer.orphansGC()

	go func() {

		rc := pool.Get()

		psc := redis.PubSubConn{Conn: rc}
		if err := psc.PSubscribe(FunctionsChannel); err != nil {
			logger.Error(err.Error())
		}

		for {
			switch v := psc.Receive().(type) {
			default:
				data, _ := json.Marshal(v)
				logger.Debug(string(data))
			case redis.Message:

				var tuples []*HeartbeatTuple

				err = json.Unmarshal(v.Data, &tuples)
				if err != nil {
					logger.Error(fmt.Sprintf("Unexpected notification on redis listener: %v", err))
				} else {
					go fServer.heartbeat(tuples)
				}

			}
		}

	}()

	err = fServer.reconstructServices(context.Background())
	if err != nil {
		echan <- err
	}

	select {}

}

// StopServer is stopping server gracefully
func StopServer() {
	if grpcServer != nil {
		grpcServer.GracefulStop()
	}
}

type HeartbeatTuple struct {
	NamespaceName      string
	NamespaceID        string
	WorkflowPath       string
	WorkflowID         string
	FunctionDefinition *model.ReusableFunctionDefinition
}

func (fServer *functionsServer) heartbeat(tuples []*HeartbeatTuple) {

	logger.Debugf("Workflow functions heartbeat received.")

	ctx := context.Background()

	for _, tuple := range tuples {

		size := int32(tuple.FunctionDefinition.Size)
		minscale := int32(tuple.FunctionDefinition.Scale)
		path := tuple.WorkflowPath
		path = strings.TrimPrefix(path, "/")
		path = strings.ReplaceAll(path, "_", "__")
		path = strings.ReplaceAll(path, "/", "_")

		in := &igrpc.CreateFunctionRequest{
			Info: &igrpc.BaseInfo{
				Name:      &tuple.FunctionDefinition.ID,
				Namespace: &tuple.NamespaceName,
				Workflow:  &path,
				Image:     &tuple.FunctionDefinition.Image,
				Cmd:       &tuple.FunctionDefinition.Cmd,
				Size:      &size,
				MinScale:  &minscale,
			},
		}

		name, _, err := GenerateServiceName(tuple.NamespaceName, path, tuple.FunctionDefinition.ID)
		if err != nil {
			logger.Errorf("Failed to generate service name for workflow function in heartbeat: %v", err)
			continue
		}

		fServer.reusableCacheLock.Lock()

		ct, exists := fServer.reusableCache[tuple.WorkflowID]
		if exists {
			ct.Add(name)
		} else {
			ct = new(cacheTuple)
			ct.Add(name)
			fServer.reusableCache[tuple.WorkflowID] = ct
		}
		fServer.reusableCacheIndex[name] = ct
		fServer.reusableCacheLock.Unlock()

		logger.Debugf("Creating workflow function in heartbeat: %s", name)

		_, err = fServer.CreateFunction(ctx, in)
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

		logger.Debugf("Reusable heartbeat garbage collector running.")

		cutoff := time.Now().Add(time.Minute * -15)

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

	ct.t = time.Now()

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

	logger.Debugf("Reusable heartbeat garbage collector purging workflow functions: %s", k)

	for i := range x.names {

		name := x.names[i]

		in := &igrpc.GetFunctionRequest{
			ServiceName: &name,
		}

		logger.Debugf("Reusable heartbeat garbage collector purging workflow function: %s", name)

		_, err := fServer.DeleteFunction(ctx, in)
		if err != nil {
			logger.Errorf("Reusable heartbeat garbage collector failed to purge workflow function: %v", err)
			continue
		}

	}

}

func (fServer *functionsServer) orphansGC() {

	ticker := time.NewTicker(time.Minute * 2)

	for {

		<-ticker.C

		logger.Debugf("Reusable orphans garbage collector running.")

		ctx := context.Background()

		filtered := map[string]string{
			"direktiv.io/scope": "workflow",
		}

		cs, err := fetchServiceAPI()
		if err != nil {
			err = fmt.Errorf("error getting clientset for knative: %v", err)
			logger.Errorf("Reusable orphans garbage collector failed to list workflow functions: %v", err)
			continue
		}

		lo := metav1.ListOptions{LabelSelector: labels.Set(filtered).String()}
		l, err := cs.ServingV1().Services(functionsConfig.Namespace).List(context.Background(), lo)
		if err != nil {
			logger.Errorf("Reusable orphans garbage collector failed to list workflow functions: %v", err)
			continue
		}

		for i := range l.Items {

			item := l.Items[i]

			fServer.reusableCacheLock.Lock()
			_, exists := fServer.reusableCacheIndex[item.Name]
			fServer.reusableCacheLock.Unlock()

			if !exists {

				logger.Debugf("Reusable orphans garbage collector deleting detected orphan function: %s", item.Name)

				_, err := fServer.DeleteFunction(ctx, &igrpc.GetFunctionRequest{
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
