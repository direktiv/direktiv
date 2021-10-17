package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/vorteil/direktiv/pkg/flow"

	"github.com/vorteil/direktiv/pkg/flow/grpc"
)

func testInstanceSubflowRetry(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-retry-subflow",
		Source: []byte(`
states:
  - id: a
    type: getter
    transition: b
    transform: 
      newValue: jq(.var.Counter + 1)
    variables:
    - key: Counter
      scope: workflow 
  - id: b
    type: setter
    transition: c
    variables:
    - key: Counter
      scope: workflow 
      value: jq(.newValue)
  - id: c
    type: switch
    conditions:
    - condition: jq(.newValue > 10)
      transition: d
    defaultTransition: e
  - id: d
    type: noop
  - id: e
    type: error
    error: validation.Invalid.Counter
    message: "Counter is less than 10" 
`),
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-retry",
		Source: []byte(`
functions:
  - id: sub
    type: subflow
    workflow: testwf-retry-subflow
states:
  - id: a 
    type: action
    action:
      function: sub
      retries:
        max_attempts: 15
        delay: PT1S
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-retry",
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

	if len(iresp.Flow) != 2 || iresp.Flow[0] != "a" || iresp.Flow[1] != "b" {
		return errors.New("instance took unexpected path")
	}

	// TODO Count Instances

	// instances, err := c.Instances(ctx, &grpc.InstancesRequest{
	// 	Namespace: namespace,
	// })
	// if err != nil {
	// 	return err
	// }

	return nil

}

func testInstanceActionRetry(ctx context.Context, c grpc.FlowClient, namespace string) error {

	// TODO:
	return nil

}

func testInstanceNestedRetry(ctx context.Context, c grpc.FlowClient, namespace string) error {
	// TODO:
	return nil

}

func testInstanceParallelRetry(ctx context.Context, c grpc.FlowClient, namespace string) error {
	// TODO:
	return nil

}
