package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/vorteil/direktiv/pkg/api/client/directory"
	"github.com/vorteil/direktiv/pkg/api/client/instances"
	"github.com/vorteil/direktiv/pkg/api/client/logs"
	"github.com/vorteil/direktiv/pkg/api/client/node"
	"github.com/vorteil/direktiv/pkg/api/client/secrets"
	"github.com/vorteil/direktiv/pkg/api/client/workflows"

	direktivsdk "github.com/vorteil/direktiv/pkg/api/client"
	"github.com/vorteil/direktiv/pkg/api/client/namespaces"
)

func testAPICreateNamespace(ctx context.Context, c direktivsdk.Direktivsdk, namespace string) error {

	resp, aerr := c.Namespaces.CreateNamespace(&namespaces.CreateNamespaceParams{
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*namespaces.CreateNamespaceDefault)
		return errors.New(err.Payload.Error)
	}

	if resp.Payload["namespace"] != nil {
		if resp.Payload["namespace"].(map[string]interface{})["name"] == namespace {
			return nil
		}
	}

	return errors.New("output of create namespace via api returned unexpected results")

}

func testAPICreateNamespaceDuplicate(ctx context.Context, c direktivsdk.Direktivsdk, namespace string) error {
	_, aerr := c.Namespaces.CreateNamespace(&namespaces.CreateNamespaceParams{
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*namespaces.CreateNamespaceDefault)
		return errors.New(err.Payload.Error)
	}

	_, aerr = c.Namespaces.CreateNamespace(&namespaces.CreateNamespaceParams{
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*namespaces.CreateNamespaceDefault)
		if err.Code() != 409 {
			return errors.New(err.Payload.Error)
		}
	}

	return nil
}

func testAPICreateNamespaceRegex(ctx context.Context, c direktivsdk.Direktivsdk, namespace string) error {

	_, aerr := c.Namespaces.CreateNamespace(&namespaces.CreateNamespaceParams{
		Namespace: namespace + "_",
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*namespaces.CreateNamespaceDefault)
		if err.Code() != 406 {
			return errors.New(err.Payload.Error)
		}
	}

	_, aerr = c.Namespaces.CreateNamespace(&namespaces.CreateNamespaceParams{
		Namespace: namespace + "Aa",
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*namespaces.CreateNamespaceDefault)
		if err.Code() != 406 {
			return errors.New(err.Payload.Error)
		}
	}

	return nil
}

func testAPIDeleteNamespace(ctx context.Context, c direktivsdk.Direktivsdk, namespace string) error {
	_, aerr := c.Namespaces.CreateNamespace(&namespaces.CreateNamespaceParams{
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*namespaces.CreateNamespaceDefault)
		return errors.New(err.Payload.Error)
	}
	_, aerr = c.Namespaces.DeleteNamespace(&namespaces.DeleteNamespaceParams{
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*namespaces.DeleteNamespaceDefault)
		return errors.New(err.Payload.Error)
	}

	return nil
}

func testAPIServerLogs(ctx context.Context, c direktivsdk.Direktivsdk, namespace string) error {
	_, aerr := c.Namespaces.CreateNamespace(&namespaces.CreateNamespaceParams{
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*namespaces.CreateNamespaceDefault)
		return errors.New(err.Payload.Error)
	}

	resp, aerr := c.Logs.ServerLogs(&logs.ServerLogsParams{
		Context: ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*logs.ServerLogsDefault)
		return errors.New(err.Payload.Error)
	}

	logs := resp.Payload["edges"].([]interface{})

	var k int

	for _, edge := range logs {
		// type asserting not the greatest fan but generic responses lead to this
		obj := edge.(map[string]interface{})
		node := obj["node"].(map[string]interface{})
		msg := node["msg"].(string)
		if strings.Contains(msg, namespace) {
			k++
		}
	}

	if k == 0 {
		return fmt.Errorf("server logs contain no record of recently created namespace")
	}

	return nil
}

