package api

// Service ...
type Service interface {
	Run(token []byte, hook string, args map[string]string) ([]byte, error)
	Register(hook string, hookCommand *Hook) error
}
