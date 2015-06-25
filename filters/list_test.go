package filters

import (
	"fmt"
	"testing"

	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

func TestListStringViaMake(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		input := "(list (id) (id))"
		structs := "[list [id] [id]]"
		final := "<list [<id> <id>]>"
		l.ExpectDebugf("MakeFromString %s -> %s", input, structs)
		l.ExpectDebugf("Make %s -> %s", "[id]", "<id>")
		l.ExpectDebugf("Make %s -> %s", "[id]", "<id>")
		l.ExpectDebugf("Make %s -> %s", structs, final)
		Make(input)
	})
}

func TestListMakeWithoutArgument(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(list)", "[list]")
		util.ExpectPanic(t, "<list []> requires at least 1 argument(s), got 0: []",
			func() { Make("(list)") })
	})
}

func TestListFilterWithoutSetArgs(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		util.ExpectPanic(t, "<list []> requires at least 1 argument(s), got 0: []",
			func() { (&list{}).Filter([]target.Target{}) })
	})
}

func TestListSetArgs(t *testing.T) {
	input := "(list (id) (id))"
	f := Make(input).(*list)
	if len(f.children) != 2 || f.children[0].String() != "<id>" || f.children[1].String() != "<id>" {
		t.Error("children", f.children)
	}
	if len(f.args) != 2 || fmt.Sprintf("%s", f.args[0]) != "[id]" || fmt.Sprintf("%s", f.args[1]) != "[id]" {
		t.Error("args", len(f.args), f.args)
	}
}

type appendString struct {
	args           []interface{}
	stringToAppend string
}

func (f *appendString) Filter(targets []target.Target) []target.Target {
	util.RequireArgumentsAtLeast(f, 1, f.args)
	for i := range targets {
		targets[i].Host += f.stringToAppend
	}
	return targets
}
func (f *appendString) SetArgs(args []interface{}) {
	util.RequireArgumentsAtLeast(f, 1, args)
	f.args = args
	f.stringToAppend = string(args[0].([]byte))
}
func (f *appendString) String() string {
	return fmt.Sprintf("<%s %s>", "append-string", f.stringToAppend)
}

func TestListOperation(t *testing.T) {
	filterMakerMap["append-string"] = func() interfaces.TargetFilter {
		return &appendString{}
	}

	f := Make("(list (append-string foo) (append-string bar))").(*list)

	var ts []target.Target
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("Targets after filter %s: %s", "<append-string foo>", "[onefoo twofoo]")
		l.ExpectDebugf("Targets after filter %s: %s", "<append-string bar>", "[onefoobar twofoobar]")
		ts = f.Filter(target.GivenTargets("one", "two"))
	})

	if len(ts) != 2 || ts[0].Host != "onefoobar" || ts[1].Host != "twofoobar" {
		t.Error(len(ts), ts)
	}
	delete(filterMakerMap, "append-string")
}
