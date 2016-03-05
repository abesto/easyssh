package executors

import (
	"fmt"
	"testing"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

var externalDefs = []struct {
	name        string
	mode        externalMode
	interactive bool
}{
	{nameExternal, externalModeSingleRun, false},
	{nameExternalInteractive, externalModeSingleRun, true},
	{nameExternalParallel, externalModeParallel, false},
	{nameExternalSequential, externalModeSequential, false},
	{nameExternalSequentialInteractive, externalModeSequential, true},
}

func TestExternalStringViaMake(t *testing.T) {
	for _, item := range externalDefs {
		name := item.name
		util.WithLogAssertions(t, func(l *util.MockLogger) {
			input := fmt.Sprintf("(%s ssh)", name)
			structs := fmt.Sprintf("[%s ssh]", name)
			final := fmt.Sprintf("<%s [ssh]>", name)
			l.ExpectDebugf("MakeFromString %s -> %s", input, structs)
			l.ExpectDebugf("Make %s -> %s", structs, final)
			executor := Make(input).(*external)
			if executor.interactive != item.interactive {
				t.Errorf("executor.interactive is %t for %s, expected %t", executor.interactive, name, item.interactive)
			}
			if executor.mode != item.mode {
				t.Errorf("executor.mode is %v for %s, expected %v", executor.mode, name, item.mode)
			}
		})
	}
}

func TestExternalMakeWithoutArgument(t *testing.T) {
	for _, item := range externalDefs {
		name := item.name
		util.WithLogAssertions(t, func(l *util.MockLogger) {
			l.ExpectDebugf("MakeFromString %s -> %s", fmt.Sprintf("(%s)", name), fmt.Sprintf("[%s]", name))
			util.ExpectPanic(t, fmt.Sprintf("<%s %s> requires at least 1 argument(s), got 0: []", name, []string{}),
				func() { Make(fmt.Sprintf("(%s)", name)) })
		})
	}
}

func TestExternalExecWithoutSetArgs(t *testing.T) {
	for _, item := range externalDefs {
		name := item.name
		util.WithLogAssertions(t, func(l *util.MockLogger) {
			util.ExpectPanic(t, fmt.Sprintf("<%s %s> requires at least 1 argument(s), got 0: []", name, []string{}),
				func() {
					(&external{mode: item.mode, interactive: item.interactive}).Exec([]target.Target{}, []string{})
				})
		})
	}
}

func TestExternalSetArgs(t *testing.T) {
	for _, item := range externalDefs {
		input := fmt.Sprintf("(%s ssh)", item.name)
		e := Make(input).(*external)
		if fmt.Sprintf("%s", e.initialArgs) != "[ssh]" {
			t.Error("initialArgs", e.initialArgs)
		}
	}
}

func TestExternalMakeSingleRunJob(t *testing.T) {
	targets := target.GivenTargets("foo", "bar")
	command := []string{"ls"}
	for _, item := range externalDefs {
		executor := Make(fmt.Sprintf("(%s csshx -l root)", item.name)).(*external)
		job := executor.makeSingleRunJob(targets, command)
		if job.Interactive != item.interactive {
			t.Errorf("job.Interactive for output of %s.makeSingleRunJob is %t, expected %t", item.name, job.Interactive, item.interactive)
		}
		expectedLabel := "foo bar"
		if job.Label != expectedLabel {
			t.Errorf("job.Label for output of %s.makeSingleRunJob is \"%s\", expected \"%s\"", item.name, job.Label, expectedLabel)
		}
		expectedArgv := []string{"csshx", "-l", "root", "foo", "bar", "ls"}
		util.AssertStringListEquals(t, expectedArgv, job.Argv)
	}
}

func TestExternalMakeJobPerTarget(t *testing.T) {
	targets := target.GivenTargets("foo", "bar")
	command := []string{"ls"}
	for _, item := range externalDefs {
		executor := Make(fmt.Sprintf("(%s ssh)", item.name)).(*external)
		jobs := executor.makeJobPerTarget(targets, command)
		if len(jobs) != len(targets) {
			t.Errorf("Expected to get %d jobs from %s, got %d instead: %v", len(targets), item.name, len(jobs), jobs)
		}
		for i, target := range targets {
			job := jobs[i]
			if job.Interactive != item.interactive {
				t.Errorf("job.Interactive for output of %s.makeJobPerTarget is %t, expected %t", item.name, job.Interactive, item.interactive)
			}
			expectedLabel := target.FriendlyName()
			if job.Label != expectedLabel {
				t.Errorf("job.Label for output of %s.makeJobPerTarget is \"%s\", expected \"%s\"", item.name, job.Label, expectedLabel)
			}
			expectedArgv := []string{"ssh", target.SSHTarget(), "ls"}
			util.AssertStringListEquals(t, expectedArgv, job.Argv)
		}
	}
}

func TestExternalExec(t *testing.T) {
	targets := target.GivenTargets("foo", "bar")
	command := []string{"ls"}
	for _, item := range externalDefs {
		executor := Make(fmt.Sprintf("(%s ssh)", item.name)).(*external)
		executor.commandRunner = &util.MockInteractiveCommandRunner{}

		util.WithLogAssertions(t, func(l *util.MockLogger) {
			m := executor.commandRunner.(*util.MockInteractiveCommandRunner)
			if executor.mode == externalModeSingleRun {
				m.When("Run", executor.makeSingleRunJob(targets, command)).Times(1)
			} else if executor.mode == externalModeSequential {
				for _, job := range executor.makeJobPerTarget(targets, command) {
					m.When("Run", job).Times(1)
				}
			} else if executor.mode == externalModeParallel {
				l.ExpectInfof("Parallelly executing %s on %s", "[ls]", "[foo bar]")
				m.When("RunParallel", executor.makeJobPerTarget(targets, command))
			}

			executor.Exec(targets, command)
			util.VerifyMocks(t, m, l)
		})
	}
}

func TestExternalUnknownMode(t *testing.T) {
	var mode externalMode = 128
	e := Make("(external ssh)").(*external)
	e.mode = mode
	util.ExpectPanic(t, fmt.Sprintf("Unknown externalMode %v", mode), func() {
		e.Exec([]target.Target{}, []string{})
	})
}
