package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/gorilla/mux"
)

type vars struct {
	*server
	listener net.Listener
	http     *http.Server
	router   *mux.Router
}

func initVarsServer(ctx context.Context, srv *server) (*vars, error) {

	var err error

	vars := new(vars)

	vars.server = srv
	vars.listener, err = net.Listen("tcp", "0.0.0.0:9999")
	if err != nil {
		return nil, err
	}

	vars.router = mux.NewRouter()

	vars.router.HandleFunc("/api/vars/namespaces/{namespace}/vars/{var}", vars.nsHandler)
	vars.router.HandleFunc("/api/vars/namespaces/{namespace}/workflows/{path:.*}/vars/{var}", vars.wfHandler)
	vars.router.HandleFunc("/api/vars/namespaces/{namespace}/instances/{instance}/vars/{var}", vars.inHandler)

	vars.http = &http.Server{
		Addr:              ":9999",
		Handler:           vars.router,
		ReadHeaderTimeout: time.Second * 60,
	}

	go func() {

		defer func() {
			_ = recover()
		}()

		<-ctx.Done()

		err := vars.Close()
		if err != nil {
			vars.sugar.Error(err)
		}

	}()

	return vars, nil

}

func (vars *vars) Close() error {

	err := vars.http.Close()
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	err = vars.listener.Close()
	if err != nil {
		if !errors.Is(err, net.ErrClosed) {
			return err
		}
	}

	return nil

}

func (vars *vars) Run() error {

	err := vars.http.Serve(vars.listener)
	if err != nil {
		if err != http.ErrServerClosed {
			return err
		}
	}

	return nil

}

func (vars *vars) nsHandler(w http.ResponseWriter, r *http.Request) {

	vars.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := r.Context()

	namespace := mux.Vars(r)["namespace"]
	key := mux.Vars(r)["var"]
	mimeType := r.Header.Get("Content-Type")

	switch r.Method {
	case http.MethodGet:

		resp, err := vars.flow.NamespaceVariable(ctx, &grpc.NamespaceVariableRequest{
			Namespace: namespace,
			Key:       key,
		})
		if err != nil {
			code := http.StatusInternalServerError
			msg := http.StatusText(code)
			http.Error(w, msg, code)
			return
		}

		_, err = io.Copy(w, bytes.NewReader(resp.Data))
		if err != nil {
			vars.sugar.Error(err)
			return
		}

		return

	case http.MethodPut:

		data, err := io.ReadAll(r.Body)
		if err != nil {
			vars.sugar.Error(err)
			return
		}

		req := new(grpc.SetNamespaceVariableRequest)
		req.Data = data
		req.Key = key
		req.Namespace = namespace
		req.TotalSize = int64(len(data))
		req.MimeType = mimeType

		_, err = vars.flow.SetNamespaceVariable(ctx, req)
		if err != nil {
			code := http.StatusInternalServerError
			msg := http.StatusText(code)
			http.Error(w, msg, code)
			return
		}

		return

	}

}

func (vars *vars) wfHandler(w http.ResponseWriter, r *http.Request) {

	vars.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := r.Context()

	namespace := mux.Vars(r)["namespace"]
	path := mux.Vars(r)["path"]
	key := mux.Vars(r)["var"]
	mimeType := r.Header.Get("Content-Type")

	switch r.Method {
	case http.MethodGet:

		resp, err := vars.flow.WorkflowVariable(ctx, &grpc.WorkflowVariableRequest{
			Namespace: namespace,
			Path:      path,
			Key:       key,
		})
		if err != nil {
			code := http.StatusInternalServerError
			msg := http.StatusText(code)
			http.Error(w, msg, code)
			return
		}

		_, err = io.Copy(w, bytes.NewReader(resp.Data))
		if err != nil {
			vars.sugar.Error(err)
			return
		}

		return

	case http.MethodPut:

		data, err := io.ReadAll(r.Body)
		if err != nil {
			vars.sugar.Error(err)
			return
		}

		req := new(grpc.SetWorkflowVariableRequest)
		req.Data = data
		req.Key = key
		req.Namespace = namespace
		req.Path = path
		req.TotalSize = int64(len(data))
		req.MimeType = mimeType

		_, err = vars.flow.SetWorkflowVariable(ctx, req)
		if err != nil {
			code := http.StatusInternalServerError
			msg := http.StatusText(code)
			http.Error(w, msg, code)
			return
		}

		return

	}

}

func (vars *vars) inHandler(w http.ResponseWriter, r *http.Request) {

	vars.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := r.Context()

	namespace := mux.Vars(r)["namespace"]
	instance := mux.Vars(r)["instance"]
	key := mux.Vars(r)["var"]
	mimeType := r.Header.Get("Content-Type")

	switch r.Method {
	case http.MethodGet:

		resp, err := vars.flow.InstanceVariable(ctx, &grpc.InstanceVariableRequest{
			Namespace: namespace,
			Instance:  instance,
			Key:       key,
		})
		if err != nil {
			code := http.StatusInternalServerError
			msg := http.StatusText(code)
			http.Error(w, msg, code)
			return
		}

		_, err = io.Copy(w, bytes.NewReader(resp.Data))
		if err != nil {
			vars.sugar.Error(err)
			return
		}

		return

	case http.MethodPut:

		data, err := io.ReadAll(r.Body)
		if err != nil {
			vars.sugar.Error(err)
			return
		}

		req := new(grpc.SetInstanceVariableRequest)
		req.Data = data
		req.Key = key
		req.Namespace = namespace
		req.Instance = instance
		req.TotalSize = int64(len(data))
		req.MimeType = mimeType

		_, err = vars.flow.SetInstanceVariable(ctx, req)
		if err != nil {
			code := http.StatusInternalServerError
			msg := http.StatusText(code)
			http.Error(w, msg, code)
			return
		}

		return

	}

}
