package e2e

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

// PowerShellCommand is a wrapper to run something in powershell.  It takes a command and args.
// And it returns stdout as a string
func PowerShellCommand(commandName string, args []string) (string, error) {
	cmdArgs := append([]string{commandName}, args...)
	stdout := bytes.Buffer{}
	psCommand := exec.Command("powershell", cmdArgs...)
	mw := io.MultiWriter(os.Stdout, &stdout)
	psCommand.Stdout = mw
	psCommand.Stderr = os.Stderr
	err := psCommand.Run()
	return stdout.String(), err
}
