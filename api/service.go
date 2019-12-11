package api

// Service ...
type Service interface {
	Run(hook string) error
	Command(hook string) *Hook
}
