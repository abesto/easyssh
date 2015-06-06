package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type Target struct {
	host string
	user string
}

func (t Target) String() string {
	if t.host == "" {
		Abort("Target host cannot be empty")
	}
	if t.user == "" {
		return t.host
	}
	return fmt.Sprintf("%s@%s", t.user, t.host)
}

type Discovery interface {
	Discover(input string) []string
}

type IdentityDiscovery struct{}

func (d IdentityDiscovery) Discover(input string) []string {
	return []string{input}
}

type KnifeSearchDiscovery struct{}

func (d KnifeSearchDiscovery) Discover(input string) []string {
	fmt.Printf("Looking up nodes with knife matching %s\n", input)

	var output, error = exec.Command("knife", "search", "node", "-F", "json", input).Output()
	if error != nil {
		Abort(error.Error())
	}

	var data map[string]interface{}
	json.Unmarshal(output, &data)

	var ips = []string{}
	for _, row := range data["rows"].([]interface{}) {
		var automatic = row.(map[string]interface{})["automatic"].(map[string]interface{})
		if cloud_v2, ok := automatic["cloud_v2"]; ok && cloud_v2 != nil {
			ips = append(ips, cloud_v2.(map[string]interface{})["public_ipv4"].(string))
		} else {
			ips = append(ips, automatic["ipaddress"].(string))
		}
	}

	fmt.Printf("Matched nodes: %s\n", ips)
	return ips
}

var discoveryMap = map[string]Discovery{
	"identity": IdentityDiscovery{},
	"knife":    KnifeSearchDiscovery{},
}

type Command interface {
	Exec(targets []Target, args []string)
}

func Ssh(args []string) (ssh string, argv []string, env []string) {
	var lookErr error
	ssh, lookErr = exec.LookPath("ssh")
	if lookErr != nil {
		Abort(lookErr.Error())
	}
	argv = append([]string{ssh}, args...)
	return ssh, argv, os.Environ()
}

type SshLoginCommand struct{}

func (c SshLoginCommand) Exec(targets []Target, args []string) {
	if len(targets) != 1 {
		Abort(fmt.Sprintf("ssh-login expects exactly one target, got %d: %s", len(targets), targets))
	}
	if len(args) > 0 {
		Abort(fmt.Sprintf("ssh-login doesn't accept any arguments, got: %s", strings.Join(args, " ")))
	}

	fmt.Printf("Logging in with interactive session to %s\n", targets[0])
	binary, argv, env := Ssh([]string{targets[0].String()})
	syscall.Exec(binary, argv, env)
}

type SshExecCommand struct{}

func (c SshExecCommand) Exec(targets []Target, args []string) {
	if len(targets) < 1 {
		Abort(fmt.Sprintf("ssh-exec expects at least one target"))
	}
	if len(args) < 1 {
		Abort(fmt.Sprintf("ssh-exec requires at least one argument"))
	}

	for _, target := range targets {
		binary, argv, env := Ssh(append([]string{target.String()}, args...))
		var cmd = exec.Command(binary, argv[1:]...)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = env

		fmt.Printf("Executing %s\n", argv)
		cmd.Run()
	}
}

type SshExecParallelCommand struct{}

func (c SshExecParallelCommand) Exec(targets []Target, args []string) {
	if len(targets) < 1 {
		Abort(fmt.Sprintf("ssh-exec expects at least one target"))
	}
	if len(args) < 1 {
		Abort(fmt.Sprintf("ssh-exec requires at least one argument"))
	}

	fmt.Printf("Parallelly executing %s on %s\n", args, targets)
	var cmds = []*exec.Cmd{}
	for _, target := range targets {
		binary, argv, env := Ssh(append([]string{target.String()}, args...))
		var cmd = exec.Command(binary, argv[1:]...)

		cmd.Stdin = os.Stdin
		// TODO prefix with target ip, maybe color by node?
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = env

		cmds = append(cmds, cmd)

		fmt.Printf("Executing %s\n", argv)
		cmd.Start()
	}

	for _, cmd := range cmds {
		cmd.Wait()
	}
}

var commandMap = map[string]Command{
	"ssh-login":         SshLoginCommand{},
	"ssh-exec":          SshExecCommand{},
	"ssh-exec-parallel": SshExecParallelCommand{},
}

func MakeCommand(name string) Command {
	if _, ok := commandMap[name]; !ok {
		// TODO: Fine, what can I use instead, then?
		Abort(fmt.Sprintf("Command \"%s\" is not known", name))
	}
	return commandMap[name]
}

func MakeDiscovery(name string) Discovery {
	// TODO DRY with MakeCommand
	if _, ok := discoveryMap[name]; !ok {
		// TODO: Fine, what can I use instead, then?
		Abort(fmt.Sprintf("Discovery \"%s\" is not known", name))
	}
	return discoveryMap[name]
}

func Abort(msg string) {
	fmt.Print(msg + "\n")
	os.Exit(1)
}

func main() {
	var (
		discoveryName string
		discovery     Discovery
		commandName   string
		command       Command
		user          string
	)

	flag.StringVar(&user, "l", "",
		"Specifies the user to log in as on the remote machine. If empty, it will not be passed to the called SSH tool.")
	// TODO document what discovery mechanisms and command runners are available
	flag.StringVar(&discoveryName, "d", "identity", "")
	flag.StringVar(&commandName, "c", "ssh-login", "")
	flag.Parse()

	if flag.NArg() == 0 {
		Abort("Required argument for target host lookup missing")
	}

	discovery = MakeDiscovery(discoveryName)
	command = MakeCommand(commandName)

	var targets []Target = []Target{}
	for _, host := range discovery.Discover(flag.Arg(0)) {
		targets = append(targets, Target{host: host, user: user})
	}

	var commandArgs = []string{}
	if flag.NArg() > 0 {
		commandArgs = flag.Args()[1:]
	}

	command.Exec(targets, commandArgs)
}
