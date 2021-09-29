package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/vorteil/direktiv/pkg/flow"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func testNamespaceVariablesSmall(ctx context.Context, c grpc.FlowClient) error {

	namespace := testNamespace()

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	vars, err := c.NamespaceVariables(ctx, &grpc.NamespaceVariablesRequest{
		Namespace: namespace,
	})
	if err != nil {
		return err
	}

	if len(vars.Variables.Edges) != 0 {
		return errors.New("unexpected variables already exist in the namespace")
	}

	client, err := c.NamespaceVariablesStream(ctx, &grpc.NamespaceVariablesRequest{
		Namespace: namespace,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	vars, err = client.Recv()
	if err != nil {
		return err
	}
	if len(vars.Variables.Edges) != 0 {
		return errors.New("unexpected variables already exist in the namespace")
	}

	_, err = c.SetNamespaceVariable(ctx, &grpc.SetNamespaceVariableRequest{
		Namespace: namespace,
		Key:       "testvar",
		Data:      []byte("MyVar"),
	})
	if err != nil {
		return err
	}

	vars, err = client.Recv()
	if err != nil {
		return err
	}
	if len(vars.Variables.Edges) != 1 {
		return errors.New("incorrect number of variables returned by server")
	}

	v, err := c.NamespaceVariable(ctx, &grpc.NamespaceVariableRequest{
		Namespace: namespace,
		Key:       "testvar",
	})
	if err != nil {
		return err
	}

	if string(v.Data) != "MyVar" {
		return errors.New("unexpected variable data")
	}

	_, err = c.SetNamespaceVariable(ctx, &grpc.SetNamespaceVariableRequest{
		Namespace: namespace,
		Key:       "testvar",
		Data:      []byte("Direktiv"),
	})
	if err != nil {
		return err
	}

	vars, err = client.Recv()
	if err != nil {
		return err
	}
	if len(vars.Variables.Edges) != 1 {
		return errors.New("incorrect number of variables returned by server")
	}

	v, err = c.NamespaceVariable(ctx, &grpc.NamespaceVariableRequest{
		Namespace: namespace,
		Key:       "testvar",
	})
	if err != nil {
		return err
	}

	if string(v.Data) != "Direktiv" {
		return errors.New("unexpected variable data")
	}

	_, err = c.DeleteNamespaceVariable(ctx, &grpc.DeleteNamespaceVariableRequest{
		Namespace: namespace,
		Key:       "testvar",
	})
	if err != nil {
		return err
	}

	vars, err = client.Recv()
	if err != nil {
		return err
	}
	if len(vars.Variables.Edges) != 0 {
		return errors.New("unexpected variables still exist in the namespace")
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	v, err = c.NamespaceVariable(ctx, &grpc.NamespaceVariableRequest{
		Namespace: namespace,
		Key:       "testvar",
	})
	if err == nil {
		return errors.New("server returned non-existent variable without error")
	}
	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	return nil

}

func testNamespaceVariablesLarge(ctx context.Context, c grpc.FlowClient) error {

	namespace := testNamespace()

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	data := bytes.Repeat([]byte("a"), 64*1024*1024) // 64 MiB

	_, err = c.SetNamespaceVariable(ctx, &grpc.SetNamespaceVariableRequest{
		Namespace: namespace,
		Key:       "testvar",
		Data:      data,
	})
	if err == nil {
		return errors.New("server accepted oversized variable without error")
	}
	if status.Code(err) != codes.ResourceExhausted {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	client, err := c.SetNamespaceVariableParcels(ctx)
	if err != nil {
		return err
	}
	defer client.CloseSend()

	req := &grpc.SetNamespaceVariableRequest{
		Namespace: namespace,
		Key:       "testvar",
		TotalSize: int64(len(data)),
		Data:      data[:2*1024*1024],
	}

	for i := 0; i < 32; i++ {
		err = client.Send(req)
		if err != nil {
			return err
		}
	}

	resp, err := client.CloseAndRecv()
	if err != nil {
		return err
	}

	hasher := sha256.New()
	_, _ = io.Copy(hasher, bytes.NewReader(data))
	hash := hasher.Sum(nil)
	checksum := hex.EncodeToString(hash)

	if checksum != resp.Checksum {
		return errors.New("server calculated checksum doesn't match expectations")
	}

	_, err = c.NamespaceVariable(ctx, &grpc.NamespaceVariableRequest{
		Namespace: namespace,
		Key:       "testvar",
	})
	if err == nil {
		return errors.New("server tried to return oversized variable without error")
	}
	if status.Code(err) != codes.ResourceExhausted {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	vc, err := c.NamespaceVariableParcels(ctx, &grpc.NamespaceVariableRequest{
		Namespace: namespace,
		Key:       "testvar",
	})
	if err != nil {
		return err
	}
	defer vc.CloseSend()

	hasher = sha256.New()

	for {
		resp, err := vc.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		_, _ = io.Copy(hasher, bytes.NewReader(resp.Data))
	}

	hash = hasher.Sum(nil)
	returnedChecksum := hex.EncodeToString(hash)

	if checksum != returnedChecksum {
		return errors.New("server returned different data than what was sent")
	}

	return nil

}

func testWorkflowVariablesSmall(ctx context.Context, c grpc.FlowClient) error {

	namespace := testNamespace()

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source:    []byte(simpleWorkflow),
	})
	if err != nil {
		return err
	}

	vars, err := c.WorkflowVariables(ctx, &grpc.WorkflowVariablesRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err != nil {
		return err
	}

	if len(vars.Variables.Edges) != 0 {
		return errors.New("unexpected variables already exist in the workflow")
	}

	client, err := c.WorkflowVariablesStream(ctx, &grpc.WorkflowVariablesRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	vars, err = client.Recv()
	if err != nil {
		return err
	}
	if len(vars.Variables.Edges) != 0 {
		return errors.New("unexpected variables already exist in the workflow")
	}

	_, err = c.SetWorkflowVariable(ctx, &grpc.SetWorkflowVariableRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Key:       "testvar",
		Data:      []byte("MyVar"),
	})
	if err != nil {
		return err
	}

	vars, err = client.Recv()
	if err != nil {
		return err
	}
	if len(vars.Variables.Edges) != 1 {
		return errors.New("incorrect number of variables returned by server")
	}

	v, err := c.WorkflowVariable(ctx, &grpc.WorkflowVariableRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Key:       "testvar",
	})
	if err != nil {
		return err
	}

	if string(v.Data) != "MyVar" {
		return errors.New("unexpected variable data")
	}

	_, err = c.SetWorkflowVariable(ctx, &grpc.SetWorkflowVariableRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Key:       "testvar",
		Data:      []byte("Direktiv"),
	})
	if err != nil {
		return err
	}

	vars, err = client.Recv()
	if err != nil {
		return err
	}
	if len(vars.Variables.Edges) != 1 {
		return errors.New("incorrect number of variables returned by server")
	}

	v, err = c.WorkflowVariable(ctx, &grpc.WorkflowVariableRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Key:       "testvar",
	})
	if err != nil {
		return err
	}

	if string(v.Data) != "Direktiv" {
		return errors.New("unexpected variable data")
	}

	_, err = c.DeleteWorkflowVariable(ctx, &grpc.DeleteWorkflowVariableRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Key:       "testvar",
	})
	if err != nil {
		return err
	}

	vars, err = client.Recv()
	if err != nil {
		return err
	}
	if len(vars.Variables.Edges) != 0 {
		return errors.New("unexpected variables still exist in the workflow")
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	v, err = c.WorkflowVariable(ctx, &grpc.WorkflowVariableRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Key:       "testvar",
	})
	if err == nil {
		return errors.New("server returned non-existent variable without error")
	}
	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	return nil

}

