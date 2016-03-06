package discoverers

import (
	"testing"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

func TestFixedStringViaMake(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		input := "(const \"foo\" bar)"
		l.ExpectDebugf("MakeFromString %s -> %s", input, "[const foo bar]")
		l.ExpectDebugf("Transform: %s -> %s", "[const foo bar]", "[fixed foo bar]")
		l.ExpectDebugf("Make %s -> %s", "[fixed foo bar]", "<fixed [foo bar]>")
		Make(input)
	})
}

func TestFixedMakeWithoutArgument(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(fixed)", "[fixed]")
		util.ExpectPanic(t, "<fixed []> requires at least 1 argument(s), got 0: []",
			func() { Make("(fixed)") })
	})
}

func TestFixedFilterWithoutSetArgs(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		util.ExpectPanic(t, "<fixed []> requires at least 1 argument(s), got 0: []",
			func() { (&fixed{}).Discover("") })
	})
}

func TestFixedOperation(t *testing.T) {
	input := "(fixed a b c)"
	d := Make(input).(*fixed)
	expected := target.FromStrings("a", "b", "c")
	target.AssertTargetListEquals(t, expected, d.retval)
	target.AssertTargetListEquals(t, expected, d.Discover("foobar"))
}
