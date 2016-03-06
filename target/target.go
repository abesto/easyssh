package target

import (
	"net"
	"strings"

	"github.com/abesto/easyssh/util"
)

/*
Target describes a single machine that easyssh will operate on.
*/
type Target struct {
	Host     string // Used to reference the host from the outside, generally an Host
	Hostname string // What the host calls itself
	IP       string
	User     string
}

func (t Target) withUser(s string) string {
	if t.User != "" {
		return t.User + "@" + s
	}
	return s
}

func firstNonEmptyString(strs ...string) string {
	for _, s := range strs {
		if s != "" {
			return s
		}
	}
	return ""
}

func (t Target) firstNonEmptyStringWithUser(strs ...string) string {
	t.verify()
	return t.withUser(firstNonEmptyString(strs...))
}

/*
SSHTarget returns the most specific network-level descriptor of the target, along with the user specified (if any).
Specifically, the first non-empty value of IP, Host
*/
func (t Target) SSHTarget() string {
	return t.firstNonEmptyStringWithUser(t.IP, t.Host)
}

/*
FriendlyName returns the most descriptive name available for the target.
Specifically, the first non-empty value of Hostname, Host, IP
*/
func (t Target) FriendlyName() string {
	return t.firstNonEmptyStringWithUser(t.Hostname, t.Host, t.IP)
}

func (t Target) String() string {
	return t.FriendlyName()
}

func (t Target) IsEmpty() bool {
	return firstNonEmptyString(t.IP, t.Host) == ""
}

func (t Target) verify() {
	if t.IsEmpty() {
		util.Panicf("At least one of Target.IP and Target.Host must be set")
	}
}

/*
SSHTargets maps Target.SSHTarget over a []Target
*/
func SSHTargets(ts []Target) []string {
	strs := []string{}
	for _, t := range ts {
		strs = append(strs, t.SSHTarget())
	}
	return strs
}

/*
FriendlyNames maps Target.FriendlyName over a []Target
*/
func FriendlyNames(ts []Target) []string {
	strs := []string{}
	for _, t := range ts {
		strs = append(strs, t.FriendlyName())
	}
	return strs
}

/*
FromString creates a Target from a string description of the form [user@]<ip|fqdn>
*/
func FromString(str string) Target {
	if len(str) == 0 {
		util.Panicf("FromString(str string) Target got an empty string")
	}
	var parts = strings.Split(str, "@")
	var hostDef string
	var target Target

	if len(parts) == 1 {
		hostDef = parts[0]
	} else if len(parts) == 2 {
		target.User = parts[0]
		hostDef = parts[1]
	} else {
		util.Panicf("FromString(str string) Target got a string containing more than one @ character")
	}

	if net.ParseIP(hostDef) != nil {
		target.IP = hostDef
	} else {
		target.Host = hostDef
	}
	target.verify()

	return target
}

/*
FromStrings maps FromString over...string
*/
func FromStrings(targetStrings ...string) []Target {
	targets := make([]Target, len(targetStrings))
	for i := 0; i < len(targetStrings); i++ {
		targets[i] = FromString(targetStrings[i])
	}
	return targets
}
