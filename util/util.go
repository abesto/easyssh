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

var Logger log.Logger
