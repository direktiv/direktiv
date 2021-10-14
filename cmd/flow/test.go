package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
)

var skipLongTests bool
var parallelTests int
var instanceTimeout time.Duration
var testTimeout time.Duration

var testsCmd = &cobra.Command{
	Use: "tests",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		registerTest("CreateNamespace", []string{"namespaces"}, testCreateNamespace)
		registerTest("CreateNamespaceDuplicate", []string{"namespaces", "uniqueness"}, testCreateNamespaceDuplicate)
		registerTest("CreateNamespaceRegex", []string{"namespaces", "regex"}, testCreateNamespaceRegex)
		// TODO: rename namespace
		registerTest("DeleteNamespaceIdempotent", []string{"namespaces"}, testDeleteNamespaceIdempotent)
		registerTest("DeleteNamespaceRecursive", []string{"namespaces"}, testDeleteNamespaceRecursive)
		registerTest("NamespacesStream", []string{"namespaces", "stream", "race"}, testNamespacesStream)
		registerTest("ServerLogs", []string{"namespaces", "logs", "race"}, testServerLogs)
		registerTest("ServerLogsStream", []string{"namespaces", "logs", "stream", "race"}, testServerLogsStream)
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
		registerTest("SecretsAPI", []string{"secrets"}, testSecretsAPI)
		registerTest("StartWorkflow", []string{"instances"}, testStartWorkflow)
		registerTest("StateLogSimple", []string{"instances"}, testStateLogSimple)
		registerTest("StateLogJQ", []string{"instances", "jq"}, testStateLogJQ)
		registerTest("StateLogJQNested", []string{"instances", "jq"}, testStateLogJQNested)
		registerTest("StateLogJQObject", []string{"instances", "jq"}, testStateLogJQObject)
		registerTest("InstanceSimpleChain", []string{"instances"}, testInstanceSimpleChain)
		registerTest("InstanceSwitchLoop", []string{"instances", "jq"}, testInstanceSwitchLoop)
		registerTest("InstanceDelayLoop", []string{"instances", "jq", "long"}, testInstanceDelayLoop)
		registerTest("InstanceSubflowSecrets", []string{"instances", "jq", "secrets", "actions", "subflows"}, testInstanceSubflowSecrets)
		registerTest("NamespaceVariablesSmall", []string{"variables"}, testNamespaceVariablesSmall)
		registerTest("NamespaceVariablesLarge", []string{"variables", "long"}, testNamespaceVariablesLarge)
		registerTest("WorkflowVariablesSmall", []string{"variables"}, testWorkflowVariablesSmall)
		registerTest("WorkflowVariablesLarge", []string{"variables", "long"}, testWorkflowVariablesLarge)
		registerTest("InstanceNamespaceVariables", []string{"instances", "jq", "variables"}, testInstanceNamespaceVariables)
		registerTest("InstanceWorkflowVariables", []string{"instances", "jq", "variables"}, testInstanceWorkflowVariables)
		registerTest("InstanceInstanceVariables", []string{"instances", "jq", "variables"}, testInstanceInstanceVariables)

		// TODO:
		/*
			Error State
			ValidateState
			Foreach State
			Parallel State
			CloudEvents

			Delay State
			Crons

			Action Types
				Global
				Namespace
				Reusable
				Isolate

			Workflow Management Tests (revisions, tags, routers, etc)
		*/

	},
	Run: func(cmd *cobra.Command, args []string) {

		tests := getTests(args...)
		if len(tests) == 0 {
			fmt.Println("no results")
			return
		}

		runTestsParallel(tests, parallelTests)

	},
}

func getTests(labels ...string) []test {

	var subset = make([]test, 0)

	for _, test := range tests {

		x := test.Labels()

		var take bool
		var long bool

		for _, y := range x {
			if y == "long" {
				long = true
			}
		}

		for _, lbl := range labels {
			for _, y := range x {
				if lbl == y {
					take = true
				}
			}
		}

		if (take || len(labels) == 0) && (!skipLongTests || !long) {
			subset = append(subset, test)
		}

	}

	return subset

}

