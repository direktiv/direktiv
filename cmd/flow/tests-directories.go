package flow

/*

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func testCreateDirectory(ctx context.Context, c grpc.FlowClient, namespace string) error {

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

	return nil

}

func testCreateDirectoryDuplicate(ctx context.Context, c grpc.FlowClient, namespace string) error {

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

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir",
	})
	if err == nil {
		return errors.New("server accepted duplicate directory without error")
	}
	if status.Code(err) != codes.AlreadyExists {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "testdir",
	})
	if err == nil {
		return errors.New("server accepted duplicate directory without error")
	}
	if status.Code(err) != codes.AlreadyExists {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	return nil

}

func testCreateDirectoryFalseDuplicate(ctx context.Context, c grpc.FlowClient, namespace string) error {

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

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir/testdir",
	})
	if err != nil {
		return err
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir/testdir/testdir",
	})
	if err != nil {
		return err
	}

	return nil

}

func testCreateDirectoryRoot(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/",
	})
	if err == nil {
		return errors.New("server accepted duplicate directory without error")
	}
	if status.Code(err) != codes.AlreadyExists {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "",
	})
	if err == nil {
		return errors.New("server accepted duplicate directory without error")
	}
	if status.Code(err) != codes.AlreadyExists {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	return nil

}

func testCreateDirectoryRegex(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/-asd",
	})
	if err == nil {
		return errors.New("server accepted bad directory name without error")
	}
	if status.Code(err) != codes.InvalidArgument {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	return nil

}

func testCreateDirectoryIdempotent(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace:  namespace,
		Path:       "/testdir",
		Idempotent: true,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace:  namespace,
		Path:       "/testdir",
		Idempotent: true,
	})
	if err != nil {
		return err
	}

	return nil

}

func testCreateDirectoryParents(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir/a/b/c",
		Parents:   true,
	})
	if err != nil {
		return err
	}

	_, err = c.Directory(ctx, &grpc.DirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir",
	})
	if err != nil {
		return err
	}

	_, err = c.Directory(ctx, &grpc.DirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir/a",
	})
	if err != nil {
		return err
	}

	_, err = c.Directory(ctx, &grpc.DirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir/a/b",
	})
	if err != nil {
		return err
	}

	_, err = c.Directory(ctx, &grpc.DirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir/a/b/c",
	})
	if err != nil {
		return err
	}

	return nil

}

func testCreateDirectoryNoParent(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir/a/b/c",
	})
	if err == nil {
		return errors.New("server accepted directory with invalid parent without error")
	}
	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	return nil

}

func testCreateDirectoryNonDirectoryParent(ctx context.Context, c grpc.FlowClient, namespace string) error {

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

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/testwf",
	})
	if err == nil {
		return errors.New("server accepted directory clashing with workflow without error")
	}
	if status.Code(err) != codes.AlreadyExists {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/testwf/testdir",
	})
	if err == nil {
		return errors.New("server accepted directory under workflow without error")
	}
	if status.Code(err) != codes.AlreadyExists {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	return nil

}

func testDeleteDirectory(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.DeleteNode(ctx, &grpc.DeleteNodeRequest{
		Namespace: namespace,
		Path:      "/testdir",
	})
	if err == nil {
		return errors.New("server accepted delete directory that doesn't exist without error")
	}
	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir",
	})
	if err != nil {
		return err
	}

	_, err = c.DeleteNode(ctx, &grpc.DeleteNodeRequest{
		Namespace: namespace,
		Path:      "/testdir",
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

	_, err = c.DeleteNode(ctx, &grpc.DeleteNodeRequest{
		Namespace: namespace,
		Path:      "testdir",
	})
	if err != nil {
		return err
	}

	return nil

}

func testDeleteDirectoryIdempotent(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.DeleteNode(ctx, &grpc.DeleteNodeRequest{
		Namespace:  namespace,
		Path:       "/testdir",
		Idempotent: true,
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

	_, err = c.DeleteNode(ctx, &grpc.DeleteNodeRequest{
		Namespace:  namespace,
		Path:       "/testdir",
		Idempotent: true,
	})
	if err != nil {
		return err
	}

	_, err = c.DeleteNode(ctx, &grpc.DeleteNodeRequest{
		Namespace:  namespace,
		Path:       "/testdir",
		Idempotent: true,
	})
	if err != nil {
		return err
	}

	return nil

}

func testDeleteDirectoryRecursive(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir/a/b/c",
		Parents:   true,
	})
	if err != nil {
		return err
	}

	_, err = c.DeleteNode(ctx, &grpc.DeleteNodeRequest{
		Namespace: namespace,
		Path:      "/testdir",
	})
	if err == nil {
		return errors.New("server accepted delete non-empty directory without recursive flag without error")
	}
	if status.Code(err) != codes.InvalidArgument {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	_, err = c.DeleteNode(ctx, &grpc.DeleteNodeRequest{
		Namespace: namespace,
		Path:      "/testdir",
		Recursive: true,
	})
	if err != nil {
		return err
	}

	return nil

}

func testDeleteDirectoryRoot(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.DeleteNode(ctx, &grpc.DeleteNodeRequest{
		Namespace: namespace,
		Path:      "/",
	})
	if err == nil {
		return errors.New("server accepted delete root directory without error")
	}
	if status.Code(err) != codes.InvalidArgument {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	_, err = c.DeleteNode(ctx, &grpc.DeleteNodeRequest{
		Namespace: namespace,
		Path:      "",
	})
	if err == nil {
		return errors.New("server accepted delete root directory without error")
	}
	if status.Code(err) != codes.InvalidArgument {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	return nil

}

func compareDirectoryResponse(expect, actual *grpc.DirectoryResponse) error {

	if actual.Namespace != expect.Namespace {
		return fmt.Errorf("unexpected directory response namespace")
	}

	if actual.Node.Name != expect.Node.Name {
		return fmt.Errorf("unexpected directory response node name")
	}

	if actual.Node.Path != expect.Node.Path {
		return fmt.Errorf("unexpected directory response node path")
	}

	if actual.Node.Parent != expect.Node.Parent {
		return fmt.Errorf("unexpected directory response node parent")
	}

	return nil

}

func testDirectory(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	_, err = c.Directory(ctx, &grpc.DirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir",
	})
	if err == nil {
		return errors.New("server returned directory that shouldn't exist")
	}
	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir",
	})
	if err != nil {
		return err
	}

	expect := &grpc.DirectoryResponse{
		Namespace: namespace,
		Node: &grpc.Node{
			Name:   "testdir",
			Path:   "/testdir",
			Parent: "/",
		},
	}

	resp, err := c.Directory(ctx, &grpc.DirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir",
	})
	if err != nil {
		return err
	}

	err = compareDirectoryResponse(expect, resp)
	if err != nil {
		return err
	}

	resp, err = c.Directory(ctx, &grpc.DirectoryRequest{
		Namespace: namespace,
		Path:      "testdir",
	})
	if err != nil {
		return err
	}

	err = compareDirectoryResponse(expect, resp)
	if err != nil {
		return err
	}

	resp, err = c.Directory(ctx, &grpc.DirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir/",
	})
	if err != nil {
		return err
	}

	err = compareDirectoryResponse(expect, resp)
	if err != nil {
		return err
	}

	resp, err = c.Directory(ctx, &grpc.DirectoryRequest{
		Namespace: namespace,
		Path:      "testdir/",
	})
	if err != nil {
		return err
	}

	err = compareDirectoryResponse(expect, resp)
	if err != nil {
		return err
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "testdir/a",
	})
	if err != nil {
		return err
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "testdir/b",
	})
	if err != nil {
		return err
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "testdir/c",
	})
	if err != nil {
		return err
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "testdir/c/d",
	})
	if err != nil {
		return err
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "testdi",
	})
	if err != nil {
		return err
	}

	resp, err = c.Directory(ctx, &grpc.DirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir",
	})
	if err != nil {
		return err
	}

	if len(resp.Children.Results) != 3 {
		return errors.New("incorrect number of directory children")
	}

	if resp.Children.Results[0].Name != "a" ||
		resp.Children.Results[1].Name != "b" ||
		resp.Children.Results[2].Name != "c" {
		return errors.New("incorrect directory children")
	}

	return nil

}

func testDirectoryStream(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
		Name: namespace,
	})
	if err != nil {
		return err
	}

	client, err := c.DirectoryStream(ctx, &grpc.DirectoryRequest{
		Namespace: namespace,
		Path:      "/",
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	resp, err := client.Recv()
	if err != nil {
		return err
	}

	if len(resp.Children.Results) != 0 {
		return errors.New("unexpected nodes in test directory")
	}

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir",
	})
	if err != nil {
		return err
	}

	resp, err = client.Recv()
	if err != nil {
		return err
	}

	if len(resp.Children.Results) != 1 {
		return errors.New("incorrect data in test directory")
	}

	_, err = c.DeleteNode(ctx, &grpc.DeleteNodeRequest{
		Namespace: namespace,
		Path:      "/testdir",
		Recursive: true,
	})
	if err != nil {
		return err
	}

	resp, err = client.Recv()
	if len(resp.Children.Results) != 0 {
		return errors.New("unexpected nodes in test directory")
	}

	err = client.CloseSend()
	if err != nil {
		return err
	}

	return nil

}

func testDirectoryStreamDisconnect(ctx context.Context, c grpc.FlowClient, namespace string) error {

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

	client, err := c.DirectoryStream(ctx, &grpc.DirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir",
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	_, err = client.Recv()
	if err != nil {
		return err
	}

	_, err = c.DeleteNode(ctx, &grpc.DeleteNodeRequest{
		Namespace: namespace,
		Path:      "/testdir",
		Recursive: true,
	})
	if err != nil {
		return err
	}

	_, err = client.Recv()
	if err == nil {
		return fmt.Errorf("directory stream failed to disconnect when directory was deleted")
	}
	if err != io.EOF {
		return err
	}

	return nil

}

func testDirectoryStreamDisconnectParent(ctx context.Context, c grpc.FlowClient, namespace string) error {

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

	_, err = c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir/a",
	})
	if err != nil {
		return err
	}

	client, err := c.DirectoryStream(ctx, &grpc.DirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir/a",
	})
	if err != nil {
		return err
	}
	defer client.CloseSend()

	_, err = client.Recv()
	if err != nil {
		return err
	}

	_, err = c.DeleteNode(ctx, &grpc.DeleteNodeRequest{
		Namespace: namespace,
		Path:      "/testdir",
		Recursive: true,
	})
	if err != nil {
		return err
	}

	_, err = client.Recv()
	if err == nil {
		return fmt.Errorf("directory stream failed to disconnect when directory parent was deleted")
	}
	if err != io.EOF {
		return err
	}

	return nil

}

func testDirectoryStreamDisconnectNamespace(ctx context.Context, c grpc.FlowClient, namespace string) error {

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

	client, err := c.DirectoryStream(ctx, &grpc.DirectoryRequest{
		Namespace: namespace,
		Path:      "/testdir",
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
		Name:      namespace,
		Recursive: true,
	})
	if err != nil {
		return err
	}

	_, err = client.Recv()
	if err == nil {
		return fmt.Errorf("directory stream failed to disconnect when namespace was deleted")
	}
	if err != io.EOF {
		return err
	}

	return nil

}

*/
