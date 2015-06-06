package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"regexp"
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

func TargetStrings(ts []Target) []string {
	var strs = []string{}
	for _, t := range ts {
		strs = append(strs, t.String())
	}
	return strs
}

type Discoverer interface {
	Discover(input string) []string
	SetArgs(args []string)
}

type CommaSeparatedDiscoverer struct{}

func (d *CommaSeparatedDiscoverer) Discover(input string) []string {
	return strings.Split(input, ",")
}
func (d *CommaSeparatedDiscoverer) SetArgs(args []string) {
	if len(args) > 0 {
		Abort("%s takes no configuration, got %s", d, args)
	}
}
func (d *CommaSeparatedDiscoverer) String() string {
	return "comma-separated"
}

type KnifeSearchDiscoverer struct{}

func (d *KnifeSearchDiscoverer) Discover(input string) []string {
	if !strings.Contains(input, ":") {
		fmt.Printf("Host lookup string doesn't contain ':', it won't match anything in a knife search node query\n")
		return []string{}
	}

	fmt.Printf("Looking up nodes with knife matching %s\n", input)

	var cmd = exec.Command("knife", "search", "node", "-F", "json", input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	var error = cmd.Run()
	if error != nil {
		fmt.Print(stderr.String())
		Abort(error.Error())
	}

	var data map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &data)

	var ips = []string{}
	for _, row := range data["rows"].([]interface{}) {
		var automatic = row.(map[string]interface{})["automatic"].(map[string]interface{})
		if cloud_v2, ok := automatic["cloud_v2"]; ok && cloud_v2 != nil {
			ips = append(ips, cloud_v2.(map[string]interface{})["public_ipv4"].(string))
		} else {
			ips = append(ips, automatic["ipaddress"].(string))
		}
	}

	return ips
}
func (d *KnifeSearchDiscoverer) SetArgs(args []string) {
	if len(args) > 0 {
		Abort("%s takes no configuration, got %s", d, args)
	}
}
func (d *KnifeSearchDiscoverer) String() string {
	return "knife"
}


type TargetFilter interface {
	Filter(targets []Target) []Target
}

type Ec2InstanceIdFilter struct{}

func (d *Ec2InstanceIdFilter) Filter(targets []Target) []Target {
	var re = regexp.MustCompile("i-[0-9a-f]{8}")
	for idx, target := range targets {
		var instanceId = re.FindString(target.host)
		if len(instanceId) > 0 {
			var cmd = exec.Command("aws", "ec2", "describe-instances", "--instance-id", instanceId, "--region", "us-east-1")
			var output, _ = cmd.Output()
			var data map[string]interface{}
			json.Unmarshal(output, &data)

			targets[idx].host = data["Reservations"].
			([]interface{})[0].
			(map[string]interface{})["Instances"].
			([]interface{})[0].
			(map[string]interface{})["PublicIpAddress"].(string)
		}
	}
	return targets
}
func (d *Ec2InstanceIdFilter) String() string {
	return "ec2-instance-id"
}

var filterMap = map[string]TargetFilter {
	"ec2-instance-id": &Ec2InstanceIdFilter{},
}

func ApplyFilters(filterNames string, targets []Target) []Target {
	for _, filterName := range strings.Split(filterNames, ":") {
		var filter = filterMap[filterName]
		targets = filter.Filter(targets)
		fmt.Printf("Targets after filter %s: %s\n", filter, targets)
	}
	return targets
}

type FirstMatchingDiscoverer struct {
	discoverers []Discoverer
}

func (d *FirstMatchingDiscoverer) Discover(input string) []string {
	var hosts []string
	for _, discoverer := range d.discoverers {
		fmt.Printf("Trying discoverer %s\n", discoverer)
		hosts = discoverer.Discover(input)
		if len(hosts) > 0 {
			return hosts
		}
	}
	return []string{}
}
func (d *FirstMatchingDiscoverer) SetArgs(args []string) {
	d.discoverers = []Discoverer{}
	for _, name := range args {
		d.discoverers = append(d.discoverers, MakeDiscoverer(name))
	}
	fmt.Printf("Will use the first discoverer returning a non-empty host set from the discoverer list %s\n", d.discoverers)
}
func (d *FirstMatchingDiscoverer) String() string {
	return "first-matching"
}

