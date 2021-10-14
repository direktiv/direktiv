package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/vorteil/direktiv/pkg/flow"

	"github.com/vorteil/direktiv/pkg/flow/grpc"
)

func testStartWorkflow(ctx context.Context, c grpc.FlowClient, namespace string) error {

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

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
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
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	if iresp.Instance.As != "/testwf" {
		return errors.New("instance returned incorrect 'As'")
	}

	if len(iresp.Flow) != 1 || iresp.Flow[0] != "a" {
		return errors.New("incorrect flow array")
	}

	input, err := c.InstanceInput(ctx, &grpc.InstanceInputRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	if len(input.Data) != 0 {
		return errors.New("instance input should have been zero length")
	}

	output, err := c.InstanceOutput(ctx, &grpc.InstanceOutputRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	if string(output.Data) != `{"input":""}` {
		return errors.New("unexpected instance output")
	}

	logs, err := c.InstanceLogs(ctx, &grpc.InstanceLogsRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	if len(logs.Edges) == 0 {
		return errors.New("missing instance logs")
	}

	return nil

}

func testStateLogSimple(ctx context.Context, c grpc.FlowClient, namespace string) error {

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
    log: "Hello, world!"
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

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
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
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	logs, err := c.InstanceLogs(ctx, &grpc.InstanceLogsRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	var found bool
	for _, edge := range logs.Edges {
		msg := edge.Node.Msg
		if msg == `"Hello, world!"` {
			found = true
			break
		}
	}

	if !found {
		print(logs)
		return errors.New("missing simple instance log")
	}

	return nil

}

func testStateLogJQ(ctx context.Context, c grpc.FlowClient, namespace string) error {

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
    log: 'jq(.name)!'
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Input:     []byte(`{"name": "Direktiv"}`),
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
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
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	logs, err := c.InstanceLogs(ctx, &grpc.InstanceLogsRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	var found bool
	for _, edge := range logs.Edges {
		msg := edge.Node.Msg
		if msg == `"Direktiv!"` {
			found = true
			break
		}
	}

	if !found {
		print(logs)
		return errors.New("missing jq instance log")
	}

	return nil

}

func testStateLogJQNested(ctx context.Context, c grpc.FlowClient, namespace string) error {

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
    log: 'Hello, jq(.name)!'
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Input:     []byte(`{"name": "Direktiv"}`),
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
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
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	logs, err := c.InstanceLogs(ctx, &grpc.InstanceLogsRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	var found bool
	for _, edge := range logs.Edges {
		msg := edge.Node.Msg
		if msg == `"Hello, Direktiv!"` {
			found = true
			break
		}
	}

	if !found {
		print(logs)
		return errors.New("missing jq instance log")
	}

	return nil

}

func testStateLogJQObject(ctx context.Context, c grpc.FlowClient, namespace string) error {

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
    log: 
      Name: 'jq(.name)'
      Constant: 5
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Input:     []byte(`{"name": "Direktiv"}`),
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
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
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	logs, err := c.InstanceLogs(ctx, &grpc.InstanceLogsRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	var found bool
	for _, edge := range logs.Edges {
		msg := edge.Node.Msg
		if msg == `{
  "Constant": 5,
  "Name": "Direktiv"
}` {
			found = true
			break
		}
	}

	if !found {
		print(logs)
		return errors.New("missing jq instance log")
	}

	return nil

}

func testInstanceSimpleChain(ctx context.Context, c grpc.FlowClient, namespace string) error {

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
    transition: b
  - id: b
    type: noop 
    transition: c
  - id: c
    type: noop 
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Input:     []byte(`{"name": "Direktiv"}`),
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
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
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	if len(iresp.Flow) != 3 || iresp.Flow[0] != "a" || iresp.Flow[1] != "b" || iresp.Flow[2] != "c" {
		return errors.New("instance took unexpected path")
	}

	return nil

}

func testInstanceSwitchLoop(ctx context.Context, c grpc.FlowClient, namespace string) error {

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
        transition: b
        transform: 'jq(.k -= 1)'
    defaultTransition: c
  - id: c
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

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
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
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	if len(iresp.Flow) != 8 || iresp.Flow[0] != "a" || iresp.Flow[1] != "b" || iresp.Flow[2] != "b" ||
		iresp.Flow[3] != "b" || iresp.Flow[4] != "b" || iresp.Flow[5] != "b" ||
		iresp.Flow[6] != "b" || iresp.Flow[7] != "c" {
		return errors.New("instance took unexpected path")
	}

	return nil

}

func testInstanceDelayLoop(ctx context.Context, c grpc.FlowClient, namespace string) error {

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
      currentTime: jq(now)
    transition: b
  - id: b
    type: delay
    duration: PT8S
    transition: c
  - id: c
    type: noop
    transform:
      difference: jq(now - .currentTime)
    transition: d
  - id: d
    type: switch
    conditions:
      - condition: jq(.difference < 8)
        transition: fail
    defaultTransition: e
  - id: e
    type: noop 
  - id: fail
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

	cctx, cancel := context.WithCancel(ctx)
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
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	if len(iresp.Flow) != 5 || iresp.Flow[0] != "a" || iresp.Flow[1] != "b" || iresp.Flow[2] != "c" ||
		iresp.Flow[3] != "d" || iresp.Flow[4] != "e" {
		return errors.New("instance took unexpected path")
	}

	return nil

}

func testInstanceGenerateConsumeEvent(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-consume",
		Source: []byte(`
states:
- id: a
  type: consumeEvent
  transition: b
  timeout: PT10S
  event:
    type: testcloudevent
- id: b
  type: noop
`),
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-generate",
		Source: []byte(`
states:
- id: a
  type: generateEvent
  event:
    type: testcloudevent
    source: Direktiv
    data: 
      message: "helloworld"
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-consume",
	})
	if err != nil {
		return err
	}

	_, err = c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-generate",
	})
	if err != nil {
		return err
	}

	cctx, cancel := context.WithTimeout(ctx, instanceTimeout)
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
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	if len(iresp.Flow) != 2 || iresp.Flow[0] != "a" || iresp.Flow[1] != "b" {
		return errors.New("instance took unexpected path")
	}

	return nil

}
