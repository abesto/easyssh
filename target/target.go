package target

import (
	"fmt"
	"github.com/abesto/easyssh/util"
	"strings"
)

type Target struct {
	Host string
	User string
}

func (t Target) String() string {
	if t.Host == "" {
		util.Abort("Target host cannot be empty")
	}
	if t.User == "" {
		return t.Host
	}
	return fmt.Sprintf("%s@%s", t.User, t.Host)
}

func TargetStrings(ts []Target) []string {
	var strs = []string{}
	for _, t := range ts {
		strs = append(strs, t.String())
	}
	return strs
}

func FromString(str string) Target {
	if len(str) == 0 {
		util.Abort("FromString(str string) Target got an empty string")
	}
	var parts = strings.Split(str, "@")
	var target Target
	if len(parts) == 1 {
		target = Target{Host: parts[0]}
	} else if len(parts) == 2 {
		target = Target{User: parts[0], Host: parts[1]}
	} else {
		util.Abort("FromString(str string) Target got a string containing more than one @ character")
	}
	return target
}
