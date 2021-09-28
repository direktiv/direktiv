package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
)

var testsCmd = &cobra.Command{
	Use: "tests",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		registerTest("CreateNamespace", []string{"namespaces"}, testCreateNamespace)
		registerTest("CreateNamespaceDuplicate", []string{"namespaces", "uniqueness"}, testCreateNamespaceDuplicate)
		registerTest("CreateNamespaceRegex", []string{"namespaces", "regex"}, testCreateNamespaceRegex)
		// TODO: rename namespace
		registerTest("DeleteNamespaceIdempotent", []string{"namespaces"}, testDeleteNamespaceIdempotent)
		registerTest("DeleteNamespaceRecursive", []string{"namespaces"}, testDeleteNamespaceRecursive)
		registerTest("NamespacesStream", []string{"namespaces", "stream"}, testNamespacesStream)
		registerTest("ServerLogs", []string{"namespaces", "logs"}, testServerLogs)
		registerTest("ServerLogsStream", []string{"namespaces", "logs", "stream"}, testServerLogsStream)
		registerTest("NamespaceLogsStreamDisconnect", []string{"namespaces", "logs", "stream"}, testNamespaceLogsStreamDisconnect)

		registerTest("CreateDirectory", []string{"directories"}, testCreateDirectory)
		registerTest("CreateDirectoryDuplicate", []string{"directories", "uniqueness"}, testCreateDirectoryDuplicate)
		registerTest("CreateDirectoryFalseDuplicate", []string{"directories", "uniqueness"}, testCreateDirectoryFalseDuplicate)
		registerTest("CreateDirectoryRoot", []string{"directories", "uniqueness"}, testCreateDirectoryRoot)
		registerTest("CreateDirectoryRegex", []string{"directories", "regex"}, testCreateDirectoryRegex)
		registerTest("CreateDirectoryIdempotent", []string{"directories"}, testCreateDirectoryIdempotent)
		registerTest("CreateDirectoryParents", []string{"directories"}, testCreateDirectoryParents)
		registerTest("CreateDirectoryNoParent", []string{"directories"}, testCreateDirectoryNoParent)
		registerTest("CreateDirectoryNonDirectoryParent", []string{"directories"}, testCreateDirectoryNonDirectoryParent)
		// TODO: rename directory
		registerTest("DeleteDirectory", []string{"directories"}, testDeleteDirectory)
		registerTest("DeleteDirectoryIdempotent", []string{"directories"}, testDeleteDirectoryIdempotent)
		registerTest("DeleteDirectoryRecursive", []string{"directories"}, testDeleteDirectoryRecursive)
		registerTest("DeleteDirectoryRoot", []string{"directories"}, testDeleteDirectoryRoot)
		registerTest("Directory", []string{"directories"}, testDirectory)
		registerTest("DirectoryStream", []string{"directories", "stream"}, testDirectoryStream)
		registerTest("DirectoryStreamDisconnect", []string{"directories", "stream"}, testDirectoryStreamDisconnect)
		registerTest("DirectoryStreamDisconnectParent", []string{"directories", "stream"}, testDirectoryStreamDisconnectParent)
		registerTest("DirectoryStreamDisconnectNamespace", []string{"namespaces", "directories", "stream"}, testDirectoryStreamDisconnectNamespace)

		// TODO: workflow management

		registerTest("StartWorkflow", []string{"instances"}, testStartWorkflow)
		registerTest("StateLogSimple", []string{"instances"}, testStateLogSimple)
		registerTest("StateLogJQ", []string{"instances", "jq"}, testStateLogJQ)
		registerTest("StateLogJQNested", []string{"instances", "jq"}, testStateLogJQNested)
		registerTest("StateLogJQObject", []string{"instances", "jq"}, testStateLogJQObject)
		registerTest("InstanceSimpleChain", []string{"instances"}, testInstanceSimpleChain)
		registerTest("InstanceSwitchLoop", []string{"instances", "jq"}, testInstanceSwitchLoop)

	},
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			runTests(tests)
			return
		}

		tests := getTests(args...)
		if len(tests) == 0 {
			fmt.Println("no results")
			return
		}

		runTests(tests)

	},
}

func getTests(labels ...string) []test {

	var subset = make([]test, 0)

	for _, test := range tests {

		x := test.Labels()

		for _, lbl := range labels {
			for _, y := range x {
				if lbl == y {
					subset = append(subset, test)
					goto breakout
				}
			}
		}

	breakout:
	}

	return subset

}

func runTests(tests []test) {

	ctx := context.Background()

	c, closer, err := client()
	if err != nil {
		exit(err)
	}
	defer closer.Close()

	var total, success, fail int

	for _, test := range tests {

		err = testReset(ctx, c)
		if err != nil {
			exit(err)
		}

		total++

		msg := fmt.Sprintf("Running test %s...", test.Name())
		if len(msg) < 60 {
			msg += strings.Repeat(" ", 60-len(msg))
		}
		fmt.Fprint(os.Stderr, msg)

		err := test.Run(ctx, c)
		if err != nil {
			fail++
			fmt.Fprint(os.Stderr, "FAIL\n")
			fmt.Fprintf(os.Stderr, "\tError: %v\n", err)
		} else {
			success++
			fmt.Fprint(os.Stderr, "SUCCESS\n")
		}

	}

}

func testReset(ctx context.Context, c grpc.FlowClient) error {

	namespaces, err := c.Namespaces(ctx, &grpc.NamespacesRequest{})
	if err != nil {
		return err
	}

	prefix := testNamespace()

	for _, edge := range namespaces.Edges {
		if strings.HasPrefix(edge.Node.Name, prefix) {
			_, err = c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
				Name:       edge.Node.Name,
				Idempotent: true,
				Recursive:  true,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil

}

var tests []test

type test interface {
	Name() string
	Labels() []string
	Run(context.Context, grpc.FlowClient) error
}

type testImpl struct {
	name   string
	labels []string
	run    func(context.Context, grpc.FlowClient) error
}

func (t *testImpl) Name() string {
	return t.name
}

func (t *testImpl) Labels() []string {
	return append(t.labels, t.name)
}

func (t *testImpl) Run(ctx context.Context, c grpc.FlowClient) error {
	return t.run(ctx, c)
}

func registerTest(name string, labels []string, fn func(context.Context, grpc.FlowClient) error) {
	tests = append(tests, &testImpl{
		name:   name,
		labels: labels,
		run:    fn,
	})
}
