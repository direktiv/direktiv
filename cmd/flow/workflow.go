package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
)

var workflowCmd = &cobra.Command{
	Use:  "workflow NAMESPACE PATH [REF]",
	Args: cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		path := args[1]
		ref := "latest"
		if len(args) > 2 {
			ref = args[2]
		}

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.WorkflowRequest{
			Namespace: namespace,
			Path:      path,
			Ref:       ref,
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

	data, err := ioutil.ReadFile(filein)
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

var editRouterCmd = &cobra.Command{
	Use:  "edit-router NAMESPACE PATH ACTIVE",
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		path := args[1]
		active := args[2]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}

		defer closer.Close()

		dv, _ := strconv.ParseBool(active)
		resp, err := c.EditRouter(ctx, &grpc.EditRouterRequest{
			Namespace: namespace,
			Path:      path,
			Live:      dv,
		})
		if err != nil {
			exit(err)
		}

		print(resp)
	},
}

var routerCmd = &cobra.Command{
	Use:  "router NAMESPACE PATH",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		path := args[1]

		// source, err := loadSource()
		// if err != nil {
		// 	exit(err)
		// }

		c, closer, err := client()
		if err != nil {
			exit(err)
		}

		defer closer.Close()

		resp, err := c.Router(ctx, &grpc.RouterRequest{
			Namespace: namespace,
			Path:      path,
		})
		if err != nil {
			exit(err)
		}

		print(resp)
	},
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

var saveHeadCmd = &cobra.Command{
	Use:  "save-head NAMESPACE PATH",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		path := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.SaveHead(ctx, &grpc.SaveHeadRequest{
			Namespace: namespace,
			Path:      path,
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

var discardHeadCmd = &cobra.Command{
	Use:  "discard-head NAMESPACE PATH",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		path := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.DiscardHead(ctx, &grpc.DiscardHeadRequest{
			Namespace: namespace,
			Path:      path,
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

func init() {
	addPaginationFlags(tagsCmd)
}

var tagsCmd = &cobra.Command{
	Use:  "tags NAMESPACE PATH",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		path := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.TagsRequest{
			Pagination: &grpc.Pagination{
				After:  after,
				First:  first,
				Before: before,
				Last:   last,
				Order: &grpc.PageOrder{
					Field:     orderField,
					Direction: orderDirection,
				},
				Filter: &grpc.PageFilter{
					Field: filterField,
					Type:  filterType,
					Val:   filterVal,
				},
			},
			Namespace: namespace,
			Path:      path,
		}

		if stream {

			srv, err := c.TagsStream(ctx, req)
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

			resp, err := c.Tags(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}

	},
}

var refsCmd = &cobra.Command{
	Use:  "refs NAMESPACE PATH",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		path := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.RefsRequest{
			Pagination: &grpc.Pagination{
				After:  after,
				First:  first,
				Before: before,
				Last:   last,
				Order: &grpc.PageOrder{
					Field:     orderField,
					Direction: orderDirection,
				},
				Filter: &grpc.PageFilter{
					Field: filterField,
					Type:  filterType,
					Val:   filterVal,
				},
			},
			Namespace: namespace,
			Path:      path,
		}

		if stream {

			srv, err := c.RefsStream(ctx, req)
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

			resp, err := c.Refs(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}

	},
}

var revisionsCmd = &cobra.Command{
	Use:  "revisions NAMESPACE PATH",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		path := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.RevisionsRequest{
			Pagination: &grpc.Pagination{
				After:  after,
				First:  first,
				Before: before,
				Last:   last,
				Order: &grpc.PageOrder{
					Field:     orderField,
					Direction: orderDirection,
				},
				Filter: &grpc.PageFilter{
					Field: filterField,
					Type:  filterType,
					Val:   filterVal,
				},
			},
			Namespace: namespace,
			Path:      path,
		}

		if stream {

			srv, err := c.RevisionsStream(ctx, req)
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

			resp, err := c.Revisions(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}

	},
}
