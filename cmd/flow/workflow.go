package flow

import (
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/spf13/cobra"
)

var workflowCmd = &cobra.Command{
	Use:  "workflow NAMESPACE PATH [REF]",
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		path := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.WorkflowRequest{
			Namespace: namespace,
			Path:      path,
		}

		if stream {

			srv, err := c.WorkflowStream(ctx, req)
			if err != nil {
				exit(err)
			}

			for {

				resp, err := srv.Recv()
				if err != nil {
					exit(err)
				}

				print(resp)

			}

		} else {

			resp, err := c.Workflow(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}
	},
}

func loadSource() ([]byte, error) {
	if stdin && filein != "" {
		return nil, errors.New("--stdin & --input flags are mutually exclusive")
	}

	if stdin {
		buf := new(bytes.Buffer)
		_, err := io.Copy(buf, os.Stdin)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	if filein == "" {
		return nil, nil
	}

	data, err := os.ReadFile(filein)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func init() {
	cmd := createWorkflowCmd
	cmd.Flags().BoolVar(&stdin, "stdin", false, "")
	cmd.Flags().StringVar(&filein, "input", "", "")
}

var createWorkflowCmd = &cobra.Command{
	Use:  "create-workflow NAMESPACE PATH",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		path := args[1]

		source, err := loadSource()
		if err != nil {
			exit(err)
		}

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.CreateWorkflow(ctx, &grpc.CreateWorkflowRequest{
			Namespace: namespace,
			Path:      path,
			Source:    source,
		})
		if err != nil {
			exit(err)
		}

		print(resp)
	},
}

func init() {
	cmd := updateWorkflowCmd
	cmd.Flags().BoolVar(&stdin, "stdin", false, "")
	cmd.Flags().StringVar(&filein, "input", "", "")
}

var updateWorkflowCmd = &cobra.Command{
	Use:  "update-workflow NAMESPACE PATH",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		path := args[1]

		source, err := loadSource()
		if err != nil {
			exit(err)
		}

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.UpdateWorkflow(ctx, &grpc.UpdateWorkflowRequest{
			Namespace: namespace,
			Path:      path,
			Source:    source,
		})
		if err != nil {
			exit(err)
		}

		print(resp)
	},
}

func init() {
	addPaginationFlags(workflowCmd)
}
