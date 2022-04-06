package main

import (
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/spf13/cobra"
)

func init() {
	cmd := startWorkflowCmd
	cmd.Flags().BoolVar(&stdin, "stdin", false, "")
	cmd.Flags().StringVar(&filein, "input", "", "")
	cmd.Flags().BoolVar(&stream, "stream", false, "")
}

var startWorkflowCmd = &cobra.Command{
	Use:  "start-workflow NAMESPACE PATH [REF]",
	Args: cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {

		source, err := loadSource()
		if err != nil {
			exit(err)
		}

		namespace := args[0]
		path := args[1]
		ref := ""
		if len(args) > 2 {
			ref = args[2]
		}

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.StartWorkflowRequest{
			Namespace: namespace,
			Path:      path,
			Ref:       ref,
			Input:     source,
		}

		if stream {

			// TODO
			/*
				srv, err := c.RunWorkflow(ctx, req)
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
			*/

		} else {

			resp, err := c.StartWorkflow(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}

	},
}

func init() {
	addPaginationFlags(instancesCmd)
}

var instancesCmd = &cobra.Command{
	Use:  "instances NAMESPACE",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.InstancesRequest{
			Pagination: &grpc.Pagination{
				After:  after,
				First:  first,
				Before: before,
				Last:   last,
				Order: []*grpc.PageOrder{{
					Field:     orderField,
					Direction: orderDirection,
				}},
				Filter: []*grpc.PageFilter{{
					Field: filterField,
					Type:  filterType,
					Val:   filterVal,
				}},
			},
			Namespace: namespace,
		}

		if stream {

			srv, err := c.InstancesStream(ctx, req)
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

			resp, err := c.Instances(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}

	},
}

var instanceCmd = &cobra.Command{
	Use:  "instance NAMESPACE INSTANCE",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		instance := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.InstanceRequest{
			Namespace: namespace,
			Instance:  instance,
		}

		if stream {

			srv, err := c.InstanceStream(ctx, req)
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

			resp, err := c.Instance(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}

	},
}

var instanceInputCmd = &cobra.Command{
	Use:  "instance-input NAMESPACE INSTANCE",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		instance := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.InstanceInputRequest{
			Namespace: namespace,
			Instance:  instance,
		}

		resp, err := c.InstanceInput(ctx, req)
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

var instanceOutputCmd = &cobra.Command{
	Use:  "instance-output NAMESPACE INSTANCE",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		instance := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.InstanceOutputRequest{
			Namespace: namespace,
			Instance:  instance,
		}

		resp, err := c.InstanceOutput(ctx, req)
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}
