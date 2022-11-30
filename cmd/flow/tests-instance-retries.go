package main

/*

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
)

func testInstanceLongRetry(ctx context.Context, c grpc.FlowClient, namespace string) error {

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
    - condition: jq(.newValue > 1)
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
        max_attempts: 5
        delay: PT8S
        codes:
        - "validation.Invalid.Counter"
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

		if iresp.Instance.Status != util.InstanceStatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != util.InstanceStatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	return nil

}

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
        codes:
        - "validation.Invalid.Counter"
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

		if iresp.Instance.Status != util.InstanceStatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != util.InstanceStatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	return nil

}

func testInstanceActionRetry(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	counterUUID := uuid.New()
	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-retry-action",
		Source: []byte(fmt.Sprintf(`
functions:
- id: counter
  image: jkizo/persistent-counter:v1
  type: reusable
states:
- id: a
  type: action
  action:
    function: counter
    retries:
        max_attempts: 12
        delay: PT1S
        codes:
          - "com.invalid-value.error"
    input:
      uuid: "%s"
      min: 10
`, counterUUID)),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-retry-action",
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

		if iresp.Instance.Status != util.InstanceStatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != util.InstanceStatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	return nil

}

func testInstanceNestedRetry(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	counterUUID := uuid.New()
	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf-retry-action",
		Source: []byte(fmt.Sprintf(`
functions:
- id: counter
  image: jkizo/persistent-counter:v1
  type: reusable
states:
- id: a
  type: action
  action:
    function: counter
    retries:
        max_attempts: 5
        delay: PT1S
        codes:
          - "com.invalid-value.error"
    input:
      uuid: "%s"
      min: 25
`, counterUUID)),
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
    workflow: testwf-retry-action
states:
  - id: a
    type: action
    action:
      function: sub
      retries:
        max_attempts: 5
        delay: PT1S
        codes:
        - "direktiv.retries.exceeded"
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

		if iresp.Instance.Status != util.InstanceStatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != util.InstanceStatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	return nil

}

func testInstanceParallelRetry(ctx context.Context, c grpc.FlowClient, namespace string) error {
	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	counterUUID1 := uuid.New()
	counterUUID2 := uuid.New()
	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/testwf",
		Source: []byte(fmt.Sprintf(`
functions:
- id: counter
  image: jkizo/persistent-counter:v1
  type: reusable
states:
- id: runpara
  type: parallel
  actions:
  - function: counter
    retries:
      max_attempts: 8
      delay: PT1S
      codes:
        - "com.invalid-value.error"
    input:
      uuid: "%s"
      min: 3
  - function: counter
    retries:
      max_attempts: 8
      delay: PT1S
      codes:
        - "com.invalid-value.error"
    input:
      uuid: "%s"
      min: 6
  mode: and
`, counterUUID1, counterUUID2)),
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

		if iresp.Instance.Status != util.InstanceStatusPending {
			break
		}
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	if iresp.Instance.Status != util.InstanceStatusComplete {
		return fmt.Errorf("instance failed: %s : %s", iresp.Instance.ErrorCode, iresp.Instance.ErrorMessage)
	}

	return nil
}

*/
