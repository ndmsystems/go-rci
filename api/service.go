package api

// Service ...
type Service interface {
	Run(token []byte, hook string) ([]byte, error)
	Register(hook string, hookCommand *Hook) error
}
