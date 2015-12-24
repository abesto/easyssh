package executors

import (
	"fmt"
	"testing"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
	"github.com/abesto/go-mock"
)

func TestIfCommandStringViaMake(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		input := "(if-command (external ssh) (external tmux-cssh))"
		structs := "[if-command [external ssh] [external tmux-cssh]]"
		final := "<if-command <external [ssh]> <external [tmux-cssh]>>"
		l.ExpectDebugf("MakeFromString %s -> %s", input, structs)
		l.ExpectDebugf("Make %s -> %s", "[external ssh]", "<external [ssh]>")
		l.ExpectDebugf("Make %s -> %s", "[external tmux-cssh]", "<external [tmux-cssh]>")
		l.ExpectDebugf("Make %s -> %s", structs, final)
		Make(input)
	})
}

func TestIfCommandMakeWithoutArgument(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(if-command)", fmt.Sprintf("[if-command]"))
		util.ExpectPanic(t, fmt.Sprintf("<if-command %v %v> requires exactly 2 argument(s), got 0: []", nil, nil),
			func() { Make(fmt.Sprintf("(if-command)")) })
	})
}

func TestIfCommandMakeWithOneArgument(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(if-command (external ssh))", fmt.Sprintf("[if-command [external ssh]]"))
		util.ExpectPanic(t, fmt.Sprintf("<if-command %v %v> requires exactly 2 argument(s), got 1: [[external ssh]]", nil, nil),
			func() { Make(fmt.Sprintf("(if-command (external ssh))")) })
	})
}

func TestIfCommandMakeWithTooManyArguments(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(if-command foo bar baz)", "[if-command foo bar baz]")
		util.ExpectPanic(t, fmt.Sprintf("<if-command %v %v> requires exactly 2 argument(s), got 3: [foo bar baz]", nil, nil),
			func() { Make("(if-command foo bar baz)") })
	})
}

func TestIfCommandExecWithoutSetArgs(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		util.ExpectPanic(t, fmt.Sprintf("<if-command %v %v> requires exactly 2 argument(s), got 0: []", nil, nil),
			func() { (&ifCommand{}).Exec([]target.Target{}, []string{}) })
	})
}

func TestIfCommandSetArgs(t *testing.T) {
	input := "(if-command (ssh-exec) (csshx))"
	e := Make(input).(*ifCommand)
	if fmt.Sprintf("%s", e.withCommand) != "<assert-command <external-sequential [ssh]>>" {
		t.Error("one", e.withCommand)
	}
	if fmt.Sprintf("%s", e.withoutCommand) != "<assert-no-command <external-interactive [csshx]>>" {
		t.Error("more", e.withoutCommand)
	}
	if fmt.Sprintf("%s", e.initialArgs) != "[[ssh-exec] [csshx]]" {
		t.Error("initialArgs", e.initialArgs)
	}
}

func TestIfCommandGetsCommand(t *testing.T) {
	withMockInMakerMap(func() {
		e := Make("(if-command (mock) (mock))").(*ifCommand)
		targets := target.GivenTargets("foo")
		command := []string{"ssh", "-l", "root"}

		withCommand := e.withCommand.(*mockExecutor)
		withCommand.When("Exec", targets, command).Times(1)

		withoutCommand := e.withoutCommand.(*mockExecutor)
		withoutCommand.When("Exec", mock.Any, mock.Any).Times(0)

		e.Exec(targets, command)
		util.VerifyMocks(t, withCommand, withoutCommand)
	})
}

func TestIfCommandGetsNoCommand(t *testing.T) {
	withMockInMakerMap(func() {
		e := Make("(if-command (mock) (mock))").(*ifCommand)
		targets := target.GivenTargets("foo")
		command := []string{}

		withCommand := e.withCommand.(*mockExecutor)
		withCommand.When("Exec", targets, command).Times(0)

		withoutCommand := e.withoutCommand.(*mockExecutor)
		withoutCommand.When("Exec", mock.Any, mock.Any).Times(1)

		e.Exec(targets, command)
		util.VerifyMocks(t, withCommand, withoutCommand)
	})
}
