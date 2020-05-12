package rci

import (
	"fmt"
	"net/url"

	rciApi "github.com/tdx/go-rci/api"
)

func (s *svc) Register(hook string, hookCommand *rciApi.Hook) error {

	u, err := url.Parse(hook)
	if err != nil {
		return err
	}

	path := u.Path

	s.mu.RLock()
	_, ok := s.hooks[path]
	s.mu.RUnlock()

	if ok {
		return fmt.Errorf("hook '%s' already registered", path)
	}

	s.mu.Lock()
	s.hooks[path] = hookCommand
	s.mu.Unlock()

	return nil
}
