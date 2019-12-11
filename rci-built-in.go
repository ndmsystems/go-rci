package rci

import (
	rciApi "github.com/tdx/go-rci/api"
)

func (s *svc) addBuiltInHooks() {
	cmdPing := &rciApi.Hook{
		Hook: "/rci/ping",
		Name: "RCI ping",
		Type: rciApi.CommandTypeShellScript,
		Data: rciApi.HookData{
			Output: "log",
			Return: "state",
			Error:  "fail",
			Execute: []string{
				"echo pong",
			},
		},
	}

	s.hooks[cmdPing.Hook] = cmdPing
}
