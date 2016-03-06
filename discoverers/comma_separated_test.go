package discoverers

import (
	"testing"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

func TestCommaSeparatedStringViaMake(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(comma-separated)", "[comma-separated]")
		l.ExpectDebugf("Make %s -> %s", "[comma-separated]", "<comma-separated>")
		d := Make("(comma-separated)")
		if d.String() != "<comma-separated>" {
			t.Error(d)
		}
	})
}

func TestCommaSeparatedSetArgs(t *testing.T) {
	util.ExpectPanic(t, "<comma-separated> doesn't take any arguments, got 1: [foobar]", func() {
		Make("(comma-separated foobar)")
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
