package main

import (
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/spf13/cobra"
)

var namespaceCmd = &cobra.Command{
	Use:  "namespace NAME",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.Namespace(ctx, &grpc.NamespaceRequest{
			Name: args[0],
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

func init() {
	addPaginationFlags(namespacesCmd)
}

var namespacesCmd = &cobra.Command{
	Use:  "namespaces",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.NamespacesRequest{
			Pagination: &grpc.Pagination{
				Limit:  limit,
				Offset: offset,
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

			srv, err := c.NamespacesStream(ctx, req)
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

			resp, err := c.Namespaces(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}

	},
}

var createNamespaceCmd = &cobra.Command{
	Use:  "create-namespace NAME",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.CreateNamespace(ctx, &grpc.CreateNamespaceRequest{
			Name: args[0],
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

var deleteNamespaceCmd = &cobra.Command{
	Use:  "delete-namespace NAME",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.DeleteNamespace(ctx, &grpc.DeleteNamespaceRequest{
			Name: args[0],
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

var renameNamespaceCmd = &cobra.Command{
	Use:  "rename-namespace OLDNAME NEWNAME",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.RenameNamespace(ctx, &grpc.RenameNamespaceRequest{
			Old: args[0],
			New: args[1],
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

var annotationSetNamespaceCmd = &cobra.Command{
	Use:  "annotation-set-namespace NAMESPACE KEY VALUE",
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.SetNamespaceAnnotation(ctx, &grpc.SetNamespaceAnnotationRequest{
			Namespace: args[0],
			Key:       args[1],
			Data:      []byte(args[2]),
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

var annotationGetNamespaceCmd = &cobra.Command{
	Use:  "annotation-get-namespace NAMESPACE KEY",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.NamespaceAnnotation(ctx, &grpc.NamespaceAnnotationRequest{
			Namespace: args[0],
			Key:       args[1],
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

var annotationListNamespaceCmd = &cobra.Command{
	Use:  "annotation-list-namespace NAMESPACE",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.NamespaceAnnotations(ctx, &grpc.NamespaceAnnotationsRequest{
			Namespace: args[0],
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

var annotationRenameNamespaceCmd = &cobra.Command{
	Use:  "annotation-rename-namespace NAMESPACE OLDKEY NEWKEY",
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.RenameNamespaceAnnotation(ctx, &grpc.RenameNamespaceAnnotationRequest{
			Namespace: args[0],
			Old:       args[1],
			New:       args[2],
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

var annotationDeleteNamespaceCmd = &cobra.Command{
	Use:  "annotation-delete-namespace NAMESPACE KEY",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.DeleteNamespaceAnnotation(ctx, &grpc.DeleteNamespaceAnnotationRequest{
			Namespace: args[0],
			Key:       args[1],
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}
