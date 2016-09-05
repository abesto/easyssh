package executors

import (
	"fmt"
	"testing"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

func TestIfOneTargetStringViaMake(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		input := "(if-one-target (external ssh) (external tmux-cssh))"
		structs := "[if-one-target [external ssh] [external tmux-cssh]]"
		final := "<if-one-target <external [ssh]> <external [tmux-cssh]>>"
		l.ExpectDebugf("MakeFromString %s -> %s", input, structs)
		l.ExpectDebugf("Make %s -> %s", "[external ssh]", "<external [ssh]>")
		l.ExpectDebugf("Make %s -> %s", "[external tmux-cssh]", "<external [tmux-cssh]>")
		l.ExpectDebugf("Make %s -> %s", structs, final)
		Make(input)
	})
}

func TestIfOneTargetMakeWithoutArgument(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(if-one-target)", fmt.Sprintf("[if-one-target]"))
		util.ExpectPanic(t, fmt.Sprintf("<if-one-target %v %v> requires exactly 2 argument(s), got 0: []", nil, nil),
			func() { Make(fmt.Sprintf("(if-one-target)")) })
	})
}

func TestIfOneTargetMakeWithOneArgument(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(if-one-target (external ssh))", fmt.Sprintf("[if-one-target [external ssh]]"))
		util.ExpectPanic(t, fmt.Sprintf("<if-one-target %v %v> requires exactly 2 argument(s), got 1: [[external ssh]]", nil, nil),
			func() { Make(fmt.Sprintf("(if-one-target (external ssh))")) })
	})
}

func TestIfOneTargetMakeWithTooManyArguments(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(if-one-target foo bar baz)", "[if-one-target foo bar baz]")
		util.ExpectPanic(t, fmt.Sprintf("<if-one-target %v %v> requires exactly 2 argument(s), got 3: [foo bar baz]", nil, nil),
			func() { Make("(if-one-target foo bar baz)") })
	})
}

func TestIfOneTargetExecWithoutSetArgs(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		util.ExpectPanic(t, fmt.Sprintf("<if-one-target %v %v> requires exactly 2 argument(s), got 0: []", nil, nil),
			func() { (&ifOneTarget{}).Exec([]target.Target{}, []string{}) })
	})
}

func TestIfOneTargetSetArgs(t *testing.T) {
	input := "(if-one-target (ssh-exec) (csshx))"
	e := Make(input).(*ifOneTarget)
	if fmt.Sprintf("%s", e.one) != "<assert-command <external-sequential [ssh]>>" {
		t.Error("one", e.one)
	}
	if fmt.Sprintf("%s", e.more) != "<assert-no-command <external-interactive [csshx]>>" {
		t.Error("more", e.more)
	}
	if fmt.Sprintf("%s", e.initialArgs) != "[[ssh-exec] [csshx]]" {
		t.Error("initialArgs", e.initialArgs)
	}
}

func TestIfOneTargetGetsOneTarget(t *testing.T) {
	withMockInMakerMap(func() {
		e := Make("(if-one-target (mock) (mock))").(*ifOneTarget)
		targets := target.FromStrings("foo")
		command := []string{"ssh", "-l", "root"}

		one := e.one.(*mockExecutor)
		one.On("Exec", targets, command).Times(1)

		more := e.more.(*mockExecutor)

		e.Exec(targets, command)
		one.AssertExpectations(t)
		more.AssertExpectations(t)
	})
}

func TestIfOneTargetGetsMultipleTargets(t *testing.T) {
	withMockInMakerMap(func() {
		e := Make("(if-one-target (mock) (mock))").(*ifOneTarget)
		targets := target.FromStrings("bar", "baz")
		command := []string{"ls", "/"}

		one := e.one.(*mockExecutor)

		more := e.more.(*mockExecutor)
		more.On("Exec", targets, command).Times(1)

		e.Exec(targets, command)
		one.AssertExpectations(t)
		more.AssertExpectations(t)
	})
}
