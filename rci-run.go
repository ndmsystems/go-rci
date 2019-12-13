package rci

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	rciApi "github.com/tdx/go-rci/api"
)

// Replyer ...
type Replyer interface {
	Format([]byte) ([]byte, error)
}

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

	return s.formatShellScript(command, bytes.TrimSpace(output))
}

//
func (s *svc) runBuiltIn(command *rciApi.Hook) ([]byte, error) {
	if command.Data.BuiltIn == nil {
		return nil, fmt.Errorf("built-in hook '%s' is nil", command.Hook)
	}

	return command.Data.BuiltIn()
}

//
func (s *svc) formatShellScript(
	command *rciApi.Hook, data []byte) ([]byte, error) {

	parts := strings.Split(strings.TrimPrefix(command.Hook, "/rci/"), "/")

	var buf bytes.Buffer

	buf.WriteString("{")
	for _, part := range parts {
		buf.WriteString("\"")
		buf.WriteString(part)
		buf.WriteString("\":{")
	}
	buf.WriteString("\"result\":\"")
	buf.Write(data)
	buf.WriteString("\"")
	for i := 0; i < len(parts); i++ {
		buf.WriteString("}")
	}
	buf.WriteString("}")

	return buf.Bytes(), nil
}
