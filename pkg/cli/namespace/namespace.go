package namespace

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/vorteil/direktiv/pkg/cli/util"
)

// CreateCommand create the namespace command and subcommands
func CreateCommand() *cobra.Command {

	cmd := util.GenerateCmd("namespaces", "List, create and delete namespaces", "", nil, nil)

	cmd.AddCommand(listCmd)
	cmd.AddCommand(createCmd)
	cmd.AddCommand(deleteCmd)

	return cmd

}

type namespacesList struct {
	Namespaces []struct {
		Name      string `json:"name"`
		CreatedAt struct {
			Seconds string `json:"seconds"`
			Nanos   int    `json:"nanos"`
		} `json:"createdAt"`
	} `json:"namespaces"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

var listCmd = util.GenerateCmd("list", "Returns a list of namespaces", "", func(cmd *cobra.Command, args []string) {

	ns, err := util.DoRequest(http.MethodGet, "/namespaces/")
	if err != nil {
		log.Fatalf("error gettting namespaces: %v", err)
	}

	var r namespacesList
	err = json.Unmarshal(ns, &r)
	if err != nil {
		log.Fatalf("error gettting namespaces: %v", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name"})
	for _, namespace := range r.Namespaces {
		table.Append([]string{
			namespace.Name,
		})
	}

	table.Render()

}, cobra.ExactArgs(0))

var createCmd = util.GenerateCmd("create NAME", "Creates a namespaces", "", func(cmd *cobra.Command, args []string) {

	ns, err := util.DoRequest(http.MethodPost, fmt.Sprintf("/namespaces/%s", args[0]))
	if err != nil {
		log.Fatalf("error gettting namespaces: %v", err)
	}

	fmt.Printf(">>JJJ %s", ns)

}, cobra.ExactArgs(1))

var deleteCmd = util.GenerateCmd("delete", "Deletes a namespace", "", func(cmd *cobra.Command, args []string) {

}, cobra.ExactArgs(0))

// // SendEvent sends the provided Cloud Event file to the specified namespace.
// func SendEvent(conn *grpc.ClientConn, namespace string, filepath string) (string, error) {
// 	client, ctx, cancel := util.CreateClient(conn)
// 	defer cancel()
//
// 	// read Cloud Event file
// 	event, err := ioutil.ReadFile(filepath)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	// prepare request
// 	request := ingress.BroadcastEventRequest{
// 		Namespace:  &namespace,
// 		Cloudevent: event,
// 	}
//
// 	// send grpc request
// 	_, err = client.BroadcastEvent(ctx, &request)
// 	if err != nil {
// 		s := status.Convert(err)
// 		return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
// 	}
//
// 	return fmt.Sprintf("Successfully sent event to '%s'", namespace), nil
// }
//
// // List returns a list of namespaces
// func List(conn *grpc.ClientConn) ([]*ingress.GetNamespacesResponse_Namespace, error) {
// 	client, ctx, cancel := util.CreateClient(conn)
// 	defer cancel()
//
// 	// prepare request
// 	request := ingress.GetNamespacesRequest{}
//
// 	// send grpc request
// 	resp, err := client.GetNamespaces(ctx, &request)
// 	if err != nil {
// 		s := status.Convert(err)
// 		return nil, fmt.Errorf("[%v] %v", s.Code(), s.Message())
// 	}
//
// 	return resp.Namespaces, nil
// }
//
// // Delete a namespace
// func Delete(name string, conn *grpc.ClientConn) (string, error) {
// 	client, ctx, cancel := util.CreateClient(conn)
// 	defer cancel()
//
// 	// prepare request
// 	request := ingress.DeleteNamespaceRequest{
// 		Name: &name,
// 	}
//
// 	// send grpc request
// 	resp, err := client.DeleteNamespace(ctx, &request)
// 	if err != nil {
// 		s := status.Convert(err)
// 		return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
// 	}
//
// 	return fmt.Sprintf("Deleted namespace: %s", resp.GetName()), nil
// }
//
// // Create a new namespace
// func Create(name string, conn *grpc.ClientConn) (string, error) {
// 	client, ctx, cancel := util.CreateClient(conn)
// 	defer cancel()
//
// 	// prepare request
// 	request := ingress.AddNamespaceRequest{
// 		Name: &name,
// 	}
//
// 	// send grpc request
// 	resp, err := client.AddNamespace(ctx, &request)
// 	if err != nil {
// 		s := status.Convert(err)
// 		return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
// 	}
//
// 	return fmt.Sprintf("Created namespace: %s", resp.GetName()), nil
// }
