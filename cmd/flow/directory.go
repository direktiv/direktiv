package main

import (
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/spf13/cobra"
)

func init() {
	addPaginationFlags(directoryCmd)
}

var nodeCmd = &cobra.Command{
	Use:  "node NAMESPACE PATH",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		path := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.NodeRequest{
			Namespace: namespace,
			Path:      path,
		}

		resp, err := c.Node(ctx, req)
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

var directoryCmd = &cobra.Command{
	Use:  "directory NAMESPACE PATH",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		path := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.DirectoryRequest{
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

			srv, err := c.DirectoryStream(ctx, req)
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

			resp, err := c.Directory(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}

	},
}

var createDirectoryCmd = &cobra.Command{
	Use:  "create-directory NAMESPACE PATH",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		path := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.CreateDirectory(ctx, &grpc.CreateDirectoryRequest{
			Namespace: namespace,
			Path:      path,
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

var deleteNodeCmd = &cobra.Command{
	Use:  "delete-node NAMESPACE PATH",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		path := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.DeleteNode(ctx, &grpc.DeleteNodeRequest{
			Namespace: namespace,
			Path:      path,
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

var renameNodeCmd = &cobra.Command{
	Use:  "rename-node NAMESPACE OLDPATH NEWPATH",
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		oldpath := args[1]
		newpath := args[2]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.RenameNode(ctx, &grpc.RenameNodeRequest{
			Namespace: namespace,
			Old:       oldpath,
			New:       newpath,
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}
