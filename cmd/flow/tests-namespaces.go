package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func testCreateNamespace(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	return nil

}

func testCreateNamespaceDuplicate(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err == nil {
		return errors.New("server accepted duplicate namespace without error")
	}
	if status.Code(err) != codes.AlreadyExists {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	_, err = c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name:       namespace,
		Idempotent: true,
	})
	if err != nil {
		return err
	}

	resp, err := c.Namespaces(ctx, &grpc.NamespacesRequest{
		Pagination: &grpc.Pagination{
			Filter: &grpc.PageFilter{
				Field: "NAME",
				Type:  "CONTAINS",
				Val:   namespace,
			},
		},
	})
	if err != nil {
		return err
	}

	var total int

	for _, edge := range resp.Edges {
		if edge.Node.Name == namespace {
			total++
		}
	}

	if total > 1 {
		return fmt.Errorf("duplicate namespace on server")
	}
	if total == 0 {
		return fmt.Errorf("namespace page filtering failed")
	}

	_, err = c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	return nil

}

func testCreateNamespaceRegex(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace + "/ns",
	})
	if err == nil {
		return errors.New("server accepted bad namespace name without error")
	}
	if status.Code(err) != codes.InvalidArgument {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	_, err = c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace + "_",
	})
	if err == nil {
		return errors.New("server accepted bad namespace name without error")
	}
	if status.Code(err) != codes.InvalidArgument {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	_, err = c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace + "Aa",
	})
	if err == nil {
		return errors.New("server accepted bad namespace name without error")
	}
	if status.Code(err) != codes.InvalidArgument {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	return nil

}

func testDeleteNamespaceIdempotent(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
		Name:       namespace,
		Idempotent: true,
	})
	if err != nil {
		return err
	}

	_, err = c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
		Name:       namespace,
		Idempotent: true,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
		Name:       namespace,
		Idempotent: true,
	})
	if err != nil {
		return err
	}

	_, err = c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
		Name:       namespace,
		Idempotent: true,
	})
	if err != nil {
		return err
	}

	_, err = c.Namespace(ctx, &grpc.NamespaceRequest{
		Name: namespace,
	})
	if err == nil {
		return fmt.Errorf("server still have a namespace that should have been deleted")
	}
	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	return nil

}

func testDeleteNamespaceRecursive(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir",
	})
	if err != nil {
		return err
	}

	_, err = c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
		Name: namespace,
	})
	if err == nil {
		return fmt.Errorf("server deleted a populated namespace without requiring the recursive flag")
	}

	_, err = c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
		Name:      namespace,
		Recursive: true,
	})
	if err != nil {
		return err
	}

	_, err = c.Namespace(ctx, &grpc.NamespaceRequest{
		Name: namespace,
	})
	if err == nil {
		return fmt.Errorf("server still have a namespace that should have been deleted")
	}
	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	return nil

}

func testNamespacesStream(ctx context.Context, c grpc.FlowClient, namespace string) error {

	client, err := c.NamespacesStream(ctx, &grpc.NamespacesRequest{
		Pagination: &grpc.Pagination{
			Filter: &grpc.PageFilter{
				Field: "NAME",
				Type:  "CONTAINS",
				Val:   namespace,
			},
		},
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	resp, err := client.Recv()
	if err != nil {
		return err
	}
	if len(resp.Edges) > 0 {
		return fmt.Errorf("unexpected namespaces stream results returned by server")
	}

	_, err = c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	resp, err = client.Recv()
	if err != nil {
		return err
	}
	if len(resp.Edges) != 1 {
		return fmt.Errorf("unexpected namespaces stream results returned by server")
	}

	_, err = c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	resp, err = client.Recv()
	if err != nil {
		return err
	}
	if len(resp.Edges) > 0 {
		return fmt.Errorf("unexpected namespaces stream results returned by server")
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	return nil

}

func testServerLogs(ctx context.Context, c grpc.FlowClient, namespace string) error {

	name := namespace + "tsl"

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: name,
	})
	if err != nil {
		return err
	}

	resp, err := c.ServerLogs(ctx, &grpc.ServerLogsRequest{
		Pagination: &grpc.Pagination{
			Last: 10,
		},
	})
	if err != nil {
		return err
	}

	var k int

	for _, edge := range resp.Edges {
		if strings.Contains(edge.Node.Msg, name) {
			k++
		}
	}

	if k == 0 {
		return fmt.Errorf("server logs contain no record of recently created namespace")
	}

	_, err = c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
		Name: name,
	})
	if err != nil {
		return err
	}

	return nil

}

func testServerLogsStream(ctx context.Context, c grpc.FlowClient, namespace string) error {

	client, err := c.ServerLogsParcels(ctx, &grpc.ServerLogsRequest{
		Pagination: &grpc.Pagination{
			Last: 10,
		},
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	_, err = client.Recv()
	if err != nil {
		return err
	}

	name := namespace + "tssl"

	_, err = c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: name,
	})
	if err != nil {
		return err
	}

	deadline := time.Now().Add(time.Second*5)

	for {

		resp, err := client.Recv()
		if err != nil {
			return err
		}

		var k int

		for _, edge := range resp.Edges {
			if strings.Contains(edge.Node.Msg, name) {
				k++
			}
		}

		if k != 0 {
			break
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("server logs stream contains no record of recently created namespace")
		}

	}

	_, err = c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
		Name: name,
	})
	if err != nil {
		return err
	}

	deadline = time.Now().Add(time.Second*5)

	for {
		resp, err := client.Recv()
		if err != nil {
			return err
		}

		k := 0

		for _, edge := range resp.Edges {
			if strings.Contains(edge.Node.Msg, name) {
				k++
			}
		}

		if k != 0 {
			break
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("server logs stream contains no record of recently deleted namespace")
		}

	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	return nil

}

func testNamespaceLogsStreamDisconnect(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	client, err := c.NamespaceLogsParcels(ctx, &grpc.NamespaceLogsRequest{
		Namespace: namespace,
		Pagination: &grpc.Pagination{
			Last: 10,
		},
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	_, err = client.Recv()
	if err != nil {
		return err
	}

	_, err = c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = client.Recv()
	if err == nil {
		return fmt.Errorf("namespace logs stream failed to disconnect when namespace was deleted")
	}
	if err != io.EOF {
		return err
	}

	return nil

}
