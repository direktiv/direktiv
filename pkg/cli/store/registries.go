package store

import (
	"github.com/spf13/cobra"
	"github.com/vorteil/direktiv/pkg/cli/util"
)

// CreateCommandRegistries adds registry commands
func CreateCommandRegistries() *cobra.Command {

	cmd := util.GenerateCmd("registries", "List, create and remove registries from provided namespace", "", nil, nil)

	cmd.AddCommand(createRegistryCmd)
	cmd.AddCommand(removeRegistryCmd)
	cmd.AddCommand(listRegistriesCmd)

	return cmd

}

var createRegistryCmd = util.GenerateCmd("create NAMESPACE URL USER:TOKEN", "Creates a new registry on provided namespace", "", func(cmd *cobra.Command, args []string) {

}, cobra.ExactArgs(3))

var removeRegistryCmd = util.GenerateCmd("delete NAMESPACE URL", "Deletes a registry from the provided namespace", "", func(cmd *cobra.Command, args []string) {

}, cobra.ExactArgs(2))

var listRegistriesCmd = util.GenerateCmd("list NAMESPACE", "Returns a list of registries from the provided namespace", "", func(cmd *cobra.Command, args []string) {

}, cobra.ExactArgs(1))

// type StoreRequest struct {
// 	Key   string
// 	Value string
// }
//
// func List(conn *grpc.ClientConn, namespace string, typeOf string) (interface{}, error) {
// 	var ifc interface{}
//
// 	client, ctx, cancel := util.CreateClient(conn)
// 	defer cancel()
// 	switch typeOf {
// 	case "secret":
// 		// prepare request
// 		request := ingress.GetSecretsRequest{
// 			Namespace: &namespace,
// 		}
//
// 		// send grpc request
// 		resp, err := client.GetSecrets(ctx, &request)
// 		if err != nil {
// 			s := status.Convert(err)
// 			return nil, fmt.Errorf("[%v] %v", s.Code(), s.Message())
// 		}
//
// 		ifc = resp.Secrets
//
// 	case "registry":
//
// 		// prepare request
// 		request := ingress.GetRegistriesRequest{
// 			Namespace: &namespace,
// 		}
//
// 		// send grpc request
// 		resp, err := client.GetRegistries(ctx, &request)
// 		if err != nil {
// 			s := status.Convert(err)
// 			return nil, fmt.Errorf("[%v] %v", s.Code(), s.Message())
// 		}
// 		ifc = resp.Registries
// 	}
//
// 	return ifc, nil
// }
//
// func Delete(conn *grpc.ClientConn, namespace string, secret string, typeOf string) (string, error) {
// 	var success string
// 	var err error
//
// 	client, ctx, cancel := util.CreateClient(conn)
// 	defer cancel()
//
// 	switch typeOf {
// 	case "secret":
//
// 		// prepare request
// 		request := ingress.DeleteSecretRequest{
// 			Namespace: &namespace,
// 			Name:      &secret,
// 		}
//
// 		// send grpc request
// 		_, err := client.DeleteSecret(ctx, &request)
// 		if err != nil {
// 			s := status.Convert(err)
// 			return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
// 		}
// 		success = fmt.Sprintf("Successfully removed secret '%s'.", secret)
//
// 	case "registry":
//
// 		// prepare request
// 		request := ingress.DeleteRegistryRequest{
// 			Namespace: &namespace,
// 			Name:      &secret,
// 		}
//
// 		// send grpc request
// 		_, err := client.DeleteRegistry(ctx, &request)
// 		if err != nil {
// 			s := status.Convert(err)
// 			return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
// 		}
//
// 		success = fmt.Sprintf("Successfully removed registry '%s'.", secret)
// 	}
// 	return success, err
// }
//
// func Create(conn *grpc.ClientConn, namespace string, s *StoreRequest, typeOf string) (string, error) {
//
// 	var success string
// 	var err error
//
// 	client, ctx, cancel := util.CreateClient(conn)
// 	defer cancel()
//
// 	switch typeOf {
//
// 	case "secret":
//
// 		// prepare request
// 		request := ingress.StoreSecretRequest{
// 			Namespace: &namespace,
// 			Name:      &s.Key,
// 			Data:      []byte(s.Value),
// 		}
//
// 		// send grpc request
// 		_, err := client.StoreSecret(ctx, &request)
// 		if err != nil {
// 			s := status.Convert(err)
// 			return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
// 		}
//
// 		success = fmt.Sprintf("Successfully create secret '%s'.", s.Key)
//
// 	case "registry":
//
// 		// prepare request
// 		request := ingress.StoreRegistryRequest{
// 			Namespace: &namespace,
// 			Name:      &s.Key,
// 			Data:      []byte(s.Value),
// 		}
//
// 		// send grpc request
// 		_, err := client.StoreRegistry(ctx, &request)
// 		if err != nil {
// 			s := status.Convert(err)
// 			return "", fmt.Errorf("[%v] %v", s.Code(), s.Message())
// 		}
//
// 		success = fmt.Sprintf("Successfully created registry '%s'.", s.Key)
// 	}
//
// 	return success, err
// }
