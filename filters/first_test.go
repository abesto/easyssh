package filters

import (
	"testing"

	"reflect"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

func TestFirstStringViaMake(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(first)", "[first]")
		l.ExpectDebugf("Make %s -> %s", "[first]", "<first>")
		Make("(first)")
	})
}

func TestFirstMakeWithArgument(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(first foo)", "[first foo]")
		util.ExpectPanic(t, "<first> doesn't take any arguments, got 1: [foo]", func() { Make("(first foo)") })
	})
}

func TestFirstOperation(t *testing.T) {
	f := Make("(first)").(*first)

	cases := []struct {
		input          []string
		expectedOutput []string
	}{
		{input: []string{}, expectedOutput: []string{}},
		{input: []string{"one"}, expectedOutput: []string{"one"}},
		{input: []string{"foo", "bar", "baz"}, expectedOutput: []string{"foo"}},
	}

	for _, c := range cases {
		output := f.Filter(target.FromStrings(c.input...))
		expectedOutput := target.FromStrings(c.expectedOutput...)
		if len(output) != len(expectedOutput) {
			t.Error(c, output)
		}
		for i := 0; i < len(output); i++ {
			if !reflect.DeepEqual(output[i], expectedOutput[i]) {
				t.Error(i, c, expectedOutput, output)
			}
		}
	}
}
