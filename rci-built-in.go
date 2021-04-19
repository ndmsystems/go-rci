package rci

import (
	"encoding/json"
	"os"

	rciApi "github.com/tdx/go-rci/api"
)

//
func (s *svc) addBuiltInHooks() {
	s.addCommandPing()
	s.addCommandDescribeAPI()
	s.addCommandHostname()
	s.addCommandAsyncRunning()
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
func (s *svc) addCommandHostname() {
	cmd := &rciApi.Hook{
		Hook: "/rci/hostname",
		Name: "Hostname",
		Type: rciApi.CommandTypeBuiltIn,
		Data: rciApi.HookData{
			BuiltIn: s.hostname,
		},
	}

	s.hooks[cmd.Hook] = cmd
}

//
func (s *svc) addCommandAsyncRunning() {
	cmd := &rciApi.Hook{
		Hook: "/rci/async/running",
		Name: "Running async scripts",
		Type: rciApi.CommandTypeBuiltIn,
		Data: rciApi.HookData{
			BuiltIn: s.asyncRunning,
		},
	}

	s.hooks[cmd.Hook] = cmd
}

//
func (s *svc) describeAPI(
	token []byte, hook *rciApi.Hook, args map[string]string) ([]byte, error) {

	s.mu.RLock()
	json, err := json.Marshal(s.hooks)
	s.mu.RUnlock()

	return json, err
}

func (s *svc) hostname(
	token []byte, hook *rciApi.Hook, args map[string]string) ([]byte, error) {

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return []byte("{\"hostname\":\"" + hostname + "\"}"), nil
}

func (s *svc) asyncRunning(
	token []byte, hook *rciApi.Hook, args map[string]string) ([]byte, error) {

	s.runningHooksLock.RLock()
	defer s.runningHooksLock.RUnlock()

	json, err := json.Marshal(s.runningHooks)

	return json, err
}