func testWorkflowVariablesLarge(ctx context.Context, c grpc.FlowClient) error {

	namespace := testNamespace()

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	data := bytes.Repeat([]byte("a"), 64*1024*1024) // 64 MiB

	_, err = c.SetNamespaceVariable(ctx, &grpc.SetNamespaceVariableRequest{
		Namespace: namespace,
		Key:       "testvar",
		Data:      data,
	})
	if err == nil {
		return errors.New("server accepted oversized variable without error")
	}
	if status.Code(err) != codes.ResourceExhausted {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	client, err := c.SetNamespaceVariableParcels(ctx)
	if err != nil {
		return err
	}
	defer client.CloseSend()

	req := &grpc.SetNamespaceVariableRequest{
		Namespace: namespace,
		Key:       "testvar",
		TotalSize: int64(len(data)),
		Data:      data[:2*1024*1024],
	}

	for i := 0; i < 32; i++ {
		err = client.Send(req)
		if err != nil {
			return err
		}
	}

	resp, err := client.CloseAndRecv()
	if err != nil {
		return err
	}

	hasher := sha256.New()
	_, _ = io.Copy(hasher, bytes.NewReader(data))
	hash := hasher.Sum(nil)
	checksum := hex.EncodeToString(hash)

	if checksum != resp.Checksum {
		return errors.New("server calculated checksum doesn't match expectations")
	}

	_, err = c.NamespaceVariable(ctx, &grpc.NamespaceVariableRequest{
		Namespace: namespace,
		Key:       "testvar",
	})
	if err == nil {
		return errors.New("server tried to return oversized variable without error")
	}
	if status.Code(err) != codes.ResourceExhausted {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	vc, err := c.NamespaceVariableParcels(ctx, &grpc.NamespaceVariableRequest{
		Namespace: namespace,
		Key:       "testvar",
	})
	if err != nil {
		return err
	}
	defer vc.CloseSend()

	hasher = sha256.New()

	for {
		resp, err := vc.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		_, _ = io.Copy(hasher, bytes.NewReader(resp.Data))
	}

	hash = hasher.Sum(nil)
	returnedChecksum := hex.EncodeToString(hash)

	if checksum != returnedChecksum {
		return errors.New("server returned different data than what was sent")
	}

	return nil

}

func testInstanceNamespaceVariables(ctx context.Context, c grpc.FlowClient) error {

	namespace := testNamespace()

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf1",
		Source: []byte(`
states:
  - id: a
    type: noop 
    transform: 
      k: 5
    transition: b
  - id: b
    type: switch 
    conditions:
      - condition: 'jq(.k > 0)'
        transition: c
        transform: 'jq(.k -= 1)'
    defaultTransition: e
  - id: c
    type: getter
    variables:
      - key: x
        scope: namespace 
    transform: 'jq(.var.x += 1)'
    transition: d
  - id: d
    type: setter
    variables:
      - key: x
        scope: namespace 
        value: 'jq(.var.x)'
    transition: b
  - id: e
    type: noop
`),
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf2",
		Source: []byte(`
states:
  - id: a
    type: noop 
    transform: 
      k: 5
    transition: b
  - id: b
    type: switch 
    conditions:
      - condition: 'jq(.k > 0)'
        transition: c
        transform: 'jq(.k -= 1)'
    defaultTransition: e
  - id: c
    type: getter
    variables:
      - key: x
        scope: namespace 
    transform: 'jq(.var.x += 1)'
    transition: d
  - id: d
    type: setter
    variables:
      - key: x
        scope: namespace
        value: 'jq(.var.x)'
    transition: b
  - id: e
    type: noop
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf1",
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrCode, iresp.Instance.ErrMessage)
	}

	v, err := c.NamespaceVariable(ctx, &grpc.NamespaceVariableRequest{
		Namespace: namespace,
		Key:       "x",
	})
	if err != nil {
		return err
	}

	if string(v.Data) != "5" {
		return errors.New("unexpected namespace variable data")
	}

	resp, err = c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf2",
	})
	if err != nil {
		return err
	}

	cctx, cancel = context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	client, err = c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrCode, iresp.Instance.ErrMessage)
	}

	v, err = c.NamespaceVariable(ctx, &grpc.NamespaceVariableRequest{
		Namespace: namespace,
		Key:       "x",
	})
	if err != nil {
		return err
	}

	if string(v.Data) != "10" {
		return errors.New("unexpected namespace variable data")
	}

	return nil

}

func testInstanceWorkflowVariables(ctx context.Context, c grpc.FlowClient) error {

	namespace := testNamespace()

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source: []byte(`
states:
  - id: a
    type: noop 
    transform: 
      k: 5
    transition: b
  - id: b
    type: switch 
    conditions:
      - condition: 'jq(.k > 0)'
        transition: c
        transform: 'jq(.k -= 1)'
    defaultTransition: e
  - id: c
    type: getter
    variables:
      - key: x
        scope: workflow 
    transform: 'jq(.var.x += 1)'
    transition: d
  - id: d
    type: setter
    variables:
      - key: x
        scope: workflow 
        value: 'jq(.var.x)'
    transition: b
  - id: e
    type: noop
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrCode, iresp.Instance.ErrMessage)
	}

	v, err := c.WorkflowVariable(ctx, &grpc.WorkflowVariableRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Key:       "x",
	})
	if err != nil {
		return err
	}

	if string(v.Data) != "5" {
		return errors.New("unexpected workflow variable data")
	}

	resp, err = c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err != nil {
		return err
	}

	cctx, cancel = context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	client, err = c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrCode, iresp.Instance.ErrMessage)
	}

	v, err = c.WorkflowVariable(ctx, &grpc.WorkflowVariableRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Key:       "x",
	})
	if err != nil {
		return err
	}

	if string(v.Data) != "10" {
		return errors.New("unexpected workflow variable data")
	}

	return nil

}

