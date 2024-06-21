package core

type TSServiceManager interface {
	Run(circuit *Circuit) error
}
