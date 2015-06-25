package filters

import (
	"fmt"
	"os"
	"testing"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
	"github.com/maraino/go-mock"
)

func TestExternalStringViaMake(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		input := "(external grep myservice)"
		structs := "[external grep myservice]"
		final := "<external [grep myservice]>"
		l.ExpectDebugf("MakeFromString %s -> %s", input, structs)
		l.ExpectDebugf("Make %s -> %s", structs, final)
		Make(input)
	})
}

func TestExternalMakeWithoutArgument(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(external)", "[external]")
		util.ExpectPanic(t, "<external []> requires at least 1 argument(s), got 0: []",
			func() { Make("(external)") })
	})
}

func TestExternalFilterWithoutSetArgs(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		util.ExpectPanic(t, "<external []> requires at least 1 argument(s), got 0: []",
			func() { (&external{}).Filter(target.GivenTargets()) })
	})
}

func TestExternalSetArgs(t *testing.T) {
	f := Make("(external grep foobar)").(*external)
	if len(f.argv) != 2 || f.argv[0] != "grep" || f.argv[1] != "foobar" {
		t.Error("argv", f.argv)
	}
	if len(f.initialArgs) != 2 || fmt.Sprintf("%s", f.initialArgs[0]) != "grep" || fmt.Sprintf("%s", f.initialArgs[1]) != "foobar" {
		t.Error("initialArgs", len(f.initialArgs), f.initialArgs)
	}
	if f.tmpFileMaker == nil {
		t.Error("tmpFileMaker == nil")
	}
}

type mockTmpFileMaker struct {
	mock.Mock
}

func (m *mockTmpFileMaker) make(dir, prefix string) (*os.File, error) {
	ret := m.Called(dir, prefix)
	return ret.GetType(0, os.Stdin).(*os.File), ret.Error(1)
}

func TestExternalOperation(t *testing.T) {
	// This filter
	f := Make("(external grep -v bar)").(*external)
	// Will call "grep -v bar", which will return "foo\baz"
	r := &util.MockCommandRunner{}
	r.When("RunWithStdinGetOutputOrPanic", os.Stdin, "grep", []string{"-v", "bar", os.Stdin.Name()}).Return([]byte("foo\nbaz")).Times(1)
	f.commandRunner = r
	// Via this temporary file
	m := &mockTmpFileMaker{}
	m.When("make", "", "easyssh").Return(os.Stdin, nil)
	f.tmpFileMaker = m
	// When passed these targets
	input := target.GivenTargets("foo", "bar", "foobar", "baz")
	// And return these.
	expectedOutput := []target.Target{input[0], input[3]}
	output := f.Filter(input)
	if len(output) != len(expectedOutput) || output[0] != expectedOutput[0] || output[1] != expectedOutput[1] {
		t.Error(input, output, expectedOutput)
	}
	util.VerifyMocks(t, r, m)
}
