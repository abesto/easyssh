package filters

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

type external struct {
	initialArgs   []interface{}
	argv          []string
	commandRunner util.CommandRunner
	tmpFileMaker  tmpFileMaker
}

type tmpFileMaker interface {
	make(dir, prefix string) (*os.File, error)
}
type realTmpFileMaker struct{}

func (m *realTmpFileMaker) make(dir, prefix string) (*os.File, error) {
	return ioutil.TempFile(dir, prefix)
}

func (f *external) Filter(targets []target.Target) []target.Target {
	util.RequireArgumentsAtLeast(f, 1, f.initialArgs)
	tmpFile, err := f.tmpFileMaker.make("", "easyssh")
	defer os.Remove(tmpFile.Name())
	if err != nil {
		util.Panicf(err.Error())
	}
	tmpFile.Write([]byte(strings.Join(target.SSHTargets(targets), "\n")))
	output := f.commandRunner.CombinedOutputWithStdinOrPanic(os.Stdin, f.argv[0], append(f.argv[1:], tmpFile.Name()))
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	newTargets := make([]target.Target, len(lines))
	for i := 0; i < len(lines); i++ {
		newTargets[i] = target.FromString(lines[i])
	}
	return newTargets
}

func (f *external) SetArgs(args []interface{}) {
	util.RequireArgumentsAtLeast(f, 1, args)
	f.initialArgs = args
	f.argv = make([]string, len(args))
	for i := 0; i < len(args); i++ {
		f.argv[i] = string(args[i].([]uint8))
	}
}

func (f *external) String() string {
	return fmt.Sprintf("<%s %s>", nameExternal, f.argv)
}
