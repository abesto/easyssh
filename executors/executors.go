package executors

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/abesto/easyssh/fromsexp"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/golog"
)

func Make(input string) interfaces.Executor {
	return fromsexp.MakeFromString(input, aliases, makeByName).(interfaces.Executor)
}

func SupportedExecutorNames() []string {
	var keys = make([]string, len(executorMakerMap))
	var i = 0
	for key := range executorMakerMap {
		keys[i] = key
		i++
	}
	return keys
}

func makeFromSExp(data []interface{}) interfaces.Executor {
	return fromsexp.Make(data, aliases, makeByName).(interfaces.Executor)
}

const (
	nameSshLogin        = "ssh-login"
	nameCsshx           = "knife"
	nameSshExec         = "ssh-exec"
	nameSshExecParallel = "ssh-exec-parallel"
	nameTmuxCssh        = "tmux-cssh"
	nameIfOneTarget     = "if-one-target"
	nameIfCommand       = "if-command"
)

var executorMakerMap = map[string]func() interfaces.Executor{
	nameSshLogin:        func() interfaces.Executor { return &sshLogin{} },
	nameCsshx:           func() interfaces.Executor { return &csshx{} },
	nameSshExec:         func() interfaces.Executor { return &sshExec{} },
	nameSshExecParallel: func() interfaces.Executor { return &sshExecParallel{} },
	nameTmuxCssh:        func() interfaces.Executor { return &tmuxCssh{} },
	nameIfOneTarget:     func() interfaces.Executor { return &ifOneTarget{} },
	nameIfCommand:       func() interfaces.Executor { return &ifCommand{} },
}

var aliases = fromsexp.Aliases{
	fromsexp.Alias{Name: nameIfCommand, Alias: "if-args"},
}

func makeByName(name string) interface{} {
	var d interfaces.Executor
	for key, maker := range executorMakerMap {
		if key == name {
			d = maker()
		}
	}
	if d == nil {
		util.Panicf("Executor \"%s\" is not known", name)
	}
	return d
}

func requireExactlyOneTarget(e interfaces.Executor, targets []target.Target) {
	if len(targets) != 1 {
		util.Panicf("%s expects exactly one target, got %d: %s", e, len(targets), targets)
	}
}

func requireAtLeastOneTarget(e interfaces.Executor, targets []target.Target) {
	if len(targets) < 1 {
		util.Panicf("%s expects at least one target.", e)
	}
}

func requireNoCommand(e interfaces.Executor, command []string) {
	if len(command) > 0 {
		util.Panicf("%s doesn't accept a command, got: %s", e, command)
	}
}

func requireCommand(e interfaces.Executor, command []string) {
	if len(command) == 0 {
		util.Panicf("%s requires a command.", e)
	}
}

func myExec(binaryName string, args ...string) {
	var binary = util.LookPathOrAbort(binaryName)

	var argv = append([]string{binary}, args...)
	util.Logger.Infof("Executing %s", argv)

	cmd := exec.Command(binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		util.Panicf("%s failed: %s", cmd.Args, err)
	}
}

type sshLogin struct{}

func (e *sshLogin) Exec(targets []target.Target, command []string) {
	requireExactlyOneTarget(e, targets)
	requireNoCommand(e, command)
	myExec("ssh", targets[0].String())
}
func (e *sshLogin) SetArgs(args []interface{}) {
	util.RequireNoArguments(e, args)
}
func (e *sshLogin) String() string {
	return fmt.Sprintf("<%s>", nameSshLogin)
}

type sshExec struct{}

func (e *sshExec) Exec(targets []target.Target, command []string) {
	requireAtLeastOneTarget(e, targets)
	requireCommand(e, command)

	for _, target := range targets {
		var binary = util.LookPathOrAbort("ssh")
		var cmd = makeLoggedCommand(binary, target, append([]string{target.String()}, command...))
		util.Logger.Infof("Executing %s", cmd.Args)
		cmd.Run()
	}
}
func (e *sshExec) SetArgs(args []interface{}) {
	util.RequireNoArguments(e, args)
}
func (e *sshExec) String() string {
	return fmt.Sprintf("<%s>", nameSshExec)
}

type sshExecParallel struct{}

