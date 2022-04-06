package main

import (
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/spf13/cobra"
)

func init() {
	addPaginationFlags(namespaceLogsCmd)
}

var serverLogsCmd = &cobra.Command{
	Use:  "server-logs",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.ServerLogsRequest{
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
		}

		if stream {

			srv, err := c.ServerLogsParcels(ctx, req)
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

			resp, err := c.ServerLogs(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}

	},
}

var namespaceLogsCmd = &cobra.Command{
	Use:  "namespace-logs NAMESPACE",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.NamespaceLogsRequest{
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

			srv, err := c.NamespaceLogsParcels(ctx, req)
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

			resp, err := c.NamespaceLogs(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}

	},
}

var workflowLogsCmd = &cobra.Command{
	Use:  "workflow-logs NAMESPACE PATH",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		path := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.WorkflowLogsRequest{
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
			Path:      path,
		}

		if stream {

			srv, err := c.WorkflowLogsParcels(ctx, req)
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

			resp, err := c.WorkflowLogs(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}

	},
}

var instanceLogsCmd = &cobra.Command{
	Use:  "instance-logs NAMESPACE INSTANCE",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		instance := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.InstanceLogsRequest{
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
			Instance:  instance,
		}

		if stream {

			srv, err := c.InstanceLogsParcels(ctx, req)
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

			resp, err := c.InstanceLogs(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}

	},
}
