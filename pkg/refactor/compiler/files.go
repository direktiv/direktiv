package compiler

type File struct {
	Name       string
	Scope      string
	Permission int64
}

func DefaultFile() *File {
	return &File{
		Scope:      fileScopeLocal,
		Permission: 0755,
	}
}

const (
	// fileScopeFlow      = "flow"
	// fileScopeNamespace = "namespace"
	fileScopeLocal = "local"
	// fileScopeShared    = "shared"
)

func (f *File) Validate() *Messages {
	m := newMessages()

	if f.Name == "" {
		m.addError("file name is required")
	}

	return m
}
