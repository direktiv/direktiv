package util

const (
	MirrorActivityTypeInit        = "init"
	MirrorActivityTypeReconfigure = "reconfigure"
	MirrorActivityTypeLocked      = "locked"
	MirrorActivityTypeUnlocked    = "unlocked"
	MirrorActivityTypeCronSync    = "scheduled-sync"
	MirrorActivityTypeSync        = "sync"
)

const (
	MirrorActivityStatusComplete  = "complete"
	MirrorActivityStatusPending   = "pending"
	MirrorActivityStatusExecuting = "executing"
	MirrorActivityStatusFailed    = "failed"
)
