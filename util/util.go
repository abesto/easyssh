package util

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"bufio"
	"bytes"

	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/golog"
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
	CombinedOutputWithStdinOrPanic(stdin io.Reader, name string, args []string) []byte
	CombinedOutputOrPanic(name string, args []string) []byte
	Outputs(name string, args []string) CommandRunnerOutputs
}

type RealCommandRunner struct{}

func combinedOutputOrPanic(cmd *exec.Cmd) []byte {
	Logger.Debugf("Executing, bailing out if exits with non-zero: %s", cmd.Args)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return output
	}
	panic(err.Error() + "\n" + string(output))
}

func (c RealCommandRunner) CombinedOutputWithStdinOrPanic(stdin io.Reader, name string, args []string) []byte {
	cmd := exec.Command(name, args...)
	cmd.Stdin = stdin
	cmd.Env = os.Environ()
	return combinedOutputOrPanic(cmd)
}

func (c RealCommandRunner) CombinedOutputOrPanic(name string, args []string) []byte {
	return combinedOutputOrPanic(exec.Command(name, args...))
}

type CommandRunnerOutputs struct {
	Error    error
	Stderr   []byte
	Stdout   []byte
	Combined []byte
}

func multicast(c chan int, input io.Reader, outputs ...io.Writer) {
	s := bufio.NewScanner(input)
	for s.Scan() {
		//		Logger.Debugf("Multicast got line: %s", s.Text())
		for _, o := range outputs {
			o.Write(s.Bytes())
			o.Write([]byte("\n"))
		}
	}
	if s.Err() != nil {
		panic(s.Err())
	}
	//	Logger.Debugf("One multicast done")
	c <- 0
}

func (c RealCommandRunner) Outputs(name string, args []string) CommandRunnerOutputs {
	var (
		err            error
		stderrPipe     io.ReadCloser
		combinedBuffer bytes.Buffer
		stderrBuffer   bytes.Buffer
		stdoutBuffer   bytes.Buffer
		stdoutPipe     io.ReadCloser
		outputs        CommandRunnerOutputs
	)
	cmd := exec.Command(name, args...)
	if stderrPipe, err = cmd.StderrPipe(); err != nil {
		Panicf(err.Error())
	}
	if stdoutPipe, err = cmd.StdoutPipe(); err != nil {
		Panicf(err.Error())
	}

	stdoutChannel := make(chan int)
	stderrChannel := make(chan int)
	go multicast(stdoutChannel, stdoutPipe, &combinedBuffer, &stdoutBuffer)
	go multicast(stderrChannel, stderrPipe, &combinedBuffer, &stderrBuffer)

	Logger.Debugf("Executing: %s", cmd.Args)

	outputs.Error = cmd.Start()
	if outputs.Error != nil {
		return outputs
	}

	//	Logger.Debugf("Waiting for multicasts")
	<-stdoutChannel
	<-stderrChannel

	//	Logger.Debugf("Waiting for process to exit")
	outputs.Error = cmd.Wait()

	//	Logger.Debugf("All done, reading buffers and returning")
	outputs.Combined = combinedBuffer.Bytes()
	outputs.Stderr = stderrBuffer.Bytes()
	outputs.Stdout = stdoutBuffer.Bytes()

	return outputs
}
