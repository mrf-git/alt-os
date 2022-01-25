package exe

import (
	"bytes"
	"errors"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func makeExecEnv() []string {
	var execEnv []string
	goBinPath := os.Getenv("GOPATH")
	if goBinPath == "" {
		goBinPath = build.Default.GOPATH
	}
	if goBinPath == "" {
		execEnv = os.Environ()
	} else {
		goBinPath = filepath.Clean(filepath.Join(goBinPath, "bin"))
	}
	addedPath := false
	for _, val := range os.Environ() {
		if strings.HasPrefix(val, "PATH=") {
			existingPath := string(val[5:])
			newPath := goBinPath + string(os.PathListSeparator) + existingPath
			execEnv = append(execEnv, "PATH="+newPath)
			addedPath = true
			os.Setenv("PATH", newPath)
		} else {
			execEnv = append(execEnv, val)
		}
	}
	if !addedPath {
		execEnv = append(execEnv, "PATH="+goBinPath)
		os.Setenv("PATH", goBinPath)
	}
	return execEnv
}

// Doexec executes the command with the specified arguments and returns the output.
func Doexec(dir, name string, args ...string) (stdOutStr, stdErrStr string, err error) {
	outBuff := bytes.NewBuffer(nil)
	errBuff := bytes.NewBuffer(nil)
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = makeExecEnv()
	cmd.Stdout = outBuff
	cmd.Stderr = errBuff
	err = cmd.Run()
	stdOutStr = outBuff.String()
	stdErrStr = errBuff.String()
	return
}

// ErrOutput returns an error for displaying the specified stdOut and stdErr values.
func ErrOutput(stdOut, stdErr string, err error) error {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	return errors.New("\n---stdout---\n" + stdOut + "\n\n---stderr---\n" + stdErr + "\n\n---err---\n" + errStr)
}
