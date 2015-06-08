package interfaces

import (
	"github.com/abesto/easyssh/target"
)

type Discoverer interface {
	Discover(input string) []string
	SetArgs(args []interface{})
	String() string
}

type HasSetArgs interface {
	SetArgs(args []interface{})
}

type TargetFilter interface {
	Filter(targets []target.Target) []target.Target
	SetArgs(args []interface{})
	String() string
}

type Command interface {
	Exec(targets []target.Target, args []string)
	SetArgs(args []interface{})
	String() string
}
