package api

import "time"

// CommandType is a Rci command type
type CommandType string

const (
	// CommandTypeShellScript is a type for shell script
	CommandTypeShellScript = "shell-script-command"
	// CommandTypeBuiltIn is a type for built in command
	CommandTypeBuiltIn = "built-in-command"
)

// HookData ...
type HookData struct {
	Output  string
	Return  string
	Error   string
	Execute []string
	BuiltIn func(
		token []byte,
		hook *Hook,
		args map[string]string) ([]byte, error) `json:"-"`
}

// Hook ...
type Hook struct {
	Hook     string
	Name     string
	Menu     string
	Type     CommandType
	Sync     bool
	Data     HookData
	Size     int64     `json:"-"`
	ModTime  time.Time `json:"-"`
	FileName string    `json:"-"`
	Deleted  bool      `json:"-"`
}
