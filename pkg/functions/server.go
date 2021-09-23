package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
		db: db,
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

	data, _ := json.Marshal(tuples)

	logger.Info("HEARTBEAT:" + string(data))

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

		_, err := fServer.CreateFunction(ctx, in)
		if err != nil {
			logger.Error(err)
		}

	}

}