func runTestsParallel(tests []test, c int) {

	testsFullReset()
	defer testsFullReset()

	if c == 1 {
		err := runTests(tests, true, 0)
		if err != nil {
			os.Exit(1)
		}
		return
	}

	var wg sync.WaitGroup
	wg.Add(c)

	var err error
	var lock sync.Mutex

	for i := 0; i < c; i++ {
		go func(i int) {
			defer wg.Done()
			e := runTests(tests, false, i)
			lock.Lock()
			if err != nil && e != nil {
				err = e
			}
			lock.Unlock()
		}(i)
	}

	wg.Wait()

	if err != nil {
		os.Exit(1)
	}

}

func runTests(tests []test, solo bool, idx int) error {

	ctx := context.Background()

	c, closer, err := client()
	if err != nil {
		err = fmt.Errorf("failed to get client: %v", err)
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	defer closer.Close()

	var total, success, fail int

	var out io.Writer
	out = os.Stderr

	namespace := "test"
	if !solo {
		namespace = fmt.Sprintf("test-%d", idx)
	}

	err = testReset(ctx, c, namespace)
	if err != nil {
		err = fmt.Errorf("failed to reset test namespace: %v", err)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	for _, test := range tests {

		if !solo {
			lbls := test.Labels()
			race := false
			for _, lbl := range lbls {
				if lbl == "race" {
					race = true
				}
			}
			if race {
				continue
			}
		}

		total++

		var buf *bytes.Buffer

		if !solo {
			buf = new(bytes.Buffer)
			out = buf
		}

		msg := fmt.Sprintf("Running test %s...", test.Name())
		if len(msg) < 60 {
			msg += strings.Repeat(" ", 60-len(msg))
		}
		fmt.Fprint(out, msg)

		var tctx context.Context
		var cancel context.CancelFunc

		if test.IsLong() {
			tctx, cancel = context.WithCancel(ctx)
		} else {
			tctx, cancel = context.WithTimeout(ctx, testTimeout)
		}

		err := test.Run(tctx, c, namespace)
		cancel()
		if err != nil {
			fail++
			fmt.Fprint(out, "FAIL\n")
			fmt.Fprintf(out, "\tError: %v\n", err)
		} else {
			success++
			fmt.Fprint(out, "SUCCESS\n")
		}

		if buf != nil {
			_, _ = io.Copy(os.Stderr, bytes.NewReader(buf.Bytes()))
		}

		err = testReset(ctx, c, namespace)
		if err != nil {
			err = fmt.Errorf("failed to reset test namespace: %v", err)
			fmt.Fprintln(os.Stderr, err)
			return err
		}

	}

	if fail > 0 {
		return errors.New("tests failed")
	}

	return nil

}

func testsFullReset() error {

	ctx := context.Background()

	c, closer, err := client()
	if err != nil {
		err = fmt.Errorf("failed to get client: %v", err)
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	defer closer.Close()

	prefix := "test"

	namespaces, err := c.Namespaces(ctx, &grpc.NamespacesRequest{
		Pagination: &grpc.Pagination{
			Filter: &grpc.PageFilter{
				Field: "NAME",
				Type:  "CONTAINS",
				Val:   prefix,
			},
		},
	})
	if err != nil {
		return err
	}

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

func testReset(ctx context.Context, c grpc.FlowClient, namespace string) error {

	_, err := c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
		Name:       namespace,
		Idempotent: true,
		Recursive:  true,
	})
	if err != nil {
		return err
	}

	return nil

}

var tests []test

type test interface {
	Name() string
	Labels() []string
	IsLong() bool
	Run(context.Context, grpc.FlowClient, string) error
}

type testImpl struct {
	name   string
	labels []string
	run    func(context.Context, grpc.FlowClient, string) error
}

func (t *testImpl) Name() string {
	return t.name
}

func (t *testImpl) IsLong() bool {
	lbl := t.Labels()
	for i := 0; i < len(lbl); i++ {
		if lbl[i] == "long" {
			//
			return true
		}
	}
	return false
}

func (t *testImpl) Labels() []string {
	return append(t.labels, t.name)
}

func (t *testImpl) Run(ctx context.Context, c grpc.FlowClient, namespace string) error {
	return t.run(ctx, c, namespace)
}

func registerTest(name string, labels []string, fn func(context.Context, grpc.FlowClient, string) error) {
	tests = append(tests, &testImpl{
		name:   name,
		labels: labels,
		run:    fn,
	})
}
