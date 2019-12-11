package api

import "time"

// CommandType is a Rci command type
type CommandType string

const (
	// CommandTypeShellScript is a type for shell script
	CommandTypeShellScript = "shell-script-command"
)

// HookData ...
type HookData struct {
	Output  string
	Return  string
	Error   string
	Execute []string
}

// Hook ...
type Hook struct {
	Hook     string
	Name     string
	Type     CommandType
	Data     HookData
	Size     int64
	ModTime  time.Time
	FileName string
	Deleted  bool
}
