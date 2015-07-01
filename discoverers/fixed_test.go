package discoverers

import (
	"testing"

	"github.com/abesto/easyssh/util"
)

func TestFixedStringViaMake(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		input := "(const \"foo\" bar)"
		structs := "[const foo bar]"
		final := "<fixed [foo bar]>"
		l.ExpectDebugf("MakeFromString %s -> %s", input, structs)
		l.ExpectDebugf("Alias: %s -> %s", "const", "fixed")
		l.ExpectDebugf("Make %s -> %s", structs, final)
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
	util.AssertStringListEquals(t, []string{"a", "b", "c"}, d.retval)
	util.AssertStringListEquals(t, []string{"a", "b", "c"}, d.Discover("foobar"))
}
