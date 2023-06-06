package flow

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	libgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type actions struct {
	*server
	listener net.Listener
	srv      *libgrpc.Server
	grpc.UnimplementedActionsServer

	conn   *libgrpc.ClientConn
	client igrpc.FunctionsClient
}

func initActionsServer(ctx context.Context, srv *server) (*actions, error) {
	var err error

	actions := &actions{server: srv}

	actions.conn, err = util.GetEndpointTLS(srv.conf.FunctionsService + ":5555")
	if err != nil {
		return nil, err
	}

	actions.client = igrpc.NewFunctionsClient(actions.conn)

	actions.listener, err = net.Listen("tcp", ":4444")
	if err != nil {
		return nil, err
	}

	opts := util.GrpcServerOptions(unaryInterceptor, streamInterceptor)

	actions.srv = libgrpc.NewServer(opts...)

	grpc.RegisterActionsServer(actions.srv, actions)
	reflection.Register(actions.srv)

	go func() {
		<-ctx.Done()
		actions.srv.Stop()
	}()

	return actions, nil
}

func (actions *actions) Run() error {
	err := actions.srv.Serve(actions.listener)
	if err != nil {
		return err
	}

	return nil
}

func (actions *actions) SetNamespaceRegistry(ctx context.Context, req *grpc.SetNamespaceRegistryRequest) (*emptypb.Empty, error) {
	actions.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := actions.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	_, err = actions.client.StoreRegistry(ctx, &igrpc.FunctionsStoreRegistryRequest{
		Namespace: &cached.Namespace.Name,
		Data:      req.GetData(),
	})
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (actions *actions) DeleteNamespaceRegistry(ctx context.Context, req *grpc.DeleteNamespaceRegistryRequest) (*emptypb.Empty, error) {
	actions.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := actions.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	name := req.GetRegistry()

	_, err = actions.client.DeleteRegistry(ctx, &igrpc.FunctionsDeleteRegistryRequest{
		Namespace: &cached.Namespace.Name,
		Name:      &name,
	})
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil
}

type cpdRegistries struct {
	list []string
}

func newCustomPaginationDataRegistries() *cpdRegistries {
	cpd := new(cpdRegistries)

	cpd.list = make([]string, 0)

	return cpd
}

func (cpds *cpdRegistries) Total() int {
	return len(cpds.list)
}

func (cpds *cpdRegistries) ID(idx int) string {
	return cpds.list[idx]
}

func (cpds *cpdRegistries) Value(idx int) map[string]interface{} {
	return map[string]interface{}{
		"name": cpds.list[idx],
	}
}

func (cpds *cpdRegistries) Filter(filter *grpc.PageFilter) error {
	if filter == nil {
		return nil
	}

	if filter.GetField() != "" && filter.GetField() != util.PaginationKeyName {
		return fmt.Errorf("invalid filter field: %s", filter.GetField())
	}

	switch filter.GetType() {
	case "":
	default:
		return fmt.Errorf("invalid filter type: %s", filter.GetType())
	}

	arg := filter.GetVal()

	secrets := make([]string, 0)

	for _, secret := range cpds.list {
		if strings.Contains(secret, arg) {
			secrets = append(secrets, secret)
		}
	}

	cpds.list = secrets

	return nil
}

func (cpds *cpdRegistries) Order(order *grpc.PageOrder) error {
	if order.GetField() != "" && order.GetField() != util.PaginationKeyName {
		return fmt.Errorf("invalid order field: %s", order.GetField())
	}

	sort.Strings(cpds.list)

	switch order.GetDirection() {
	case "":
		fallthrough
	case paginationOrderingASC:
	case paginationOrderingDESC:
		sort.Sort(sort.Reverse(sort.StringSlice(cpds.list)))
	default:
		return fmt.Errorf("invalid order direction: %s", order.GetDirection())
	}

	return nil
}

func (cpds *cpdRegistries) Add(name string) {
	cpds.list = append(cpds.list, name)
}

func (actions *actions) NamespaceRegistries(ctx context.Context, req *grpc.NamespaceRegistriesRequest) (*grpc.NamespaceRegistriesResponse, error) {
	actions.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := actions.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	response, err := actions.client.GetRegistries(ctx, &igrpc.FunctionsGetRegistriesRequest{
		Namespace: &cached.Namespace.Name,
	})
	if err != nil {
		return nil, err
	}

	cpds := newCustomPaginationDataRegistries()
	pagination := newCustomPagination(cpds)
	for i := range response.Registries {
		cpds.Add(response.Registries[i].GetName())
	}

	cx, err := pagination.Paginate(p)
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespaceRegistriesResponse

	resp.Namespace = cached.Namespace.Name
	resp.Registries = new(grpc.Registries)
	resp.Registries.PageInfo = new(grpc.PageInfo)

	err = bytedata.ConvertDataForOutput(cx, &resp.Registries)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (actions *actions) NamespaceRegistriesStream(req *grpc.NamespaceRegistriesRequest, srv grpc.Actions_NamespaceRegistriesStreamServer) error {
	actions.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	cached := new(database.CacheData)

	err := actions.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return err
	}

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	sub := actions.pubsub.SubscribeNamespaceRegistries(cached.Namespace)
	defer actions.cleanup(sub.Close)

resend:

	namespace := cached.Namespace.ID.String()
	response, err := actions.client.GetRegistries(ctx, &igrpc.FunctionsGetRegistriesRequest{
		Namespace: &namespace,
	})
	if err != nil {
		return err
	}

	cpds := newCustomPaginationDataRegistries()
	pagination := newCustomPagination(cpds)
	for i := range response.Registries {
		cpds.Add(response.Registries[i].GetName())
	}

	cx, err := pagination.Paginate(p)
	if err != nil {
		return err
	}

	resp := new(grpc.NamespaceRegistriesResponse)

	resp.Namespace = cached.Namespace.Name
	resp.Registries = new(grpc.Registries)
	resp.Registries.PageInfo = new(grpc.PageInfo)

	err = bytedata.ConvertDataForOutput(cx, &resp.Registries)
	if err != nil {
		return err
	}

	nhash = bytedata.Checksum(resp)
	if nhash != phash {
		err = srv.Send(resp)
		if err != nil {
			return err
		}
	}
	phash = nhash

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend
}

func (actions *actions) CancelWorkflowInstance(svn, actionID string) error {
	actions.sugar.Debugf("Handling gRPC request: %s", this())

	req := &igrpc.FunctionsCancelWorkflowRequest{
		ServiceName: &svn,
		ActionID:    &actionID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := actions.client.CancelWorfklow(ctx, req)

	return err
}
