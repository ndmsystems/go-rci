package rci

import (
	"fmt"
	"os/exec"

	rciApi "github.com/tdx/go-rci/api"
)

// Run ...
func (s *svc) Run(hook string) ([]byte, error) {

	s.mu.RLock()
	cmd, ok := s.hooks[hook]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("hook '%s' does not exist", hook)
	}

	switch cmd.Type {
	case rciApi.CommandTypeShellScript:
		return s.runShellScript(cmd)
	case rciApi.CommandTypeBuiltIn:
		return s.runBuiltIn(cmd)
	default:
		return nil, fmt.Errorf("unsupported command type '%s'", cmd.Type)
	}
}

//
func (s *svc) runShellScript(command *rciApi.Hook) ([]byte, error) {
	if len(command.Data.Execute) < 1 {
		return nil, fmt.Errorf("empty 'execute' of hook '%s'", command.Hook)
	}

	cmd := exec.Command("sh", "-c", command.Data.Execute[0]+" 2>&1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	// s.log.Info().Println(
	// 	s.tag, "hook:", command.Hook, "result:",
	// 	string(bytes.TrimSuffix(output, []byte{10})))

	return output, nil
}

//
func (s *svc) runBuiltIn(command *rciApi.Hook) ([]byte, error) {
	if command.Data.BuiltIn == nil {
		return nil, fmt.Errorf("built-in hook '%s' is nil", command.Hook)
	}

	return command.Data.BuiltIn()
}
