package executors

import (
	"fmt"
	"github.com/abesto/easyssh/from_sexp"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
	"os"
	"os/exec"
	"syscall"
)

func Make(input string) interfaces.Executor {
	return from_sexp.MakeFromString(input, makeByName).(interfaces.Executor)
}

func makeFromSExp(data []interface{}) interfaces.Executor {
	return from_sexp.Make(data, makeByName).(interfaces.Executor)
}

func makeByName(name string) interface{} {
	var c interfaces.Executor
	switch name {
	case "ssh-login":
		c = &sshLogin{}
	case "csshx":
		c = &csshx{}
	case "ssh-exec":
		c = &sshExec{}
	case "ssh-exec-parallel":
		c = &sshExecParallel{}
	case "tmux-cssh":
		c = &tmuxCssh{}
	case "if-one-target":
		c = &oneOrMore{}
	case "if-args":
		c = &ifArgs{}
	default:
		util.Abort(fmt.Sprintf("Command \"%s\" is not known", name))
	}
	return c
}

func requireExactlyOneTarget(e interfaces.Executor, targets []target.Target) {
	if len(targets) != 1 {
		util.Abort("%s expects exactly one target, got %d: %s", e, len(targets), targets)
	}
}

func requireAtLeastOneTarget(e interfaces.Executor, targets []target.Target) {
	if len(targets) < 1 {
		util.Abort("%s expects at least one target.", e)
	}
}

func requireNoCommand(e interfaces.Executor, command []string) {
	if len(command) > 0 {
		util.Abort("%s doesn't accept a command, got: %s", e, command)
	}
}

func requireCommand(e interfaces.Executor, command []string) {
	if len(command) == 0 {
		util.Abort("%s requires a command.", e)
	}
}

func requireNoArguments(e interfaces.Executor, args []interface{}) {
	if len(args) > 0 {
		util.Abort("%s doesn't take any arguments, got %d: %s", e, len(args), args)
	}
}

func requireArguments(e interfaces.Executor, n int, args []interface{}) {
	if len(args) != n {
		util.Abort("%s requires exactly %d arguments, got %d: %s", e, n, len(args), args)
	}
}

type sshLogin struct{}

func (e *sshLogin) Exec(targets []target.Target, command []string) {
	requireExactlyOneTarget(e, targets)
	requireNoCommand(e, command)

	var binary = util.LookPathOrAbort("ssh")
	var argv = []string{binary, targets[0].String()}
	fmt.Printf("Executing %s\n", argv)
	syscall.Exec(binary, argv, os.Environ())
}
func (e *sshLogin) SetArgs(args []interface{}) {
	requireNoArguments(e, args)
}
func (e *sshLogin) String() string {
	return "[ssh-login]"
}

type sshExec struct{}

func (e *sshExec) Exec(targets []target.Target, command []string) {
	requireAtLeastOneTarget(e, targets)
	requireCommand(e, command)

	for _, target := range targets {
		var binary = util.LookPathOrAbort("ssh")
		var cmd = exec.Command(binary, append([]string{target.String()}, command...)...)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()

		fmt.Printf("Executing %s\n", cmd.Args)
		cmd.Run()
	}
}
func (e *sshExec) SetArgs(args []interface{}) {
	requireNoArguments(e, args)
}
func (e *sshExec) String() string {
	return "[ssh-exec]"
}

type sshExecParallel struct{}

func (e *sshExecParallel) Exec(targets []target.Target, command []string) {
	requireAtLeastOneTarget(e, targets)
	requireCommand(e, command)

	fmt.Printf("Parallelly executing %s on %s\n", command, targets)
	var binary = util.LookPathOrAbort("ssh")
	var cmds = []*exec.Cmd{}
	for _, target := range targets {

		var cmd = exec.Command(binary, append([]string{target.String()}, command...)...)

		// TODO prefix with target ip, maybe color by node?
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()

		cmds = append(cmds, cmd)

		fmt.Printf("Executing %s\n", cmd.Args)
		cmd.Start()
	}

	for _, cmd := range cmds {
		cmd.Wait()
	}
}
func (e *sshExecParallel) SetArgs(args []interface{}) {
	requireNoArguments(e, args)
}
func (e *sshExecParallel) String() string {
	return "[ssh-exec-parallel]"
}

type csshx struct{}

func (e *csshx) Exec(targets []target.Target, command []string) {
	requireAtLeastOneTarget(e, targets)
	requireNoCommand(e, command)

	var binary = util.LookPathOrAbort("csshx")
	var argv = append([]string{binary}, target.TargetStrings(targets)...)
	fmt.Printf("Executing %s\n", argv)
	syscall.Exec(binary, argv, os.Environ())
}
func (e *csshx) SetArgs(args []interface{}) {
	requireNoArguments(e, args)
}
func (e *csshx) String() string {
	return "[csshx]"
}

type tmuxCssh struct{}

func (e *tmuxCssh) Exec(targets []target.Target, command []string) {
	requireAtLeastOneTarget(e, targets)
	requireNoCommand(e, command)

	var binary = util.LookPathOrAbort("tmux-cssh")
	var argv = append([]string{binary}, target.TargetStrings(targets)...)
	fmt.Printf("Executing %s\n", argv)
	syscall.Exec(binary, argv, os.Environ())
}
func (e *tmuxCssh) SetArgs(args []interface{}) {
	requireNoArguments(e, args)
}
func (e *tmuxCssh) String() string {
	return "[tmux-cssh]"
}

type oneOrMore struct {
	one  interfaces.Executor
	more interfaces.Executor
}

func (e *oneOrMore) Exec(targets []target.Target, command []string) {
	requireAtLeastOneTarget(e, targets)
	if len(targets) == 1 {
		fmt.Printf("Got one target, using %s\n", e.one)
		e.one.Exec(targets, command)
	} else {
		fmt.Printf("Got more than one target, using %s\n", e.more)
		e.more.Exec(targets, command)
	}
}
func (e *oneOrMore) SetArgs(args []interface{}) {
	requireArguments(e, 2, args)
	e.one = makeFromSExp(args[0].([]interface{}))
	e.more = makeFromSExp(args[1].([]interface{}))
	fmt.Printf("Will use %s if one target host is found, and %s if more than one target host is found.\n", args[0], args[1])
}
func (c *oneOrMore) String() string {
	return fmt.Sprintf("[one-or-more %s %s]", c.one, c.more)
}

type ifArgs struct {
	withArgs    interfaces.Executor
	withoutArgs interfaces.Executor
}

func (c *ifArgs) Exec(targets []target.Target, args []string) {
	if len(args) < 1 {
		fmt.Printf("Got no args, using %s\n", c.withoutArgs)
		c.withoutArgs.Exec(targets, args)
	} else {
		fmt.Printf("Got args, using %s\n", c.withArgs)
		c.withArgs.Exec(targets, args)
	}
}
func (e *ifArgs) SetArgs(args []interface{}) {
	requireArguments(e, 2, args)
	e.withArgs = makeFromSExp(args[0].([]interface{}))
	e.withoutArgs = makeFromSExp(args[1].([]interface{}))
	fmt.Printf("Will use %s if a command to run is provided, and %s if not.\n", e.withArgs, e.withoutArgs)
}
func (e *ifArgs) String() string {
	return fmt.Sprintf("[if-args %s %s]", e.withArgs, e.withoutArgs)
}
