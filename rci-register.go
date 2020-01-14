package rci

import (
	"fmt"

	rciApi "github.com/tdx/go-rci/api"
)

func (s *svc) Register(hook string, hookCommand *rciApi.Hook) error {

	s.mu.RLock()
	_, ok := s.hooks[hook]
	s.mu.RUnlock()

	if ok {
		return fmt.Errorf("hook '%s' already registered", hook)
	}

	s.mu.Lock()
	s.hooks[hook] = hookCommand
	s.mu.Unlock()

	return nil
}
