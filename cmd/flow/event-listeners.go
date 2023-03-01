package flow

import (
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/spf13/cobra"
)

func init() {
	addPaginationFlags(eventListenersCmd)
}

var eventListenersCmd = &cobra.Command{
	Use:  "event-listeners NAMESPACE",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.EventListenersRequest{
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
			Namespace: namespace,
		}

		if stream {

			srv, err := c.EventListenersStream(ctx, req)
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

			resp, err := c.EventListeners(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}
	},
}
