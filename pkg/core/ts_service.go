package core

// Manager handles validation and generation of TypeScript workflow services.
type TSServiceManager interface {
	// Create compiles and generates the service file for a new TypeScript workflow.
	Create(cir *Circuit, namespace string, filePath string, fileType string) error
	// Update recompiles and regenerates the service file for an updated TypeScript workflow.
	Update(cir *Circuit, namespace string, filePath string, fileType string) error
	// Delete removes the generated service file for a TypeScript workflow.
	Delete(cir *Circuit, namespace string, filePath string, fileType string) error
}