func (e *sshExecParallel) Exec(targets []target.Target, command []string) {
	requireAtLeastOneTarget(e, targets)
	requireCommand(e, command)

	util.Logger.Infof("Parallelly executing %s on %s", command, targets)
	var binary = util.LookPathOrAbort("ssh")
	var cmds = []*exec.Cmd{}
	for _, target := range targets {
		var cmd = makeLoggedCommand(binary, target, append([]string{target.String()}, command...))
		cmds = append(cmds, cmd)
		util.Logger.Debugf("Executing %s", cmd.Args)
		cmd.Start()
	}

	for _, cmd := range cmds {
		var error = cmd.Wait()
		if error != nil {
			util.Logger.Errorf("%s: %s", cmd.Args, error)
		}
	}
}
func (e *sshExecParallel) SetArgs(args []interface{}) {
	util.RequireNoArguments(e, args)
}
func (e *sshExecParallel) String() string {
	return fmt.Sprintf("<%s>", nameSshExecParallel)
}

type csshx struct{}

func (e *csshx) Exec(targets []target.Target, command []string) {
	requireAtLeastOneTarget(e, targets)
	requireNoCommand(e, command)
	myExec("csshx", target.Strings(targets)...)
}
func (e *csshx) SetArgs(args []interface{}) {
	util.RequireNoArguments(e, args)
}
func (e *csshx) String() string {
	return fmt.Sprintf("<%s>", nameCsshx)
}

type tmuxCssh struct{}

func (e *tmuxCssh) Exec(targets []target.Target, command []string) {
	requireAtLeastOneTarget(e, targets)
	requireNoCommand(e, command)
	myExec("tmux-cssh", target.Strings(targets)...)
}
func (e *tmuxCssh) SetArgs(args []interface{}) {
	util.RequireNoArguments(e, args)
}
func (e *tmuxCssh) String() string {
	return fmt.Sprintf("<%s>", nameTmuxCssh)
}

type ifOneTarget struct {
	one  interfaces.Executor
	more interfaces.Executor
}

func (e *ifOneTarget) Exec(targets []target.Target, command []string) {
	requireAtLeastOneTarget(e, targets)
	if len(targets) == 1 {
		util.Logger.Debugf("%s got one target, using %s", e, e.one)
		e.one.Exec(targets, command)
	} else {
		util.Logger.Debugf("%s got more than one target, using %s", e, e.more)
		e.more.Exec(targets, command)
	}
}
func (e *ifOneTarget) SetArgs(args []interface{}) {
	util.RequireArguments(e, 2, args)
	e.one = makeFromSExp(args[0].([]interface{}))
	e.more = makeFromSExp(args[1].([]interface{}))
}
func (e *ifOneTarget) String() string {
	return fmt.Sprintf("<%s %s %s>", nameIfOneTarget, e.one, e.more)
}

type ifCommand struct {
	withCommand    interfaces.Executor
	withoutCommand interfaces.Executor
}

func (e *ifCommand) Exec(targets []target.Target, args []string) {
	if len(args) < 1 {
		util.Logger.Debugf("%s got no args, using %s", e, e.withoutCommand)
		e.withoutCommand.Exec(targets, args)
	} else {
		util.Logger.Debugf("%s got args, using %s", e, e.withCommand)
		e.withCommand.Exec(targets, args)
	}
}
func (e *ifCommand) SetArgs(args []interface{}) {
	util.RequireArguments(e, 2, args)
	e.withCommand = makeFromSExp(args[0].([]interface{}))
	e.withoutCommand = makeFromSExp(args[1].([]interface{}))
}
func (e *ifCommand) String() string {
	return fmt.Sprintf("<%s %s %s>", nameIfCommand, e.withCommand, e.withoutCommand)
}

type prefixedLogWriterProxy struct {
	prefix string
	logger *golog.Logger
}

func newPrefixedLogWriterProxy(prefix string, file *os.File) prefixedLogWriterProxy {
	return prefixedLogWriterProxy{prefix: prefix, logger: golog.New(file, log.Debug)}
}
func (w prefixedLogWriterProxy) Write(p []byte) (n int, err error) {
	var logger = *w.logger
	var lines = strings.Split(strings.TrimSpace(string(p)), "\n")
	for _, line := range lines {
		logger.Notice(w.prefix, line)
	}
	return len(p), nil
}

func makeLoggedCommand(binary string, target target.Target, args []string) *exec.Cmd {
	var cmd = exec.Command(binary, args...)
	var prefixStdout = fmt.Sprintf("[%s] (STDOUT)", target)
	var prefixStderr = fmt.Sprintf("[%s] (STDERR)", target)

	cmd.Stdout = newPrefixedLogWriterProxy(prefixStdout, os.Stdout)
	cmd.Stderr = newPrefixedLogWriterProxy(prefixStderr, os.Stderr)
	cmd.Env = os.Environ()

	return cmd
}
