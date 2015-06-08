package commands

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

func Make(input string) interfaces.Command {
	return from_sexp.MakeFromString(input, makeByName).(interfaces.Command)
}

func makeFromSExp(data []interface{}) interfaces.Command {
	return from_sexp.Make(data, makeByName).(interfaces.Command)
}

func makeByName(name string) interface{} {
	var c interfaces.Command
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

type sshLogin struct{}

func (c *sshLogin) Exec(targets []target.Target, args []string) {
	if len(targets) != 1 {
		util.Abort("%s expects exactly one target, got %d: %s", c, len(targets), targets)
	}
	if len(args) > 0 {
		util.Abort("%s doesn't accept any arguments, got: %s", c, args)
	}

	var binary = util.LookPathOrAbort("ssh")
	var argv = []string{binary, targets[0].String()}
	fmt.Printf("Executing %s\n", argv)
	syscall.Exec(binary, argv, os.Environ())
}
func (c *sshLogin) SetArgs(args []interface{}) {
	if len(args) > 0 {
		util.Abort("%s doesn't take any arguments as %s:arg", c, c)
	}
}
func (c *sshLogin) String() string {
	return "[ssh-login]"
}

type sshExec struct{}

func (c *sshExec) Exec(targets []target.Target, args []string) {
	if len(targets) < 1 {
		util.Abort("%s expects at least one target", c)
	}
	if len(args) < 1 {
		util.Abort("%s requires at least one argument", c)
	}

	for _, target := range targets {
		var binary = util.LookPathOrAbort("ssh")
		var cmd = exec.Command(binary, append([]string{target.String()}, args...)...)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()

		fmt.Printf("Executing %s\n", cmd.Args)
		cmd.Run()
	}
}
func (c *sshExec) SetArgs(args []interface{}) {
	if len(args) > 0 {
		util.Abort("%s doesn't take any arguments as %s:arg", c, c)
	}
}
func (c *sshExec) String() string {
	return "[ssh-exec]"
}

type sshExecParallel struct{}

func (c *sshExecParallel) Exec(targets []target.Target, args []string) {
	if len(targets) < 1 {
		util.Abort("%s expects at least one target", c)
	}
	if len(args) < 1 {
		util.Abort("%s requires at least one argument", c)
	}

	fmt.Printf("Parallelly executing %s on %s\n", args, targets)
	var binary = util.LookPathOrAbort("ssh")
	var cmds = []*exec.Cmd{}
	for _, target := range targets {

		var cmd = exec.Command(binary, append([]string{target.String()}, args...)...)

		cmd.Stdin = os.Stdin
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
func (c *sshExecParallel) SetArgs(args []interface{}) {
	if len(args) > 0 {
		util.Abort("%s doesn't take any arguments as %s:arg", c, c)
	}
}
func (c *sshExecParallel) String() string {
	return "[ssh-exec-parallel]"
}

type csshx struct{}

func (c *csshx) Exec(targets []target.Target, args []string) {
	if len(targets) < 1 {
		util.Abort("%s expects at least one target", c)
	}
	if len(args) > 0 {
		util.Abort("%s doesn't accept any arguments, got: %s", c, args)
	}

	var binary = util.LookPathOrAbort("csshx")
	var argv = append([]string{binary}, target.TargetStrings(targets)...)
	fmt.Printf("Executing %s\n", argv)
	syscall.Exec(binary, argv, os.Environ())
}
func (c *csshx) SetArgs(args []interface{}) {
	if len(args) > 0 {
		util.Abort("%s doesn't take any arguments as %s:arg", c, c)
	}
}
func (c *csshx) String() string {
	return "[csshx]"
}

type tmuxCssh struct{}

func (c *tmuxCssh) Exec(targets []target.Target, args []string) {
	if len(targets) < 1 {
		util.Abort("%s expects at least one target", c)
	}
	if len(args) > 0 {
		util.Abort("%s doesn't accept any arguments, got: %s", c, args)
	}

	var binary = util.LookPathOrAbort("tmux-cssh")
	var argv = append([]string{binary}, target.TargetStrings(targets)...)
	fmt.Printf("Executing %s\n", argv)
	syscall.Exec(binary, argv, os.Environ())
}
func (c *tmuxCssh) SetArgs(args []interface{}) {
	if len(args) > 0 {
		util.Abort("%s doesn't take any arguments as %s:arg", c, c)
	}
}
func (c *tmuxCssh) String() string {
	return "[tmux-cssh]"
}

type oneOrMore struct {
	one  interfaces.Command
	more interfaces.Command
}

func (c *oneOrMore) Exec(targets []target.Target, args []string) {
	if c.one == nil || c.more == nil {
		util.Abort(fmt.Sprint(&c))
	}
	if len(targets) < 1 {
		util.Abort("one-or-more expects at least one target")
	} else if len(targets) == 1 {
		fmt.Printf("Got one target, using %s\n", c.one)
		c.one.Exec(targets, args)
	} else {
		fmt.Printf("Got more than one target, using %s\n", c.more)
		c.more.Exec(targets, args)
	}
}
func (c *oneOrMore) SetArgs(args []interface{}) {
	if len(args) != 2 {
		util.Abort("one-or-more expects exactly two command names, for example one-or-more:ssh-login:csshx")
	}
	c.one = makeFromSExp(args[0].([]interface{}))
	c.more = makeFromSExp(args[1].([]interface{}))
	fmt.Printf("Will use %s if one target host is found, and %s if more than one target host is found.\n", args[0], args[1])
}
func (c *oneOrMore) String() string {
	return fmt.Sprintf("[one-or-more %s %s]", c.one, c.more)
}

type ifArgs struct {
	withArgs    interfaces.Command
	withoutArgs interfaces.Command
}

func (c *ifArgs) Exec(targets []target.Target, args []string) {
	if c.withArgs == nil || c.withoutArgs == nil {
		util.Abort(fmt.Sprint(&c))
	}
	if len(args) < 1 {
		fmt.Printf("Got no args, using %s\n", c.withoutArgs)
		c.withoutArgs.Exec(targets, args)
	} else {
		fmt.Printf("Got args, using %s\n", c.withArgs)
		c.withArgs.Exec(targets, args)
	}
}
func (c *ifArgs) SetArgs(args []interface{}) {
	if len(args) != 2 {
		util.Abort("%s expects exactly two command names, for example %s:ssh-login:ssh-exec", c, c)
	}
	c.withArgs = makeFromSExp(args[0].([]interface{}))
	c.withoutArgs = makeFromSExp(args[1].([]interface{}))
	fmt.Printf("Will use %s if a command to run is provided, and %s if not.\n", c.withArgs, c.withoutArgs)
}
func (c *ifArgs) String() string {
	return fmt.Sprintf("[if-args %s %s]", c.withArgs, c.withoutArgs)
}