func testAPICreateDirectoryDeleteDirectory(ctx context.Context, c direktivsdk.Direktivsdk, namespace string) error {
	_, aerr := c.Namespaces.CreateNamespace(&namespaces.CreateNamespaceParams{
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*namespaces.CreateNamespaceDefault)
		return errors.New(err.Payload.Error)
	}

	cd := directory.NewCreateDirectoryParams().WithDefaults()
	cd.SetContext(ctx)
	cd.SetNamespace(namespace)
	cd.SetDirectory("testdir")

	resp, err := c.Directory.CreateDirectory(cd, nil)
	if err != nil {
		err := aerr.(*directory.CreateDirectoryDefault)
		return errors.New(err.Payload.Error)
	}

	nodeCheck := resp.Payload["node"].(map[string]interface{})
	if nodeCheck["name"] != "testdir" {
		return errors.New("directory 'testdir' was never created")
	}

	dn := node.NewDeleteNodeParams().WithDefaults()
	dn.SetContext(ctx)
	dn.SetNamespace(namespace)
	dn.SetNode("testdir")

	_, err = c.Node.DeleteNode(dn, nil)
	if err != nil {
		err := aerr.(*node.DeleteNodeDefault)
		return errors.New(err.Payload.Error)
	}

	// if no error no need check to resp as it doesn't respond with anything

	return nil
}

func testAPIDirectory(ctx context.Context, c direktivsdk.Direktivsdk, namespace string) error {
	_, aerr := c.Namespaces.CreateNamespace(&namespaces.CreateNamespaceParams{
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*namespaces.CreateNamespaceDefault)
		return errors.New(err.Payload.Error)
	}

	gd := node.NewGetNodesParams().WithDefaults()
	gd.SetNamespace(namespace)
	gd.SetNodePath("testdir")

	_, aerr = c.Node.GetNodes(gd, nil)
	if aerr == nil {
		return errors.New("server returned directory that shouldn't exist")
	}

	berr := aerr.(*node.GetNodesDefault)
	if berr.Code() != 404 {
		return fmt.Errorf("incorrect error from server: %s", berr.Payload.Error)
	}

	cd := directory.NewCreateDirectoryParams().WithDefaults()
	cd.SetContext(ctx)
	cd.SetNamespace(namespace)
	cd.SetDirectory("testdir")

	resp, err := c.Directory.CreateDirectory(cd, nil)
	if err != nil {
		err := aerr.(*directory.CreateDirectoryDefault)
		return errors.New(err.Payload.Error)
	}

	nodeCheck := resp.Payload["node"].(map[string]interface{})
	name := nodeCheck["name"].(string)
	if name != "testdir" {
		return errors.New("create directory did not create test dir")
	}

	dn := node.NewDeleteNodeParams().WithDefaults()
	dn.SetContext(ctx)
	dn.SetNamespace(namespace)
	dn.SetNode("testdir")

	_, err = c.Node.DeleteNode(dn, nil)
	if err != nil {
		err := aerr.(*node.DeleteNodeDefault)
		return errors.New(err.Payload.Error)
	}

	return nil
}