var discovererMap = map[string]Discoverer{
	"comma-separated": &CommaSeparatedDiscoverer{},
	"knife":           &KnifeSearchDiscoverer{},
	"first-matching":  &FirstMatchingDiscoverer{},
}

type Command interface {
	Exec(targets []Target, args []string)
	SetArgs(arg []string)
}

func LookPathOrAbort(binaryName string) string {
	var binary, lookErr = exec.LookPath(binaryName)
	if lookErr != nil {
		Abort(lookErr.Error())
	}
	return binary
}

type SshLoginCommand struct{}

func (c *SshLoginCommand) Exec(targets []Target, args []string) {
	if len(targets) != 1 {
		Abort("%s expects exactly one target, got %d: %s", c, len(targets), targets)
	}
	if len(args) > 0 {
		Abort("%s doesn't accept any arguments, got: %s", c, args)
	}

	var binary = LookPathOrAbort("ssh")
	var argv = []string{binary, targets[0].String()}
	fmt.Printf("Executing %s\n", argv)
	syscall.Exec(binary, argv, os.Environ())
}
func (c *SshLoginCommand) SetArgs(args []string) {
	if len(args) > 0 {
		Abort("%s doesn't take any arguments as %s:arg", c, c)
	}
}
func (c *SshLoginCommand) String() string {
	return "ssh-login"
}

type SshExecCommand struct{}

func (c *SshExecCommand) Exec(targets []Target, args []string) {
	if len(targets) < 1 {
		Abort("%s expects at least one target", c)
	}
	if len(args) < 1 {
		Abort("%s requires at least one argument", c)
	}

	for _, target := range targets {
		var binary = LookPathOrAbort("ssh")
		var cmd = exec.Command(binary, append([]string{target.String()}, args...)...)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()

		fmt.Printf("Executing %s\n", cmd.Args)
		cmd.Run()
	}
}
func (c *SshExecCommand) SetArgs(args []string) {
	if len(args) > 0 {
		Abort("%s doesn't take any arguments as %s:arg", c, c)
	}
}
func (c *SshExecCommand) String() string {
	return "ssh-exec"
}

type SshExecParallelCommand struct{}

