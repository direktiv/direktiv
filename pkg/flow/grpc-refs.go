package flow

import (
	"github.com/direktiv/direktiv/pkg/model"
)

func loadSource(rev []byte) (*model.Workflow, error) {
	workflow := new(model.Workflow)

	err := workflow.Load(rev)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}
