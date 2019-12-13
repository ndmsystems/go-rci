package rci

import (
	"encoding/json"

	rciApi "github.com/tdx/go-rci/api"
)

//
func (s *svc) addBuiltInHooks() {
	s.addCommandPing()
	s.addCommandDescribeAPI()
}

//
func (s *svc) addCommandPing() {
	cmd := &rciApi.Hook{
		Hook: "/rci/ping",
		Name: "RCI ping",
		Type: rciApi.CommandTypeShellScript,
		Data: rciApi.HookData{
			Execute: []string{
				"echo pong",
			},
		},
	}

	s.hooks[cmd.Hook] = cmd
}

//
func (s *svc) addCommandDescribeAPI() {
	cmd := &rciApi.Hook{
		Hook: "/rci/describe-api",
		Name: "Describe API",
		Type: rciApi.CommandTypeBuiltIn,
		Data: rciApi.HookData{
			BuiltIn: s.describeAPI,
		},
	}

	s.hooks[cmd.Hook] = cmd
}

//
func (s *svc) describeAPI() ([]byte, error) {
	s.mu.RLock()
	json, err := json.Marshal(s.hooks)
	s.mu.RUnlock()

	return json, err
}
