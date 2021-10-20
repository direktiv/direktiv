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

	"github.com/vorteil/direktiv/pkg/api/client/namespaces"

	"github.com/go-openapi/strfmt"
	"github.com/spf13/cobra"
	direktivsdk "github.com/vorteil/direktiv/pkg/api/client"
)

var testsAPICmd = &cobra.Command{
	Use: "testapi",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		registerAPITest("CreateNamespace", []string{"namespaces"}, testAPICreateNamespace)
		registerAPITest("CreateNamespaceDuplicate", []string{"namespaces"}, testAPICreateNamespaceDuplicate)
		registerAPITest("CreateNamespaceRegex", []string{"namespaces"}, testAPICreateNamespaceRegex)
		registerAPITest("DeleteNamespace", []string{"namespaces"}, testAPIDeleteNamespace)
		registerAPITest("ServerLogs", []string{"namespaces"}, testAPIServerLogs)
		registerAPITest("CreateDirectoryDeleteDirectory", []string{"directories"}, testAPICreateDirectoryDeleteDirectory)
		registerAPITest("Directory", []string{"directories"}, testAPIDirectory)
		registerAPITest("Secrets", []string{"secrets"}, testAPISecrets)
		registerAPITest("Workflow", []string{"workflow"}, testAPIWorkflow)
	},
	Run: func(cmd *cobra.Command, args []string) {

		tests := getAPITests(args...)
		if len(tests) == 0 {
			fmt.Println("no results")
			return
		}

		runTestsAPIParallel(tests, parallelTests)

	},
}

var testsApi []testApi

type testApi interface {
	Name() string
	Labels() []string
	IsLong() bool
	Run(context.Context, direktivsdk.Direktivsdk, string) error
}

type testApiImpl struct {
	name   string
	labels []string
	run    func(context.Context, direktivsdk.Direktivsdk, string) error
}

func (t *testApiImpl) Name() string {
	return t.name
}

func (t *testApiImpl) IsLong() bool {
	lbl := t.Labels()
	for i := 0; i < len(lbl); i++ {
		if lbl[i] == "long" {
			//
			return true
		}
	}
	return false
}

func (t *testApiImpl) Labels() []string {
	return append(t.labels, t.name)
}

func (t *testApiImpl) Run(ctx context.Context, c direktivsdk.Direktivsdk, namespace string) error {
	return t.run(ctx, c, namespace)
}

func registerAPITest(name string, labels []string, fn func(context.Context, direktivsdk.Direktivsdk, string) error) {
	testsApi = append(testsApi, &testApiImpl{
		name:   name,
		labels: labels,
		run:    fn,
	})
}
func getAPITests(labels ...string) []testApi {
	var subset = make([]testApi, 0)

	for _, test := range testsApi {

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
func runTestsAPIParallel(tests []testApi, c int) {
	testsFullReset()
	if !persistTest || c != 1 {
		defer testsFullReset()
	}

	if c == 1 {
		err := runAPITests(tests, true, 0)
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
			e := runAPITests(tests, false, i)
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
func runAPITests(tests []testApi, solo bool, idx int) error {
	ctx := context.Background()

	tc := direktivsdk.TransportConfig{
		Host:     addr,
		BasePath: "/",
		Schemes:  []string{"http"},
	}

	rc := direktivsdk.NewHTTPClientWithConfig(strfmt.Default, &tc)

	var total, success, fail int

	var out io.Writer
	out = os.Stderr

	namespace := "test"
	if !solo {
		namespace = fmt.Sprintf("test-%d", idx)
	}

	err := testResetApi(ctx, *rc, namespace)
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

		err := test.Run(tctx, *rc, namespace)
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

		err = testResetApi(ctx, *rc, namespace)
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

func testResetApi(ctx context.Context, c direktivsdk.Direktivsdk, namespace string) error {
	_, aerr := c.Namespaces.DeleteNamespace(&namespaces.DeleteNamespaceParams{
		Namespace: namespace,
		Context:   ctx,
	}, nil)
	if aerr != nil {
		err := aerr.(*namespaces.DeleteNamespaceDefault)
		if err.Code() != 404 {
			return errors.New(err.Payload.Error)
		}
	}
	return nil
}
