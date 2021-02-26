package rci

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	rciApi "github.com/tdx/go-rci/api"

	"github.com/lithammer/shortuuid"
)

type cmdState struct {
	Command  string `json:"command"`
	Pid      int    `json:"pid"`
	Finished bool   `json:"finished"`
	Err      string `json:"error"`
}

type cmdResult struct {
	Finished bool     `json:"finished"`
	Err      string   `json:"error"`
	Log      []string `json:"log"`
}

type startOk struct {
	UID string `json:"uid"`
}

type hookFailed struct {
	UID   string `json:"uid"`
	Where string `json:"where"`
	Err   string `json:"error"`
}

func (s *svc) runShellScriptAsync(
	token []byte, hook *rciApi.Hook, args map[string]string) ([]byte, error) {

	// check result argument
	uid := args["result"]
	if uid != "" {
		return s.result(uid)
	}

	if len(hook.Data.Execute) < 1 {
		return nil, fmt.Errorf("empty 'execute' of hook '%s'", hook.Hook)
	}

	uid = shortuuid.New()
	logFile := filepath.Join(s.pathLocal, "async", uid+".log")
	stateFile := filepath.Join(s.pathLocal, "async", uid+".json")

	script := hook.Data.Execute[0] + " >" + logFile
	cmd := exec.Command("sh", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return failed(uid, "script start", err)
	}

	pid := cmd.Process.Pid
	err := scriptStarted(script, stateFile, pid)

	go func() {
		err := cmd.Wait()
		scriptFinished(script, stateFile, pid, err)
	}()

	if err != nil {
		return failed(uid, "write state file", err)
	}

	return startSuccess(uid)
}

//
// returns
//
func (s *svc) result(uid string) ([]byte, error) {
	stateFile := filepath.Join(s.pathLocal, "async", uid+".json")
	cmdState, err := readCommandState(stateFile)
	if err != nil {
		return failed(uid, "read state file", err)
	}

	logFile := filepath.Join(s.pathLocal, "async", uid+".log")
	logData, err := ioutil.ReadFile(logFile)
	if err != nil {
		return failed(uid, "read log file", err)
	}

	return json.Marshal(&cmdResult{
		Finished: cmdState.Finished,
		Err:      cmdState.Err,
		Log:      strings.Split(string(logData), "\n"),
	})
}

func failed(uid, where string, err error) ([]byte, error) {
	return json.Marshal(&hookFailed{
		UID:   uid,
		Where: where,
		Err:   merror(err),
	})
}

func startSuccess(uid string) ([]byte, error) {
	return json.Marshal(&startOk{
		UID: uid,
	})
}

//
// state file
//
func scriptStarted(cmd, fileName string, pid int) error {
	return writeCommandState(cmd, fileName, pid, false, nil)
}

func scriptFinished(cmd, fileName string, pid int, err error) {
	writeCommandState(cmd, fileName, pid, true, err)
}

func writeCommandState(
	cmd, fileName string, pid int, finished bool, err error) error {

	s := cmdState{
		Pid:      pid,
		Finished: finished,
		Err:      merror(err),
	}

	data, err := json.Marshal(&s)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, data, 0660)
}

func readCommandState(fileName string) (*cmdState, error) {

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	s := cmdState{}
	if err = json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func merror(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

//
func chkAsync(path string, info os.FileInfo) error {

	if strings.HasSuffix(path, ".json") {
		state, err := readCommandState(path)
		if err != nil {
			return fmt.Errorf("readCommandState failed: %w", err)
		}
		if !state.Finished {
			// check process is running and update finished if process stopped
			process, err := os.FindProcess(state.Pid)
			if err != nil {
				return fmt.Errorf("find process %d failed: %w", state.Pid, err)
			}
			if err = process.Signal(syscall.Signal(0)); err != nil {
				// finished
				var err2 error
				if state.Err != "" {
					err2 = errors.New(state.Err)
				}
				scriptFinished(state.Command, path, state.Pid, err2)
			}
		}
	}

	if time.Now().Sub(info.ModTime()) > tenHours {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove file failed: %w", err)
		}
	}

	return nil
}
