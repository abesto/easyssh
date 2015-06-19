package util

import (
	"github.com/alexcesaro/log"
	"os"
	"os/exec"
)

func Abort(msg string, args ...interface{}) {
	Logger.Criticalf(msg, args...)
	os.Exit(1)
}

func LookPathOrAbort(binaryName string) string {
	var binary, lookErr = exec.LookPath(binaryName)
	if lookErr != nil {
		Abort(lookErr.Error())
	}
	return binary
}

func RequireNoArguments(e interface{}, args []interface{}) {
	if len(args) > 0 {
		Abort("%s doesn't take any arguments, got %d: %s", e, len(args), args)
	}
}

func RequireArguments(e interface{}, n int, args []interface{}) {
	if len(args) != n {
		Abort("%s requires exactly %d arguments, got %d: %s", e, n, len(args), args)
	}
}

func RequireArgumentsAtLeast(e interface{}, n int, args []interface{}) {
	if len(args) < n {
		Abort("%s requires at least %d arguments, got %d: %s", e, n, len(args), args)
	}
}
var Logger log.Logger
