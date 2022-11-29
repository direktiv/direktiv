package main

/*

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

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/spf13/cobra"
)

var skipLongTests bool
var persistTest bool
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
		registerTest("InstanceEventAnd", []string{"instances", "events"}, testInstanceEventAnd)
		registerTest("InstanceEventXor", []string{"instances", "events"}, testInstanceEventXor)
		registerTest("InstanceValidate", []string{"instances"}, testInstanceValidate)
		registerTest("InstanceDelayLoop", []string{"instances", "long"}, testInstanceDelayLoop)
		registerTest("InstanceForeach", []string{"instances", "long"}, testInstanceForeach)
		registerTest("InstanceParallel", []string{"instances", "long"}, testInstanceParallel)
		registerTest("InstanceError", []string{"instances"}, testInstanceError)
		registerTest("InstanceGenerateConsumeEvent", []string{"instances", "event"}, testInstanceGenerateConsumeEvent)
		registerTest("InstanceTimeoutKill", []string{"instances", "timeout"}, testInstanceTimeoutKill)
		registerTest("InstanceTimeoutKillLong", []string{"instances", "timeout", "long"}, testInstanceTimeoutKillLong)
		registerTest("InstanceTimeoutInterrupt", []string{"instances", "timeout"}, testInstanceTimeoutInterrupt)
		registerTest("InstanceTimeoutInterruptLong", []string{"instances", "timeout", "long"}, testInstanceTimeoutInterruptLong)

		registerTest("InstanceSubflowRetry", []string{"instances", "retry", "long"}, testInstanceSubflowRetry)
		registerTest("InstanceLongRetry", []string{"instances", "retry", "long"}, testInstanceLongRetry)
		registerTest("InstanceActionRetry", []string{"instances", "retry", "long"}, testInstanceActionRetry)
		registerTest("InstanceNestedRetry", []string{"instances", "retry", "long"}, testInstanceNestedRetry)
		registerTest("InstanceParallelRetry", []string{"instances", "retry", "long"}, testInstanceParallelRetry)

		registerTest("InstanceSubflowSecrets", []string{"instances", "jq", "secrets", "actions", "subflows"}, testInstanceSubflowSecrets)
		registerTest("NamespaceVariablesEmpty", []string{"variables"}, testNamespaceVariablesEmpty)
		registerTest("NamespaceVariablesSmall", []string{"variables"}, testNamespaceVariablesSmall)
		registerTest("NamespaceVariablesLarge", []string{"variables", "long"}, testNamespaceVariablesLarge)
		registerTest("WorkflowVariablesEmpty", []string{"variables"}, testWorkflowVariablesEmpty)
		registerTest("WorkflowVariablesSmall", []string{"variables"}, testWorkflowVariablesSmall)
		registerTest("WorkflowVariablesLarge", []string{"variables", "long"}, testWorkflowVariablesLarge)
		registerTest("InstanceNamespaceVariables", []string{"instances", "jq", "variables"}, testInstanceNamespaceVariables)
		registerTest("InstanceWorkflowVariables", []string{"instances", "jq", "variables"}, testInstanceWorkflowVariables)
		registerTest("InstanceInstanceVariables", []string{"instances", "jq", "variables"}, testInstanceInstanceVariables)

		// start types
		registerTest("StartTypeEvent", []string{"events", "start"}, testStartTypeEvent)
		registerTest("StartTypeEventAnd", []string{"events", "start"}, testStartTypeEventAnd)
		registerTest("StartTypeEventXor", []string{"events", "start"}, testStartTypeEventXor)
		registerTest("StartTypeCron", []string{"cron", "start", "long"}, testStartTypeCron)

		// TODO:
			// Error State (uses a validate state to check email and then gets caught by a catch to the error state)
			// ValidateState (done checks if valid or invalid in two different workflows)
			// Foreach State (done runs a foreach for 3 objects)
			// Parallel State (done runs 3 separate workflows 1 mode or with a failing action, 1 mode and with a failing action, 1 mode and with completed actions)
			// CloudEvents (done eventAnd, evenXor and start type events)

			// Crons (runs for 3 minutes checks how many instances it created to see if it matched)

			// Action Types
			// 	Global
			// 	Namespace
			// 	Reusable
			// 	Isolate

			// Workflow Management Tests (revisions, tags, routers, etc)

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

	err := testsFullReset()
	if err != nil {
		os.Exit(1)
	}

	if !persistTest || c != 1 {
		defer func() {
			err := testsFullReset()
			if err != nil {
				os.Exit(1)
			}
		}()
	}

	if c == 1 {
		err := runTests(tests, true, 0)
		if err != nil {
			os.Exit(1)
		}
		return
	}

	var wg sync.WaitGroup
	wg.Add(c)

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

	for i, test := range tests {

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
			if persistTest && solo {
				// Exit on first failed
				fmt.Fprint(out, "Skipping cleanup and aborting tests...\n")
				break
			}
		} else {
			success++
			fmt.Fprint(out, "SUCCESS\n")
		}

		if buf != nil {
			_, _ = io.Copy(os.Stderr, bytes.NewReader(buf.Bytes()))
		}

		if (i == len(tests)-1) && persistTest && solo {
			// Exit on last success
			fmt.Fprint(out, "Skipping cleanup...\n")
			break
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
			Filter: []*grpc.PageFilter{{
				Field: "NAME",
				Type:  "CONTAINS",
				Val:   prefix,
			}},
		},
	})
	if err != nil {
		return err
	}

	for _, result := range namespaces.Results {
		if strings.HasPrefix(result.Name, prefix) {
			_, err = c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
				Name:       result.Name,
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

*/
