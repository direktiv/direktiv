package isolates

import (
	"context"

	log "github.com/sirupsen/logrus"
	igrpc "github.com/vorteil/direktiv/pkg/isolates/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (is *isolateServer) DeleteRegistry(ctx context.Context, in *igrpc.DeleteRegistryRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty

	// err := kubernetesDeleteSecret(in.GetName(), in.GetNamespace())

	return &resp, nil
}

func (is *isolateServer) StoreRegistry(ctx context.Context, in *igrpc.StoreRegistryRequest) (*emptypb.Empty, error) {
	var resp emptypb.Empty

	// func kubernetesAddSecret(name, namespace string, data []byte) error {

	// 	log.Debugf("adding secret %s (%s)", name, namespace)
	//
	// 	clientset, kns, err := getClientSet()
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	u, err := url.Parse(name)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	secretName := fmt.Sprintf("%s-%s-%s", secretsPrefix, namespace, u.Hostname())
	//
	// 	kubernetesDeleteSecret(name, namespace)
	//
	// 	sa := &v1.Secret{
	// 		Data: make(map[string][]byte),
	// 	}
	//
	// 	sa.Annotations = make(map[string]string)
	// 	sa.Annotations[annotationNamespace] = namespace
	// 	sa.Annotations[annotationURL] = name
	// 	sa.Annotations[annotationURLHash] = base64.StdEncoding.EncodeToString([]byte(name))
	//
	// 	sa.Name = secretName
	// 	sa.Data[".dockerconfigjson"] = data
	// 	sa.Type = "kubernetes.io/dockerconfigjson"
	//
	// 	_, err = clientset.CoreV1().Secrets(kns).Create(context.Background(), sa, metav1.CreateOptions{})
	//
	// 	return err
	//
	// }

	// create secret data, needs to be attached to service account
	// userToken := strings.SplitN(string(in.Data), ":", 2)
	// if len(userToken) != 2 {
	// 	return nil, fmt.Errorf("invalid username/token format")
	// }
	//
	// tmpl := `{
	// "auths": {
	// 	"%s": {
	// 		"username": "%s",
	// 		"password": "%s",
	// 		"auth": "%s"
	// 	}
	// }
	// }`

	// auth := fmt.Sprintf(tmpl, in.GetName(), userToken[0], userToken[1],
	// 	base64.StdEncoding.EncodeToString(in.Data))

	// err := kubernetesAddSecret(in.GetName(), in.GetNamespace(), []byte(auth))
	// if err != nil {
	// 	return nil, err
	// }

	return &resp, nil

}

func (is *isolateServer) GetRegistries(ctx context.Context, in *igrpc.GetRegistriesRequest) (*igrpc.GetRegistriesResponse, error) {

	resp := new(igrpc.GetRegistriesResponse)

	log.Debugf("GET REGISTRIES!!!!")
	// regs, err := kubernetesListRegistries(in.GetNamespace())
	//
	// if err != nil {
	// 	return resp, err
	// }
	//
	// for _, reg := range regs {
	// 	split := strings.SplitN(reg, "###", 2)
	//
	// 	if len(split) != 2 {
	// 		return nil, fmt.Errorf("invalid registry format")
	// 	}
	//
	// 	resp.Registries = append(resp.Registries, &ingress.GetRegistriesResponse_Registry{
	// 		Name: &split[0],
	// 		Id:   &split[1],
	// 	})
	// }

	return resp, nil

}
