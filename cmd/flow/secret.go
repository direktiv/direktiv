package main

import (
	"github.com/spf13/cobra"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
)

func init() {
	addPaginationFlags(secretsCmd)
}

var secretsCmd = &cobra.Command{
	Use:  "secrets NAMESPACE",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.SecretsRequest{
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
		}

		if stream {

			srv, err := c.SecretsStream(ctx, req)
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

			resp, err := c.Secrets(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}

	},
}

func init() {
	cmd := setSecretCmd
	cmd.Flags().BoolVar(&stdin, "stdin", false, "")
	cmd.Flags().StringVar(&filein, "input", "", "")
}

var setSecretCmd = &cobra.Command{
	Use:  "set-secret NAMESPACE KEY",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		key := args[1]

		source, err := loadSource()
		if err != nil {
			exit(err)
		}

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.SetSecret(ctx, &grpc.SetSecretRequest{
			Namespace: namespace,
			Key:       key,
			Data:      source,
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}

var deleteSecretCmd = &cobra.Command{
	Use:  "delete-secret NAMESPACE KEY",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		key := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		resp, err := c.DeleteSecret(ctx, &grpc.DeleteSecretRequest{
			Namespace: namespace,
			Key:       key,
		})
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}
