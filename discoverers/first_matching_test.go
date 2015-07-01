package discoverers

import (
	"fmt"
	"testing"

	"github.com/abesto/easyssh/util"
)

func TestFirstMatchingStringViaMake(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		input := "(first-matching (comma-separated) (comma-separated))"
		structs := "[first-matching [comma-separated] [comma-separated]]"
		final := "<first-matching [<comma-separated> <comma-separated>]>"
		l.ExpectDebugf("MakeFromString %s -> %s", input, structs)
		l.ExpectDebugf("Make %s -> %s", "[comma-separated]", "<comma-separated>")
		l.ExpectDebugf("Make %s -> %s", "[comma-separated]", "<comma-separated>")
		l.ExpectDebugf("Make %s -> %s", structs, final)
		Make(input)
	})
}

func TestFirstMatchingMakeWithoutArgument(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(first-matching)", "[first-matching]")
		util.ExpectPanic(t, "<first-matching []> requires at least 1 argument(s), got 0: []",
			func() { Make("(first-matching)") })
	})
}

func TestFirstMatchingFilterWithoutSetArgs(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		util.ExpectPanic(t, "<first-matching []> requires at least 1 argument(s), got 0: []",
			func() { (&firstMatching{}).Discover("") })
	})
}

func TestFirstMatchingSetArgs(t *testing.T) {
	input := "(first-matching (comma-separated) (comma-separated))"
	f := Make(input).(*firstMatching)
	if len(f.children) != 2 || f.children[0].String() != "<comma-separated>" || f.children[1].String() != "<comma-separated>" {
		t.Error("children", f.children)
	}
	if len(f.args) != 2 || fmt.Sprintf("%s", f.args[0]) != "[comma-separated]" || fmt.Sprintf("%s", f.args[1]) != "[comma-separated]" {
		t.Error("args", len(f.args), f.args)
	}
}

//type appendString struct {
//	args           []interface{}
//	stringToAppend string
//}
//
//func (f *appendString) Filter(targets []target.Target) []target.Target {
//	util.RequireArgumentsAtLeast(f, 1, f.args)
//	for i := range targets {
//		targets[i].Host += f.stringToAppend
//	}
//	return targets
//}
//func (f *appendString) SetArgs(args []interface{}) {
//	util.RequireArgumentsAtLeast(f, 1, args)
//	f.args = args
//	f.stringToAppend = string(args[0].([]byte))
//}
//func (f *appendString) String() string {
//	return fmt.Sprintf("<%s %s>", "append-string", f.stringToAppend)
//}
//
//func TestFirstMatchingOperation(t *testing.T) {
//	discovererMakerMap["append-string"] = func() interfaces.Discoverer {
//		return &appendString{}
//	}
//
//	f := Make("(first-matching (append-string foo) (append-string bar))").(*firstMatching)
//
//	var ts []target.Target
//	util.WithLogAssertions(t, func(l *util.MockLogger) {
//		l.ExpectDebugf("Targets after discoverer %s: %s", "<append-string foo>", "[onefoo twofoo]")
//		l.ExpectDebugf("Targets after discoverer %s: %s", "<append-string bar>", "[onefoobar twofoobar]")
//		ts = f.Discover("")
//	})
//
//	if len(ts) != 2 || ts[0].Host != "onefoobar" || ts[1].Host != "twofoobar" {
//		t.Error(len(ts), ts)
//	}
//	delete(discovererMakerMap, "append-string")
//}
