package api

// Service ...
type Service interface {
	Run(hook string) ([]byte, error)
	Register(hook string, hookCommand *Hook) error
}
