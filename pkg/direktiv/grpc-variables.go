package direktiv

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"

	"github.com/vorteil/direktiv/pkg/ingress"
)

func (is *ingressServer) SetNamespaceVariable(srv ingress.DirektivIngress_SetNamespaceVariableServer) error {

	ctx := srv.Context()

	in, err := srv.Recv()
	if err != nil {
		return err
	}

	namespace := in.GetNamespace()
	if namespace == "" {
		return errors.New("required namespace")
	}

	key := in.GetKey()
	if key == "" {
		return errors.New("requires variable key")
	}

	totalSize := in.GetTotalSize()

	_, err = is.wfServer.dbManager.getNamespace(namespace)
	if err != nil {
		return err
	}

	if totalSize == 0 {
		err = is.wfServer.variableStorage.Delete(ctx, key, namespace)
		if err != nil {
			return err
		}
	} else {
		w, err := is.wfServer.variableStorage.Store(ctx, key, namespace)
		if err != nil {
			return err
		}

		var totalRead int64

		for {
			data := in.GetValue()
			k, err := io.Copy(w, bytes.NewReader(data))
			if err != nil {
				return err
			}
			totalRead += k
			if totalRead >= totalSize {
				break
			}
			in, err = srv.Recv()
			if err != nil {
				return err
			}
		}

		err = w.Close()
		if err != nil {
			return err
		}

		// TODO: resolve edge-case where namespace or workflow is deleted during this process
	}

	return nil

}

func (is *ingressServer) GetNamespaceVariable(in *ingress.GetNamespaceVariableRequest, out ingress.DirektivIngress_GetNamespaceVariableServer) error {

	ctx := out.Context()

	namespace := in.GetNamespace()
	if namespace == "" {
		return errors.New("required namespace")
	}

	key := in.GetKey()
	if key == "" {
		return errors.New("requires variable key")
	}

	r, err := is.wfServer.variableStorage.Retrieve(ctx, key, namespace)
	if err != nil {
		return err
	}
	defer r.Close()

	// break data into chunks
	var chunks int
	chunkSize := int64(grpcChunkSize)
	totalSize := r.Size()
	for {
		cr := io.LimitReader(r, grpcChunkSize)
		data, err := ioutil.ReadAll(cr)
		if err != nil {
			return err
		}
		if chunks > 0 && len(data) == 0 {
			break
		}
		resp := new(ingress.GetNamespaceVariableResponse)
		resp.Value = data
		resp.TotalSize = &totalSize
		resp.ChunkSize = &chunkSize
		err = out.Send(resp)
		if err != nil {
			return err
		}
		chunks++
	}

	err = r.Close()
	if err != nil {
		return err
	}

	return nil

}

func (is *ingressServer) ListNamespaceVariables(ctx context.Context, in *ingress.ListNamespaceVariablesRequest) (*ingress.ListNamespaceVariablesResponse, error) {

	resp := new(ingress.ListNamespaceVariablesResponse)

	namespace := in.GetNamespace()
	if namespace == "" {
		return nil, errors.New("required namespace")
	}

	list, err := is.wfServer.variableStorage.List(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var names []string
	var sizes []int64

	for i, tuple := range list {
		names = append(names, tuple.Key())
		sizes = append(sizes, tuple.Size())
		resp.Variables = append(resp.Variables, &ingress.ListNamespaceVariablesResponse_Variable{
			Name: &(names[i]),
			Size: &(sizes[i]),
		})
	}

	return resp, nil

}

func (is *ingressServer) SetWorkflowVariable(srv ingress.DirektivIngress_SetWorkflowVariableServer) error {

	ctx := srv.Context()

	in, err := srv.Recv()
	if err != nil {
		return err
	}

	workflow := in.GetWorkflowUid()
	if workflow == "" {
		return errors.New("required workflow uid")
	}

	key := in.GetKey()
	if key == "" {
		return errors.New("requires variable key")
	}

	totalSize := in.GetTotalSize()

	wf, err := is.wfServer.dbManager.getWorkflowByUid(ctx, workflow)
	if err != nil {
		return err
	}

	namespace := wf.Edges.Namespace.ID
	wfId := wf.ID.String()

	if totalSize == 0 {
		err = is.wfServer.variableStorage.Delete(ctx, key, namespace, wfId)
		if err != nil {
			return err
		}
	} else {
		w, err := is.wfServer.variableStorage.Store(ctx, key, namespace, wfId)
		if err != nil {
			return err
		}

		var totalRead int64

		for {
			data := in.GetValue()
			k, err := io.Copy(w, bytes.NewReader(data))
			if err != nil {
				return err
			}
			totalRead += k
			if totalRead >= totalSize {
				break
			}
			in, err = srv.Recv()
			if err != nil {
				return err
			}
		}

		err = w.Close()
		if err != nil {
			return err
		}

		// TODO: resolve edge-case where namespace or workflow is deleted during this process
	}

	return nil

}

func (is *ingressServer) GetWorkflowVariable(in *ingress.GetWorkflowVariableRequest, out ingress.DirektivIngress_GetWorkflowVariableServer) error {

	ctx := out.Context()

	workflow := in.GetWorkflowUid()
	if workflow == "" {
		return errors.New("required workflow uid")
	}

	key := in.GetKey()
	if key == "" {
		return errors.New("requires variable key")
	}

	wf, err := is.wfServer.dbManager.getWorkflowByUid(ctx, workflow)
	if err != nil {
		return err
	}

	ns := wf.Edges.Namespace

	r, err := is.wfServer.variableStorage.Retrieve(ctx, key, ns.ID, workflow)
	if err != nil {
		return err
	}
	defer r.Close()

	// break data into chunks
	var chunks int
	chunkSize := int64(grpcChunkSize)
	totalSize := r.Size()
	for {
		cr := io.LimitReader(r, grpcChunkSize)
		data, err := ioutil.ReadAll(cr)
		if err != nil {
			return err
		}
		if chunks > 0 && len(data) == 0 {
			break
		}
		resp := new(ingress.GetWorkflowVariableResponse)
		resp.Value = data
		resp.TotalSize = &totalSize
		resp.ChunkSize = &chunkSize
		err = out.Send(resp)
		if err != nil {
			return err
		}
		chunks++
	}

	err = r.Close()
	if err != nil {
		return err
	}

	return nil

}

func (is *ingressServer) ListWorkflowVariables(ctx context.Context, in *ingress.ListWorkflowVariablesRequest) (*ingress.ListWorkflowVariablesResponse, error) {

	resp := new(ingress.ListWorkflowVariablesResponse)

	workflow := in.GetWorkflowUid()
	if workflow == "" {
		return nil, errors.New("required workflow uid")
	}

	wf, err := is.wfServer.dbManager.getWorkflowByUid(ctx, workflow)
	if err != nil {
		return nil, err
	}

	ns := wf.Edges.Namespace

	list, err := is.wfServer.variableStorage.List(ctx, ns.ID, workflow)
	if err != nil {
		return nil, err
	}

	var names []string
	var sizes []int64

	for i, tuple := range list {
		names = append(names, tuple.Key())
		sizes = append(sizes, tuple.Size())
		resp.Variables = append(resp.Variables, &ingress.ListWorkflowVariablesResponse_Variable{
			Name: &(names[i]),
			Size: &(sizes[i]),
		})
	}

	return resp, nil

}
