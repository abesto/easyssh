package discoverers

import (
	"fmt"
	"testing"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

func TestFirstMatchingStringViaMake(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		input := "(first-matching (comma-separated) (comma-separated))"
		structs := "[first-matching [comma-separated] [comma-separated]]"
		final := "<first-matching [<separated-by ,> <separated-by ,>]>"
		l.ExpectDebugf("MakeFromString %s -> %s", input, structs)
		l.ExpectDebugf("Transform: %s -> %s", "[comma-separated]", "[separated-by ,]")
		l.ExpectDebugf("Transform: %s -> %s", "[comma-separated]", "[separated-by ,]")
		l.ExpectDebugf("Make %s -> %s", "[separated-by ,]", "<separated-by ,>")
		l.ExpectDebugf("Make %s -> %s", "[separated-by ,]", "<separated-by ,>")
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
	if len(f.children) != 2 || f.children[0].String() != "<separated-by ,>" || f.children[1].String() != "<separated-by ,>" {
		t.Error("children", f.children)
	}
	if len(f.args) != 2 || fmt.Sprintf("%s", f.args[0]) != "[comma-separated]" || fmt.Sprintf("%s", f.args[1]) != "[comma-separated]" {
		t.Error("args", len(f.args), fmt.Sprintf("%s", f.args))
	}
}

func TestFirstMatchingOperation(t *testing.T) {
	f := Make("(first-matching (const foo) (const a b) (const c d))").(*firstMatching)
	// Hack to test skipping a non-matching discoverer
	f.children[0].(*fixed).retval = []target.Target{}

	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("Trying discoverer %s", "<fixed []>")
		l.ExpectDebugf("Trying discoverer %s", "<fixed [a b]>")
		target.AssertTargetListEquals(t, target.FromStrings("a", "b"), f.Discover("irrelevant string"))
	})
}
