package interfaces

import (
	"github.com/abesto/easyssh/target"
	"fmt"
)

type Discoverer interface {
	HasSetArgs
	fmt.Stringer
	Discover(input string) []string
}

type HasSetArgs interface {
	SetArgs(args []interface{})
}

type TargetFilter interface {
	HasSetArgs
	fmt.Stringer
	Filter(targets []target.Target) []target.Target
}

type Executor interface {
	HasSetArgs
	fmt.Stringer
	Exec(targets []target.Target, command []string)
}
