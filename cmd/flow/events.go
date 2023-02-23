package flow

import (
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/spf13/cobra"
)

func init() {
	addPaginationFlags(eventHistoryCmd)
}

var eventHistoryCmd = &cobra.Command{
	Use:  "event-history NAMESPACE",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.EventHistoryRequest{
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

			srv, err := c.EventHistoryStream(ctx, req)
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

			resp, err := c.EventHistory(ctx, req)
			if err != nil {
				exit(err)
			}

			print(resp)

		}

	},
}

var eventReplayCmd = &cobra.Command{
	Use:  "replay-event NAMESPACE ID",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		namespace := args[0]
		id := args[1]

		c, closer, err := client()
		if err != nil {
			exit(err)
		}
		defer closer.Close()

		req := &grpc.ReplayEventRequest{
			Namespace: namespace,
			Id:        id,
		}

		resp, err := c.ReplayEvent(ctx, req)
		if err != nil {
			exit(err)
		}

		print(resp)

	},
}
