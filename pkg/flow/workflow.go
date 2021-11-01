package flow

import (
	"time"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/direktiv/direktiv/pkg/model"
)

const defaultDeadline = time.Second * 5

func loadSource(rev *ent.Revision) (*model.Workflow, error) {

	workflow := new(model.Workflow)

	err := workflow.Load(rev.Source)
	if err != nil {
		return nil, err
	}

	return workflow, nil

}
