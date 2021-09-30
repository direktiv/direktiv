package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vorteil/direktiv/pkg/flow"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func testSecretsAPI(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	secrets, err := c.Secrets(ctx, &grpc.SecretsRequest{
		Namespace: namespace,
	})
	if err != nil {
		return err
	}

	if len(secrets.Secrets.Edges) != 0 {
		return errors.New("unexpected secrets already exist in the namespace")
	}

	client, err := c.SecretsStream(ctx, &grpc.SecretsRequest{
		Namespace: namespace,
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	secrets, err = client.Recv()
	if err != nil {
		return err
	}
	if len(secrets.Secrets.Edges) != 0 {
		return errors.New("unexpected secrets already exist in the namespace")
	}

	_, err = c.SetSecret(ctx, &grpc.SetSecretRequest{
		Namespace: namespace,
		Key:       "testSecret",
		Data:      []byte("MySecret"),
	})
	if err != nil {
		fmt.Println("++++++++++++++++++ FAIL", namespace)
		return err
	}

	secrets, err = client.Recv()
	if err != nil {
		return err
	}
	if len(secrets.Secrets.Edges) != 1 {
		return errors.New("incorrect number of secrets returned by server")
	}

	_, err = c.SetSecret(ctx, &grpc.SetSecretRequest{
		Namespace: namespace,
		Key:       "testSecret",
		Data:      []byte("MySecret2"),
	})
	if err == nil {
		return errors.New("server accepted duplicate secret without error")
	}
	if status.Code(err) != codes.AlreadyExists {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	_, err = c.DeleteSecret(ctx, &grpc.DeleteSecretRequest{
		Namespace: namespace,
		Key:       "testSecret",
	})
	if err != nil {
		return err
	}

	secrets, err = client.Recv()
	if err != nil {
		return err
	}
	if len(secrets.Secrets.Edges) != 0 {
		return errors.New("unexpected secrets still exist in the namespace")
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	return nil

}

func testInstanceSubflowSecrets(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.SetSecret(ctx, &grpc.SetSecretRequest{
		Namespace: namespace,
		Key:       "testSecret",
		Data:      []byte("MySecret"),
	})
	if err != nil {
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>> FAIL", namespace)
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/test-subflow",
		Source: []byte(`
states:
  - id: a
    type: noop
`),
	})
	if err != nil {
		return err
	}

	_, err = c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
		Namespace: namespace,
		Path:      "/test-superflow",
		Source: []byte(`
functions:
  - id: sf
    type: subflow 
    workflow: '/test-subflow'
states:
  - id: a
    type: action
    action: 
      function: sf
      secrets: ['testSecret']
`),
	})
	if err != nil {
		return err
	}

	resp, err := c.StartWorkflow(ctx, &grpc.StartWorkflowRequest{
		Namespace: namespace,
		Path:      "/test-superflow",
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

	output, err := c.InstanceOutput(ctx, &grpc.InstanceOutputRequest{
		Namespace: namespace,
		Instance:  resp.Instance,
	})
	if err != nil {
		return err
	}

	if string(output.Data) != `{"input":"","return":{"input":"","secrets":{"testSecret":"MySecret"}}}` {
		return errors.New("unexpected instance output")
	}

	return nil

}
