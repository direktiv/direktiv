package flow

import (
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
)

func loadSource(rev *filestore.Revision) (*model.Workflow, error) {
	workflow := new(model.Workflow)

	err := workflow.Load(rev.Data)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}
