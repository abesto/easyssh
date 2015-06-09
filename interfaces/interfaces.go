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

type Executor interface {
	Exec(targets []target.Target, command []string)
	SetArgs(args []interface{})
	String() string
}