func testAPISecrets(ctx context.Context, c direktivsdk.Direktivsdk, namespace string) error {
	_, aerr := c.Namespaces.CreateNamespace(&namespaces.CreateNamespaceParams{
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*namespaces.CreateNamespaceDefault)
		return errors.New(err.Payload.Error)
	}

	resp, aerr := c.Secrets.GetSecrets(&secrets.GetSecretsParams{
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*secrets.GetSecretsDefault)
		return errors.New(err.Payload.Error)
	}

	secretst := resp.Payload["secrets"]
	edges := secretst.(map[string]interface{})["edges"].([]interface{})

	if len(edges) != 0 {
		return errors.New("unexpected secrets already exist in the namespace")
	}

	_, aerr = c.Secrets.CreateSecret(&secrets.CreateSecretParams{
		Namespace:     namespace,
		Context:       ctx,
		SecretPayload: "MySecret2",
		Secret:        "testSecret",
	}, nil)
	if aerr != nil {
		err := aerr.(*secrets.CreateSecretDefault)
		return errors.New(err.Payload.Error)
	}

	resp, aerr = c.Secrets.GetSecrets(&secrets.GetSecretsParams{
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*secrets.GetSecretsDefault)
		return errors.New(err.Payload.Error)
	}

	secretst = resp.Payload["secrets"]
	edges = secretst.(map[string]interface{})["edges"].([]interface{})

	if len(edges) != 1 {
		return errors.New("incorrect number of secrets returned by server")
	}

	_, aerr = c.Secrets.CreateSecret(&secrets.CreateSecretParams{
		Namespace:     namespace,
		Context:       ctx,
		SecretPayload: "MySecret2",
		Secret:        "testSecret",
	}, nil)
	if aerr == nil {
		return errors.New("server accepted duplicate secret without error")
	}

	err := aerr.(*secrets.CreateSecretDefault)
	if err.Code() != 409 {
		return fmt.Errorf("incorrect error from server: %w", err)
	}

	_, aerr = c.Secrets.DeleteSecret(&secrets.DeleteSecretParams{
		Namespace: namespace,
		Context:   ctx,
		Secret:    "testSecret",
	}, nil)
	if aerr != nil {
		err := aerr.(*secrets.DeleteSecretDefault)
		return errors.New(err.Payload.Error)
	}

	resp, aerr = c.Secrets.GetSecrets(&secrets.GetSecretsParams{
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*secrets.GetSecretsDefault)
		return errors.New(err.Payload.Error)
	}

	secretst = resp.Payload["secrets"]
	edges = secretst.(map[string]interface{})["edges"].([]interface{})

	if len(edges) != 0 {
		return errors.New("unexpected secrets still exist in the namespace")
	}

	return nil
}

func testAPIWorkflow(ctx context.Context, c direktivsdk.Direktivsdk, namespace string) error {
	_, aerr := c.Namespaces.CreateNamespace(&namespaces.CreateNamespaceParams{
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*namespaces.CreateNamespaceDefault)
		return errors.New(err.Payload.Error)
	}

	// create a workflow
	cw := workflows.NewCreateWorkflowParams().WithDefaults()
	cw.SetNamespace(namespace)
	cw.SetContext(ctx)
	cw.SetWorkflow("test")
	cw.SetWorkflowData(`description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!
`)

	_, aerr = c.Workflows.CreateWorkflow(cw, nil)
	if aerr != nil {
		err := aerr.(*workflows.CreateWorkflowDefault)
		return errors.New(err.Payload.Error)
	}

	// execute a workflow
	ew := workflows.NewExecuteWorkflowParams().WithDefaults()
	ew.SetContext(ctx)
	ew.SetNamespace(namespace)
	ew.SetWorkflow("test")

	resp, aerr := c.Workflows.ExecuteWorkflow(ew, nil)
	if aerr != nil {
		err := aerr.(*workflows.ExecuteWorkflowDefault)
		return errors.New(err.Payload.Error)
	}

	instance := resp.Payload["instance"].(string)

	respInstance, aerr := c.Instances.GetInstance(&instances.GetInstanceParams{
		Namespace: namespace,
		Instance:  instance,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*instances.GetInstanceDefault)
		return errors.New(err.Payload.Error)
	}
	iresp := respInstance.Payload["instance"].(map[string]interface{})
	if iresp["id"] != instance {
		return errors.New("wrong instance was returned")
	}

	respInstanceLogs, aerr := c.Logs.InstanceLogs(&logs.InstanceLogsParams{
		Instance:  instance,
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*logs.InstanceLogsDefault)
		return errors.New(err.Payload.Error)
	}

	logsArr := respInstanceLogs.Payload["edges"].([]interface{})

	if len(logsArr) == 0 {
		return errors.New("logs should have been returned for completed instance")
	}

	// delete a workflow
	dw := node.NewDeleteNodeParams().WithDefaults()
	dw.SetNamespace(namespace)
	dw.SetContext(ctx)
	dw.SetNode("test")

	_, aerr = c.Node.DeleteNode(dw, nil)
	if aerr != nil {
		err := aerr.(*node.DeleteNodeDefault)
		return errors.New(err.Payload.Error)
	}
	return nil

}
