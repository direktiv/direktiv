package flow

import (
	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/direktiv/direktiv/pkg/model"
)

func loadSource(rev *ent.Revision) (*model.Workflow, error) {

	workflow := new(model.Workflow)

	err := workflow.Load(rev.Source)
	if err != nil {
		return nil, err
	}

	return workflow, nil

}
