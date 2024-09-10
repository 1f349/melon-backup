package utils

import (
	"os"
	"os/exec"
)

func CreateCmd(cmdAArgs []string, cmdEnv ...string) *exec.Cmd {
	return CreateCmdOnlyPassedEnv(cmdAArgs, append(os.Environ(), cmdEnv...)...)
}

func CreateCmdOnlyPassedEnv(cmdAArgs []string, cmdEnv ...string) *exec.Cmd {
	cmd := exec.Command(cmdAArgs[0], cmdAArgs[1:]...)
	cmd.Dir = GetCWD()
	cmd.Env = cmdEnv
	return cmd
}
