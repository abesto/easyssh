package util

import (
	"fmt"
	"os"
	"os/exec"
)

func Abort(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(1)
}

func LookPathOrAbort(binaryName string) string {
	var binary, lookErr = exec.LookPath(binaryName)
	if lookErr != nil {
		Abort(lookErr.Error())
	}
	return binary
}
