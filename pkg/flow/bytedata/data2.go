package bytedata

import (
	"path/filepath"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
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
	}
	if node.Name == "/" {
		node.Name = ""
	}
	switch node.Type {
	case string(filestore.FileTypeDirectory):
		node.ExpandedType = string(filestore.FileTypeDirectory)
	case string(filestore.FileTypeWorkflow):
		node.ExpandedType = string(filestore.FileTypeWorkflow)
	default:
		node.ExpandedType = string(filestore.FileTypeFile)
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

func ConvertRevisionToGrpcRevision(rev *filestore.Revision) *grpc.Revision {
	return &grpc.Revision{
		Name:      rev.ID.String(),
		CreatedAt: timestamppb.New(rev.CreatedAt),
		Hash:      rev.Checksum,
	}
}

func ConvertRevisionsToGrpcRevisionList(list []*filestore.Revision) []*grpc.Revision {
	var result []*grpc.Revision
	for _, f := range list {
		result = append(result, ConvertRevisionToGrpcRevision(f))
	}

	return result
}

func ConvertRevisionToGrpcRef(rev *filestore.Revision) *grpc.Ref {
	return &grpc.Ref{
		Name: rev.ID.String(),
	}
}

func ConvertRevisionsToGrpcRefList(list []*filestore.Revision) []*grpc.Ref {
	var result []*grpc.Ref
	for _, f := range list {
		result = append(result, ConvertRevisionToGrpcRef(f))
	}

	return result
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

func ConvertMirrorProcessesToGrpcMirrorActivityInfoList(list []*mirror.Process) []*grpc.MirrorActivityInfo {
	var result []*grpc.MirrorActivityInfo
	for _, f := range list {
		result = append(result, ConvertMirrorProcessToGrpcMirrorActivity(f))
	}

	return result
}

func ConvertSecretToGrpcSecret(secret *core.Secret) *grpc.Secret {
	return &grpc.Secret{
		Name: secret.Name,
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
		Checksum:  variable.Hash,
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
