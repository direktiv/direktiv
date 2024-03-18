package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	libgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type flow struct {
	*server
	listener net.Listener
	srv      *libgrpc.Server
	grpc.UnimplementedFlowServer
}

const srv = "server"

func initFlowServer(ctx context.Context, srv *server) (*flow, error) {
	var err error

	flow := &flow{server: srv}

	flow.listener, err = net.Listen("tcp", ":6666")
	if err != nil {
		return nil, err
	}

	opts := util.GrpcServerOptions(unaryInterceptor, streamInterceptor)

	flow.srv = libgrpc.NewServer(opts...)

	grpc.RegisterFlowServer(flow.srv, flow)
	reflection.Register(flow.srv)

	go func() {
		<-ctx.Done()
		flow.srv.Stop()
	}()

	go func() {
		// instance garbage collector
		ctx := context.Background()
		<-time.After(2 * time.Minute)

		for {
			<-time.After(time.Hour)
			t := time.Now().UTC().Add(time.Hour * -24)

			tx, err := srv.flow.beginSqlTx(ctx)
			if err != nil {
				flow.sugar.Error(fmt.Errorf("failed to get transaction to cleanup old instances: %w", err))
				continue
			}

			err = tx.InstanceStore().DeleteOldInstances(ctx, t)
			if err != nil {
				tx.Rollback()
				flow.sugar.Error(fmt.Errorf("failed to cleanup old instances: %w", err))
				continue
			}

			err = tx.Commit(ctx)
			if err != nil {
				flow.sugar.Error(fmt.Errorf("failed to commit tx to cleanup old instances: %w", err))
				continue
			}

			// TODO: alan: cleanup old instance variables.
		}
	}()

	go func() {
		// logs garbage collector
		<-time.After(3 * time.Minute)
		for {
			<-time.After(time.Hour)
			t := time.Now().UTC().Add(time.Hour * -48) // TODO make this a config option.
			flow.sugar.Error(fmt.Sprintf("deleting all logs since %v", t))
			err = srv.flow.runSqlTx(ctx, func(tx *sqlTx) error {
				return tx.DataStore().Logs().DeleteOldLogs(context.TODO(), t)
			})
			if err != nil {
				flow.sugar.Error(fmt.Errorf("failed to cleanup old logs: %w", err))
				continue
			}
			err = srv.flow.runSqlTx(ctx, func(tx *sqlTx) error {
				return tx.DataStore().NewLogs().DeleteOldLogs(ctx, t)
			})
			if err != nil {
				flow.sugar.Error(fmt.Errorf("failed to cleanup old logs: %w", err))
				continue
			}
		}
	}()

	go func() {
		// timed-out instance retrier
		<-time.After(1 * time.Minute)
		ticker := time.NewTicker(5 * time.Minute)
		for {
			<-ticker.C
			go flow.kickExpiredInstances()
		}
	}()

	return flow, nil
}

func (flow *flow) kickExpiredInstances() {
	ctx := context.Background()

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		flow.sugar.Error(err)
		return
	}
	defer tx.Rollback()

	list, err := tx.InstanceStore().GetHangingInstances(ctx)
	if err != nil {
		flow.sugar.Error(err)
		return
	}

	for i := range list {
		data, err := json.Marshal(&retryMessage{
			InstanceID: list[i].ID.String(),
		})
		if err != nil {
			panic(err)
		}

		flow.engine.retryWakeup(data)
	}
}

func (flow *flow) Run() error {
	err := flow.srv.Serve(flow.listener)
	if err != nil {
		return err
	}

	return nil
}

func (flow *flow) JQ(ctx context.Context, req *grpc.JQRequest) (*grpc.JQResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	var input interface{}

	data := req.GetData()

	err := json.Unmarshal(data, &input)
	if err != nil {
		err = status.Error(codes.InvalidArgument, fmt.Sprintf("invalid json data: %v", err))
		return nil, err
	}

	command := "jq(" + req.GetQuery() + ")"

	results, err := jq(input, command)
	if err != nil {
		err = status.Error(codes.InvalidArgument, fmt.Sprintf("error executing JQ command: %v", err))
		return nil, err
	}

	var strs []string

	for _, result := range results {
		x, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, err
		}

		strs = append(strs, string(x))
	}

	var resp grpc.JQResponse

	resp.Query = req.GetQuery()
	resp.Data = req.GetData()
	resp.Results = strs

	return &resp, nil
}

func (flow *flow) GetAttributes() map[string]string {
	tags := make(map[string]string)
	tags["recipientType"] = srv
	return tags
}