func (c *SshExecParallelCommand) Exec(targets []Target, args []string) {
	if len(targets) < 1 {
		Abort("%s expects at least one target", c)
	}
	if len(args) < 1 {
		Abort("%s requires at least one argument", c)
	}

	fmt.Printf("Parallelly executing %s on %s\n", args, targets)
	var binary = LookPathOrAbort("ssh")
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
func (c *SshExecParallelCommand) SetArgs(args []string) {
	if len(args) > 0 {
		Abort("%s doesn't take any arguments as %s:arg", c, c)
	}
}
func (c *SshExecParallelCommand) String() string {
	return "ssh-exec-parallel"
}

type CsshxCommand struct{}

func (c *CsshxCommand) Exec(targets []Target, args []string) {
	if len(targets) < 1 {
		Abort("%s expects at least one target", c)
	}
	if len(args) > 0 {
		Abort("%s doesn't accept any arguments, got: %s", c, args)
	}

	var binary = LookPathOrAbort("csshx")
	var argv = append([]string{binary}, TargetStrings(targets)...)
	fmt.Printf("Executing %s\n", argv)
	syscall.Exec(binary, argv, os.Environ())
}
func (c *CsshxCommand) SetArgs(args []string) {
	if len(args) > 0 {
		Abort("%s doesn't take any arguments as %s:arg", c, c)
	}
}
func (c *CsshxCommand) String() string {
	return "csshx"
}

type TmuxCsshCommand struct{}

func (c *TmuxCsshCommand) Exec(targets []Target, args []string) {
	if len(targets) < 1 {
		Abort("%s expects at least one target", c)
	}
	if len(args) > 0 {
		Abort("%s doesn't accept any arguments, got: %s", c, args)
	}

	var binary = LookPathOrAbort("tmux-cssh")
	var argv = append([]string{binary}, TargetStrings(targets)...)
	fmt.Printf("Executing %s\n", argv)
	syscall.Exec(binary, argv, os.Environ())
}
func (c *TmuxCsshCommand) SetArgs(args []string) {
	if len(args) > 0 {
		Abort("%s doesn't take any arguments as %s:arg", c, c)
	}
}
func (c *TmuxCsshCommand) String() string {
	return "tmux-cssh"
}

type OneOrMore struct {
	one  Command
	more Command
}

func (c *OneOrMore) Exec(targets []Target, args []string) {
	if c.one == nil || c.more == nil {
		Abort(fmt.Sprint(&c))
	}
	if len(targets) < 1 {
		Abort(fmt.Sprintf("one-or-more expects at least one target"))
	} else if len(targets) == 1 {
		fmt.Printf("Got one target, using %s\n", c.one)
		c.one.Exec(targets, args)
	} else {
		fmt.Printf("Got more than one targets, using %s\n", c.more)
		c.more.Exec(targets, args)
	}
}
func (c *OneOrMore) SetArgs(args []string) {
	if len(args) != 2 {
		Abort("one-or-more expects exactly two command names, for example one-or-more:ssh-login:csshx")
	}
	c.one = MakeCommand(args[0])
	c.more = MakeCommand(args[1])
	fmt.Printf("Will use %s if one target host is found, and %s if more than one target host is found.\n", args[0], args[1])
}
func (c *OneOrMore) String() string {
	return fmt.Sprintf("one-or-more:%s:%s", c.one, c.more)
}

var commandMap = map[string]Command{
	"ssh-login":         &SshLoginCommand{},
	"csshx":             &CsshxCommand{},
	"ssh-exec":          &SshExecCommand{},
	"ssh-exec-parallel": &SshExecParallelCommand{},
	"tmux-cssh":         &TmuxCsshCommand{},
	"one-or-more":       &OneOrMore{},
}

func MakeCommand(input string) Command {
	var parts = strings.Split(input, ":")
	var name = parts[0]
	if _, ok := commandMap[name]; !ok {
		// TODO: Fine, what can I use instead, then?
		Abort(fmt.Sprintf("Command \"%s\" is not known", name))
	}
	var command = commandMap[name]
	if len(parts) > 1 {
		command.SetArgs(parts[1:])
	}
	return command
}

func MakeDiscoverer(input string) Discoverer {
	// TODO DRY with MakeCommand
	var parts = strings.Split(input, ":")
	var name = parts[0]
	if _, ok := discovererMap[name]; !ok {
		// TODO: Fine, what can I use instead, then?
		Abort(fmt.Sprintf("Discoverer \"%s\" is not known", name))
	}
	var discoverer = discovererMap[name]
	if len(parts) > 1 {
		discoverer.SetArgs(parts[1:])
	}
	return discoverer
}

func Abort(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(1)
}

func main() {
	var (
		DiscovererName     string
		Discoverer         Discoverer
		commandName        string
		commandNameForArgs string
		command            Command
		user               string
		filterNames        string
	)

	flag.StringVar(&user, "l", "",
		"Specifies the user to log in as on the remote machine. If empty, it will not be passed to the called SSH tool.")
	// TODO document what Discoverer mechanisms and command runners are available
	flag.StringVar(&DiscovererName, "d", "comma-separated", "")
	flag.StringVar(&commandName, "c", "ssh-login", "")
	flag.StringVar(&commandNameForArgs, "cc", "", "")
	flag.StringVar(&filterNames, "f", "", "")
	flag.Parse()

	if flag.NArg() == 0 {
		Abort("Required argument for target host lookup missing")
	}

	Discoverer = MakeDiscoverer(DiscovererName)

	var commandArgs = []string{}
	if flag.NArg() > 0 {
		commandArgs = flag.Args()[1:]
	}

	if len(commandArgs) > 0 && len(commandNameForArgs) > 0 {
		command = MakeCommand(commandNameForArgs)
	} else {
		command = MakeCommand(commandName)
	}

	var targets []Target = []Target{}
	for _, host := range Discoverer.Discover(flag.Arg(0)) {
		targets = append(targets, Target{host: host, user: user})
	}
	if len(targets) == 0 {
		Abort("No targets found")
	}
	fmt.Printf("Targets: %s\n", targets)

	if len(filterNames) > 0 {
		targets = ApplyFilters(filterNames, targets)
	}

	command.Exec(targets, commandArgs)
}
