package metrics

import (
	"github.com/direktiv/direktiv/pkg/metrics/ent"
)

type record struct {
	r *ent.Metrics
}

func (r *record) didSucceed() bool {

	if r.r.ErrorCode == "" {
		// state finished without error
		return true
	}

	if NextEnums[r.r.Next] == NextTransition {
		// error occurred but was caught
		return true
	}

	// uncaught error
	return false
}
