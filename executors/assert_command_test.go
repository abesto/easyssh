package executors

import (
	"fmt"
	"testing"

	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
	"github.com/maraino/go-mock"
)

var names = []string{nameAssertCommand, nameAssertNoCommand}
var namesWithRequire = []struct {
	name    string
	require bool
}{
	{nameAssertCommand, true},
	{nameAssertNoCommand, false},
}

func TestAssertCommandStringViaMake(t *testing.T) {
	for _, name := range names {
		util.WithLogAssertions(t, func(l *util.MockLogger) {
			input := fmt.Sprintf("(%s (external-sequential ssh))", name)
			structs := fmt.Sprintf("[%s [external-sequential ssh]]", name)
			final := fmt.Sprintf("<%s <external-sequential [ssh]>>", name)
			l.ExpectDebugf("MakeFromString %s -> %s", input, structs)
			l.ExpectDebugf("Make %s -> %s", "[external-sequential ssh]", "<external-sequential [ssh]>")
			l.ExpectDebugf("Make %s -> %s", structs, final)
			Make(input)
		})
	}
}

func TestAssertCommandMakeWithoutArgument(t *testing.T) {
	for _, name := range names {
		util.WithLogAssertions(t, func(l *util.MockLogger) {
			l.ExpectDebugf("MakeFromString %s -> %s", fmt.Sprintf("(%s)", name), fmt.Sprintf("[%s]", name))
			util.ExpectPanic(t, fmt.Sprintf("<%s %s> requires exactly 1 argument(s), got 0: []", name, nil),
				func() { Make(fmt.Sprintf("(%s)", name)) })
		})
	}
}

func TestAssertCommandMakeWithTooManyArguments(t *testing.T) {
	for _, name := range names {
		util.WithLogAssertions(t, func(l *util.MockLogger) {
			l.ExpectDebugf("MakeFromString %s -> %s", fmt.Sprintf("(%s foo bar)", name), fmt.Sprintf("[%s foo bar]", name))
			util.ExpectPanic(t, fmt.Sprintf("<%s %s> requires exactly 1 argument(s), got 2: [foo bar]", name, nil),
				func() { Make(fmt.Sprintf("(%s foo bar)", name)) })
		})
	}
}

func TestAssertCommandExecWithoutSetArgs(t *testing.T) {
	for _, item := range namesWithRequire {
		util.WithLogAssertions(t, func(l *util.MockLogger) {
			util.ExpectPanic(t, fmt.Sprintf("<%s %s> requires exactly 1 argument(s), got 0: []", item.name, nil),
				func() { (&assertCommand{require: item.require}).Exec([]target.Target{}, []string{}) })
		})
	}
}

func TestAssertCommandSetArgs(t *testing.T) {
	for _, item := range namesWithRequire {
		input := fmt.Sprintf("(%s (ssh-exec))", item.name)
		e := Make(input).(*assertCommand)
		if e.require != item.require {
			t.Error("require")
		}
		if fmt.Sprintf("%s", e.child) != "<assert-command <external-sequential [ssh]>>" {
			t.Error("child", e.child)
		}
		if fmt.Sprintf("%s", e.initialArgs) != "[[ssh-exec]]" {
			t.Error("initialArgs", e.initialArgs)
		}
	}
}

type mockExecutor struct {
	mock.Mock
}

func (e *mockExecutor) Exec(targets []target.Target, args []string) {
	e.Called(targets, args)
}
func (e *mockExecutor) SetArgs(args []interface{}) {
	// We don't actually want to assert on this, so no call to e.Called
}
func (e *mockExecutor) String() string {
	return "<mock>"
}

func withMockInMakerMap(f func()) {
	executorMakerMap["mock"] = func() interfaces.Executor { return &mockExecutor{} }
	f()
	delete(executorMakerMap, "mock")
}

func TestAssertCommandGetsCommand(t *testing.T) {
	withMockInMakerMap(func() {
		e := Make("(assert-command (mock))").(*assertCommand)
		targets := target.GivenTargets("foo", "bar")
		command := []string{"ssh", "-l", "root"}

		m := e.child.(*mockExecutor)
		m.When("Exec", targets, command).Times(1)

		e.Exec(targets, command)
		util.VerifyMocks(t, m)
	})
}

func TestAssertCommandGetsNoCommand(t *testing.T) {
	withMockInMakerMap(func() {
		e := Make("(assert-command (mock))").(*assertCommand)
		targets := target.GivenTargets("foo", "bar")
		command := []string{}
		util.ExpectPanic(t, "<assert-command <mock>> requires a command.", func() { e.Exec(targets, command) })
	})
}

func TestAssertNoCommandGetsNoCommand(t *testing.T) {
	withMockInMakerMap(func() {
		e := Make("(assert-no-command (mock))").(*assertCommand)
		targets := target.GivenTargets("foo", "bar")
		command := []string{}

		m := e.child.(*mockExecutor)
		m.When("Exec", targets, command).Times(1)

		e.Exec(targets, command)
		util.VerifyMocks(t, m)
	})
}

func TestAssertNoCommandGetsCommand(t *testing.T) {
	withMockInMakerMap(func() {
		e := Make("(assert-no-command (mock))").(*assertCommand)
		targets := target.GivenTargets("foo", "bar")
		command := []string{"ssh", "-l", "root"}
		util.ExpectPanic(t, "<assert-no-command <mock>> doesn't accept a command, got: [ssh -l root]", func() { e.Exec(targets, command) })
	})
}