func testInstanceInstanceVariables(ctx context.Context, c grpc.FlowClient) error {

	namespace := testNamespace()

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source: []byte(`
states:
  - id: a
    type: noop 
    transform: 
      k: 5
    transition: b
  - id: b
    type: switch 
    conditions:
      - condition: 'jq(.k > 0)'
        transition: c
        transform: 'jq(.k -= 1)'
    defaultTransition: e
  - id: c
    type: getter
    variables:
      - key: x
        scope: instance 
    transform: 'jq(.var.x += 1)'
    transition: d
  - id: d
    type: setter
    variables:
      - key: x
        scope: instance 
        value: 'jq(.var.x)'
    transition: b
  - id: e
    type: noop
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	client, err := c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	var iresp, x *grpc.InstanceResponse

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrCode, iresp.Instance.ErrMessage)
	}

	v, err := c.InstanceVariable(ctx, &grpc.InstanceVariableRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
		Key:       "x",
	})
	if err != nil {
		return err
	}

	if string(v.Data) != "5" {
		return errors.New("unexpected instance variable data")
	}

	resp, err = c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err != nil {
		return err
	}

	cctx, cancel = context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	client, err = c.InstanceStream(cctx, &grpc.InstanceRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	for {
		x, err = client.Recv()
		if err != nil {
			return err
		}
		iresp = x

		if iresp.Instance.Status != flow.StatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != flow.StatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrCode, iresp.Instance.ErrMessage)
	}

	v, err = c.InstanceVariable(ctx, &grpc.InstanceVariableRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
		Key:       "x",
	})
	if err != nil {
		return err
	}

	if string(v.Data) != "5" {
		return errors.New("unexpected instance variable data")
	}

	return nil

}
