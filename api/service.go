package api

// Service ...
type Service interface {
	Run(hook string) ([]byte, error)
	Commands() []byte
}
