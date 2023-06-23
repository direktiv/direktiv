package recipient

type RecipientType string

const (
	Server    RecipientType = "server"
	Namespace RecipientType = "namespace"
	Workflow  RecipientType = "workflow"
	Instance  RecipientType = "instance"
	Mirror    RecipientType = "mirror"
)

func (rt RecipientType) String() string {
	return string(rt)
}
