package rci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	rciApi "github.com/tdx/go-rci/api"
)

// Replyer ...
type Replyer interface {
	Format([]byte) ([]byte, error)
}

// Run ...
func (s *svc) Run(
	token []byte, hook string, args map[string]string) ([]byte, error) {

	s.mu.RLock()
	cmd, ok := s.hooks[hook]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("hook '%s' does not exist", hook)
	}

	switch cmd.Type {
	case rciApi.CommandTypeShellScript:
		return s.runShellScript(token, cmd, args)
	case rciApi.CommandTypeBuiltIn:
		return s.runBuiltIn(token, cmd, args)
	default:
		return nil, fmt.Errorf("unsupported command type '%s'", cmd.Type)
	}
}

//
func (s *svc) runShellScript(
	token []byte, hook *rciApi.Hook, args map[string]string) ([]byte, error) {

	if len(hook.Data.Execute) < 1 {
		return nil, fmt.Errorf("empty 'execute' of hook '%s'", hook.Hook)
	}

	// cmd := exec.Command("sh", "-c", hook.Data.Execute[0]+" 2>&1")
	cmd := exec.Command("sh", "-c", hook.Data.Execute[0])
	ret := "result"
	output, err := cmd.CombinedOutput()
	if err != nil {
		ret = "error"
		log.Println("err:", err)
		log.Println("out:", output)
		output = bytes.TrimSpace([]byte(err.Error()))
	}

	return s.formatShellScript(hook, ret, bytes.TrimSpace(output))
}

//
func (s *svc) runBuiltIn(
	token []byte, hook *rciApi.Hook, args map[string]string) ([]byte, error) {

	if hook.Data.BuiltIn == nil {
		return nil, fmt.Errorf("built-in hook '%s' is nil", hook.Hook)
	}

	return hook.Data.BuiltIn(token, hook, args)
}

//
func (s *svc) formatShellScript(
	command *rciApi.Hook,
	ret string, // "result" | "error"
	data []byte) ([]byte, error) {

	parts := strings.Split(strings.TrimPrefix(command.Hook, "/rci/"), "/")

	var buf bytes.Buffer

	buf.WriteString("{")
	for _, part := range parts {
		buf.WriteString("\"")
		buf.WriteString(part)
		buf.WriteString("\":{")
	}
	// hasNewLines := bytes.Contains(data, []byte{10})
	// if hasNewLines {
	buf.WriteString("\"" + ret + "\":[")
	lines := bytes.Split(data, []byte{10})
	for i, line := range lines {
		jsonValue, err := json.Marshal(string(line))
		if err != nil {
			buf.WriteString("error json.Marshal():" + err.Error())
			break
		}
		buf.Write(jsonValue)
		if i < len(lines)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("]")
	// } else {
	// 	buf.WriteString("\"result\":\"")
	// 	buf.Write(data)
	// 	buf.WriteString("\"")
	// }

	for i := 0; i < len(parts); i++ {
		buf.WriteString("}")
	}
	buf.WriteString("}")

	return buf.Bytes(), nil
}
