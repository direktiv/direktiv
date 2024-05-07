package recipient

//nolint:revive
type RecipientType string

const (
	Server    RecipientType = "server"
	Namespace RecipientType = "namespace"
	Workflow  RecipientType = "workflow"
	Instance  RecipientType = "instance"
	Mirror    RecipientType = "activity"
	Route     RecipientType = "route"
)

func (rt RecipientType) String() string {
	return string(rt)
}
