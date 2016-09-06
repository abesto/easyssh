package discoverers

import (
	"testing"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

/* Test separated-by via the alias (comma-separated) */

func TestCommaSeparatedStringViaMake(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(comma-separated)", "[comma-separated]")
		l.ExpectDebugf("Transform: %s -> %s", "[comma-separated]", "[separated-by ,]")
		l.ExpectDebugf("Make %s -> %s", "[separated-by ,]", "<separated-by ,>")
		d := Make("(comma-separated)")
		if d.String() != "<separated-by ,>" {
			t.Error(d)
		}
	})
}

func TestSeparatedByMakeWithoutArgument(t *testing.T) {
	util.ExpectPanic(t, "<separated-by > requires exactly 1 argument(s), got 0: []", func() {
		Make("(separated-by)")
	})
}

func TestSeparatedByMakeWithTooManyArguments(t *testing.T) {
	util.ExpectPanic(t, "<separated-by > requires exactly 1 argument(s), got 2: [foo bar]", func() {
		Make("(separated-by foo bar)")
	})
}

func TestCommaSeparatedOperation(t *testing.T) {
	d := Make("(comma-separated)")
	cases := []struct {
		input          string
		expectedOutput []target.Target
	}{
		{"", []target.Target{}},
		{"foo", target.FromStrings("foo")},
		{"alpha,beta", target.FromStrings("alpha", "beta")},
	}

	for _, c := range cases {
		target.AssertTargetListEquals(t, c.expectedOutput, d.Discover(c.input))
	}
}
