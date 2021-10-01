package flow

import (
	"time"

	"github.com/vorteil/direktiv/pkg/flow/ent"
	"github.com/vorteil/direktiv/pkg/model"
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
