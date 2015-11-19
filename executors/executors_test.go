package executors

import (
	"testing"

	"github.com/abesto/easyssh/util"
)

func TestSupportedExecutorNames(t *testing.T) {
	util.AssertStringListEquals(t,
		[]string{"assert-command", "assert-no-command", "csshx", "external",
			"external-interactive", "external-parallel", "external-sequential",
			"external-sequential-interactive", "if-args", "if-command", "if-one-target",
			"ssh-exec", "ssh-exec-parallel", "ssh-exec-sequential", "ssh-login",
			"tmux-cssh"},
		SupportedExecutorNames())
}
