package bytedata

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Checksum is a shortcut to calculate a hash for any given input by first marshalling it to json.
func Checksum(x interface{}) string {
	data, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}

	hash, err := ComputeHash(data)
	if err != nil {
		panic(err)
	}

	return hash
}

// ComputeHash is a shortcut to calculate a hash for a byte slice.
func ComputeHash(data []byte) (string, error) {
	hasher := sha256.New()
	_, err := io.Copy(hasher, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// Marshal is a shortcut to marshal any given input to json with our preferred indentation settings.
func Marshal(x interface{}) string {
	data, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(data)
}

func ConvertEventListeners(in []*datastore.EventListener) []*grpc.EventListener {
	res := make([]*grpc.EventListener, 0, len(in))
	for _, el := range in {
		types := []*grpc.EventDef{}
		for _, v := range el.ListeningForEventTypes {
			types = append(types, &grpc.EventDef{Type: v})
		}
		wf := ""
		ins := ""
		// step := ""
		if el.TriggerWorkflow != "" {
			wf = el.Metadata
		}
		if el.TriggerInstance != "" {
			ins = el.TriggerInstance
			// step = fmt.Sprintf("%v", el.TriggerInstanceStep)
		}
		mode := ""
		switch el.TriggerType {
		case datastore.StartAnd, datastore.WaitAnd:
			mode = "and"
		case datastore.StartOR, datastore.WaitOR:
			mode = "or"
		case datastore.StartSimple, datastore.WaitSimple:
			mode = "simple"
		}
		res = append(res, &grpc.EventListener{
			Workflow:  wf,
			Instance:  ins,
			UpdatedAt: timestamppb.New(el.UpdatedAt),
			Mode:      mode,
			Events:    types,
			CreatedAt: timestamppb.New(el.CreatedAt),
		})
	}

	return res
}
