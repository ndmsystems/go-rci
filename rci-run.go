package rci

import (
	"bytes"
	"fmt"
	"os/exec"

	rciApi "github.com/tdx/go-rci/api"
)

// Run ...
func (s *svc) Run(hook string) error {

	s.mu.RLock()
	cmd, ok := s.hooks[hook]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("hook '%s' does not exist", hook)
	}

	switch cmd.Type {
	case rciApi.CommandTypeShellScript:
		return s.runShellScript(cmd)
	default:
		return fmt.Errorf("unsupported command type '%s'", cmd.Type)
	}
}

//
func (s *svc) runShellScript(command *rciApi.Hook) error {
	if len(command.Data.Execute) < 1 {
		return fmt.Errorf("empty 'execute' of hook '%s'", command.Hook)
	}

	cmd := exec.Command("sh", "-c", command.Data.Execute[0]+" 2>&1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	s.log.Info().Println(
		s.tag, "hook:", command.Hook, "result:",
		string(bytes.TrimSuffix(output, []byte{10})))

	return nil
}
