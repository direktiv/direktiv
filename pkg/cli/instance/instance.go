package instance

import (
	"github.com/spf13/cobra"

	"github.com/vorteil/direktiv/pkg/cli/util"
)

// "github.com/vorteil/direktiv/pkg/cli/util"

// CreateCommand adds instance commands
func CreateCommand() *cobra.Command {

	cmd := util.GenerateCmd("instances", "List, get and retrieve logs for instances", "", nil, nil)

	cmd.AddCommand(instanceGetCmd)
	cmd.AddCommand(instanceListCmd)
	cmd.AddCommand(instanceLogsCmd)

	return cmd

}

var instanceGetCmd = util.GenerateCmd("get ID", "Get details about a workflow instance", "", func(cmd *cobra.Command, args []string) {
	// resp, err := instance.Get(conn, args[0])
	// if err != nil {
	// 	logger.Errorf(err.Error())
	// 	os.Exit(1)
	// }
	// if flagJSON {
	// 	util.WriteJSON(resp, logger)
	// } else {
	// 	logger.Printf("ID: %s", resp.GetId())
	// 	logger.Printf("Input: %s", string(resp.GetInput()))
	// 	logger.Printf("Output: %s", string(resp.GetOutput()))
	// }
}, cobra.ExactArgs(1))

var instanceLogsCmd = util.GenerateCmd("logs ID", "Grabs all logs for the instance ID provided", "", func(cmd *cobra.Command, args []string) {
	// logs, err := instance.Logs(conn, args[0])
	// if err != nil {
	// 	logger.Errorf(err.Error())
	// 	os.Exit(1)
	// }
	// if flagJSON {
	// 	util.WriteJSON(logs, logger)
	// } else {
	// 	for _, log := range logs {
	// 		fmt.Println(log.GetMessage())
	// 	}
	// }

}, cobra.ExactArgs(1))

var instanceListCmd = util.GenerateCmd("list NAMESPACE", "List all workflow instances from the provided namespace", "", func(cmd *cobra.Command, args []string) {
	// list, err := instance.List(conn, args[0])
	// if err != nil {
	// 	logger.Errorf(err.Error())
	// 	os.Exit(1)
	// }
	//
	// if flagJSON {
	// 	util.WriteJsonList(list, logger)
	// } else {
	// 	if len(list) == 0 {
	// 		logger.Printf("No instances exist under '%s'", args[0])
	// 		return
	// 	}
	//
	// 	table := tablewriter.NewWriter(os.Stdout)
	// 	table.SetHeader([]string{"ID", "Status"})
	//
	// 	// Build string array rows
	// 	for _, instance := range list {
	// 		table.Append([]string{
	// 			instance.GetId(),
	// 			instance.GetStatus(),
	// 		})
	// 	}
	// 	table.Render()
	// }

}, cobra.ExactArgs(1))

// Logs returns all logs associated with the workflow instance ID
// func Logs(conn *grpc.ClientConn, id string) ([]*ingress.GetWorkflowInstanceLogsResponse_WorkflowInstanceLog, error) {
// 	client, ctx, cancel := util.CreateClient(conn)
// 	defer cancel()
// 	offset := int32(0)
// 	limit := int32(10000)
//
// 	// prepare request
// 	request := ingress.GetWorkflowInstanceLogsRequest{
// 		InstanceId: &id,
// 		Offset:     &offset,
// 		Limit:      &limit,
// 	}
//
// 	// send grpc request
// 	resp, err := client.GetWorkflowInstanceLogs(ctx, &request)
// 	if err != nil {
// 		s := status.Convert(err)
// 		return nil, fmt.Errorf("[%v] %v", s.Code(), s.Message())
// 	}
//
// 	return resp.GetWorkflowInstanceLogs(), nil
// }
//
// // List workflow instances
// func List(conn *grpc.ClientConn, namespace string) ([]*ingress.GetWorkflowInstancesResponse_WorkflowInstance, error) {
// 	client, ctx, cancel := util.CreateClient(conn)
// 	defer cancel()
// 	// prepare request
// 	request := ingress.GetWorkflowInstancesRequest{
// 		Namespace: &namespace,
// 	}
//
// 	// send grpc request
// 	resp, err := client.GetWorkflowInstances(ctx, &request)
// 	if err != nil {
// 		s := status.Convert(err)
// 		return nil, fmt.Errorf("[%v] %v", s.Code(), s.Message())
// 	}
//
// 	return resp.WorkflowInstances, nil
// }
//
// // Get returns a workflow instance.
// func Get(conn *grpc.ClientConn, id string) (*ingress.GetWorkflowInstanceResponse, error) {
// 	client, ctx, cancel := util.CreateClient(conn)
// 	defer cancel()
// 	// prepare request
// 	request := ingress.GetWorkflowInstanceRequest{
// 		Id: &id,
// 	}
//
// 	// send grpc request
// 	resp, err := client.GetWorkflowInstance(ctx, &request)
// 	if err != nil {
// 		s := status.Convert(err)
// 		return nil, fmt.Errorf("[%v] %v", s.Code(), s.Message())
// 	}
//
// 	return resp, nil
// }
