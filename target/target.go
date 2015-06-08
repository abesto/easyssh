package target

import (
	"fmt"
	"github.com/abesto/easyssh/util"
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
