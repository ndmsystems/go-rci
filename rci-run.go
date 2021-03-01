package rci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

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
		if cmd.Sync {
			return s.runShellScript(token, cmd, args)
		}
		return s.runShellScriptAsync(token, cmd, args)
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

	cmd := exec.Command("sh", "-c", hook.Data.Execute[0]+" 2>&1")
	ret := "result"
	output, err := cmd.CombinedOutput()
	if err != nil {
		ret = "error"
		errOutput := bytes.TrimSpace([]byte(err.Error()))
		errOutput = append(errOutput, byte(10))
		errOutput = append(errOutput, output...)
		output = errOutput
	}

	return formatShellScript(hook, ret, bytes.TrimSpace(output))
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
func formatShellScript(
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

	for i := 0; i < len(parts); i++ {
		buf.WriteString("}")
	}
	buf.WriteString("}")

	return buf.Bytes(), nil
}

//
// Running hooks: allow single running script hook
//
func (s *svc) markScriptRunning(uid, cmd string) (string, error) {
	s.runningHooksLock.Lock()
	defer s.runningHooksLock.Unlock()

	now := time.Now()
	r, ok := s.runningHooks[cmd]
	if ok {
		return r.UID, fmt.Errorf("running %s",
			now.Sub(r.TS).Truncate(time.Millisecond))
	}

	s.runningHooks[cmd] = ActiveHook{
		UID: uid,
		TS:  now,
	}

	return "", nil
}

func (s *svc) remarkScriptRunning(uid, cmd string) {
	s.runningHooksLock.Lock()
	defer s.runningHooksLock.Unlock()

	if _, ok := s.runningHooks[cmd]; ok {
		return
	}

	now := time.Now()
	s.runningHooks[cmd] = ActiveHook{
		UID: uid,
		TS:  now,
	}
}

func (s *svc) unmarkScriptRunning(cmd string) {
	s.runningHooksLock.Lock()
	defer s.runningHooksLock.Unlock()

	delete(s.runningHooks, cmd)
}
