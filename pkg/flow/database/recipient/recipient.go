package recipient

type RecipientType string

const (
	Server    RecipientType = "server"
	Namespace RecipientType = "namespace"
	Workflow  RecipientType = "workflow"
	Instance  RecipientType = "instance"
	Mirror    RecipientType = "mirror"
)

func Convert(recipientType string) (RecipientType, bool) {
	ok := false
	switch recipientType {
	case "server", "namespace", "workflow", "instance", "mirror":
		ok = true
	}
	return RecipientType(recipientType), ok
}

func (rt RecipientType) String() string {
	return string(rt)
}
