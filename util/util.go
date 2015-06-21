package util

import (
	"fmt"
	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/golog"
	"os"
	"os/exec"
)

func Panicf(msg string, args ...interface{}) {
	panic(fmt.Sprintf(msg, args...))
}

func LookPathOrAbort(binaryName string) string {
	var binary, lookErr = exec.LookPath(binaryName)
	if lookErr != nil {
		Panicf(lookErr.Error())
	}
	return binary
}

func RequireNoArguments(e interface{}, args []interface{}) {
	if len(args) > 0 {
		Panicf("%s doesn't take any arguments, got %d: %s", e, len(args), args)
	}
}

func RequireArguments(e interface{}, n int, args []interface{}) {
	if len(args) != n {
		Panicf("%s requires exactly %d argument(s), got %d: %s", e, n, len(args), args)
	}
}

func RequireArgumentsAtLeast(e interface{}, n int, args []interface{}) {
	if len(args) < n {
		Panicf("%s requires at least %d argument(s), got %d: %s", e, n, len(args), args)
	}
}

var Logger log.Logger = golog.New(os.Stdout, log.Info)

type CommandRunner interface {
	RunWithStdinGetOutputOrPanic(name string, args []string) []byte
	RunGetOutputOrPanic(name string, args []string) []byte
	RunGetOutput(name string, args []string) ([]byte, error)
}

type RealCommandRunner struct{}

func outputOrPanic(cmd *exec.Cmd) []byte {
	Logger.Debugf("Executing: %s", cmd.Args)
	if output, err := cmd.CombinedOutput(); err == nil {
		return output
	} else {
		panic(err.Error())
	}
}

func (c RealCommandRunner) RunWithStdinGetOutputOrPanic(name string, args []string) []byte {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	return outputOrPanic(cmd)
}

func (c RealCommandRunner) RunGetOutputOrPanic(name string, args []string) []byte {
	return outputOrPanic(exec.Command(name, args...))
}

func (c RealCommandRunner) RunGetOutput(name string, args []string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}
