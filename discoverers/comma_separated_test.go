package discoverers

import (
	"testing"

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
		expectedOutput []string
	}{
		{"", []string{}},
		{"foo", []string{"foo"}},
		{"alpha,beta", []string{"alpha", "beta"}},
	}

	for _, c := range cases {
		actualOutput := d.Discover(c.input)
		util.AssertStringListEquals(t, c.expectedOutput, actualOutput)
	}
}
