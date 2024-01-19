package bytedata

import (
	"path/filepath"
	"sort"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertFileToGrpcNode(f *filestore.File) *grpc.Node {
	node := &grpc.Node{
		CreatedAt:  timestamppb.New(f.CreatedAt),
		UpdatedAt:  timestamppb.New(f.UpdatedAt),
		Name:       filepath.Base(f.Path),
		Path:       f.Path,
		Parent:     filepath.Dir(f.Path),
		Type:       string(f.Typ),
		Attributes: []string{},
		Oid:        "", // NOTE: this is empty string for now for compatibility with end-to-end tests f.ID.String(),
		ReadOnly:   false,
		MimeType:   f.MIMEType,
	}
	if node.Name == "/" {
		node.Name = ""
	}

	switch node.Type {
	case string(filestore.FileTypeDirectory):
		node.ExpandedType = string(filestore.FileTypeDirectory)
		node.MimeType = ""
	case string(filestore.FileTypeWorkflow):
		node.ExpandedType = string(filestore.FileTypeWorkflow)
		node.MimeType = "application/direktiv"
	default:
		node.ExpandedType = string(filestore.FileTypeFile)
		node.MimeType = f.MIMEType
	}

	return node
}

func ConvertFilesToGrpcNodeList(list []*filestore.File) []*grpc.Node {
	var result []*grpc.Node
	for _, f := range list {
		result = append(result, ConvertFileToGrpcNode(f))
	}

	return result
}

func ConvertFileToGrpcFile(file *filestore.File) *grpc.File {
	return &grpc.File{
		Name:      file.Name(),
		CreatedAt: timestamppb.New(file.CreatedAt),
		Hash:      file.Checksum,
		MimeType:  file.MIMEType,
	}
}

func ConvertMirrorConfigToGrpcMirrorInfo(config *mirror.Config) *grpc.MirrorInfo {
	return &grpc.MirrorInfo{
		Url: config.URL,
		Ref: config.GitRef,
		// Cron: ,
		PublicKey: config.PublicKey,
		CommitId:  config.GitCommitHash,
		// LastSync: ,
		PrivateKey: config.PrivateKey,
		Passphrase: config.PrivateKeyPassphrase,
		Insecure:   config.Insecure,
	}
}

func ConvertMirrorProcessToGrpcMirrorActivity(mirror *mirror.Process) *grpc.MirrorActivityInfo {
	return &grpc.MirrorActivityInfo{
		Id:        mirror.ID.String(),
		Status:    mirror.Status,
		Type:      mirror.Typ,
		CreatedAt: timestamppb.New(mirror.CreatedAt),
		UpdatedAt: timestamppb.New(mirror.UpdatedAt),
	}
}

// ConvertMirrorProcessesToGrpcMirrorActivityInfoList converts a slice of Process pointers
// into a slice of grpc.MirrorActivityInfo pointers. The resulting slice is sorted
// by the UpdatedAt field in ascending order.
// Parameters:
// list: A slice of pointers to Process objects that need to be converted.
// Returns:
// A slice of pointers to grpc.MirrorActivityInfo objects sorted by UpdatedAt.
func ConvertMirrorProcessesToGrpcMirrorActivityInfoList(list []*mirror.Process) []*grpc.MirrorActivityInfo {
	copiedList := make([]*mirror.Process, len(list))
	copy(copiedList, list)

	// Sort the copied list by UpdatedAt
	sort.Slice(copiedList, func(i, j int) bool {
		return copiedList[i].UpdatedAt.Before(copiedList[j].UpdatedAt)
	})

	var result []*grpc.MirrorActivityInfo
	for _, f := range copiedList {
		result = append(result, ConvertMirrorProcessToGrpcMirrorActivity(f))
	}

	return result
}

func ConvertSecretToGrpcSecret(secret *core.Secret) *grpc.Secret {
	return &grpc.Secret{
		Name:        secret.Name,
		Initialized: secret.Data != nil,
	}
}

func ConvertSecretsToGrpcSecretList(list []*core.Secret) []*grpc.Secret {
	var result []*grpc.Secret
	for _, f := range list {
		result = append(result, ConvertSecretToGrpcSecret(f))
	}

	return result
}

func ConvertRuntimeVariableToGrpcVariable(variable *core.RuntimeVariable) *grpc.Variable {
	return &grpc.Variable{
		Name:      variable.Name,
		Size:      int64(variable.Size),
		MimeType:  variable.MimeType,
		CreatedAt: timestamppb.New(variable.CreatedAt),
		UpdatedAt: timestamppb.New(variable.UpdatedAt),
	}
}

func ConvertRuntimeVariablesToGrpcVariableList(list []*core.RuntimeVariable) []*grpc.Variable {
	var result []*grpc.Variable
	for _, f := range list {
		result = append(result, ConvertRuntimeVariableToGrpcVariable(f))
	}

	return result
}

func ConvertInstanceToGrpcInstance(instance *enginerefactor.Instance) *grpc.Instance {
	return &grpc.Instance{
		CreatedAt:    timestamppb.New(instance.Instance.CreatedAt),
		UpdatedAt:    timestamppb.New(instance.Instance.UpdatedAt),
		Id:           instance.Instance.ID.String(),
		As:           instance.Instance.WorkflowPath,
		Status:       instance.Instance.Status.String(),
		ErrorCode:    instance.Instance.ErrorCode,
		ErrorMessage: string(instance.Instance.ErrorMessage),
		Invoker:      instance.Instance.Invoker,
	}
}

func ConvertInstancesToGrpcInstances(instances []instancestore.InstanceData) []*grpc.Instance {
	list := make([]*grpc.Instance, 0)
	for idx := range instances {
		instance := &instances[idx]
		list = append(list, &grpc.Instance{
			CreatedAt:    timestamppb.New(instance.CreatedAt),
			UpdatedAt:    timestamppb.New(instance.UpdatedAt),
			Id:           instance.ID.String(),
			As:           instance.WorkflowPath,
			Status:       instance.Status.String(),
			ErrorCode:    instance.ErrorCode,
			ErrorMessage: string(instance.ErrorMessage),
			Invoker:      instance.Invoker,
		})
	}

	return list
}

func ConvertNamespaceToGrpc(item *core.Namespace) *grpc.Namespace {
	ns := &grpc.Namespace{
		Oid:  item.ID.String(),
		Name: item.Name,

		CreatedAt: timestamppb.New(item.CreatedAt),
		UpdatedAt: timestamppb.New(item.UpdatedAt),
	}

	return ns
}

func ConvertNamespacesListToGrpc(list []*core.Namespace) []*grpc.Namespace {
	var result []*grpc.Namespace
	for idx := range list {
		ns := list[idx]
		result = append(result, ConvertNamespaceToGrpc(ns))
	}

	return result
}
