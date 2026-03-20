package matr

import (
	"fmt"
	"os/exec"
)

type Cmd struct {
	*exec.Cmd
}

// Sh is a helper function for executing shell commands.
func Sh(cmdStr string, args ...any) *Cmd {
	cmdStr = fmt.Sprintf(cmdStr, args...)
	c := exec.Command("sh", "-c", cmdStr)
	return &Cmd{Cmd: c}
}
