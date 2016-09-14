package filters

import (
	"testing"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
	"github.com/stretchr/testify/assert"
)

func TestCoalesceStringViaMake(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		input := "(coalesce ip host hostname)"
		structs := "[coalesce ip host hostname]"
		final := "<coalesce [ip host hostname]>"
		l.ExpectDebugf("MakeFromString %s -> %s", input, structs)
		l.ExpectDebugf("Make %s -> %s", structs, final)
		Make(input)
	})
}

func TestCoalesceMakeWithoutArgument(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(coalesce)", "[coalesce]")
		util.ExpectPanic(t, "<coalesce []> requires at least 1 argument(s), got 0: []",
			func() { Make("(coalesce)") })
	})
}

func TestCoalesceFilterWithoutSetArgs(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		util.ExpectPanic(t, "<coalesce []> requires at least 1 argument(s), got 0: []",
			func() { (&coalesce{}).Filter([]target.Target{}) })
	})
}

func TestCoalesceSetArgs(t *testing.T) {
	input := "(coalesce host hostname)"
	f := Make(input).(*coalesce)
	assert.Equal(t, f.coalesceOrder, []string{"host", "hostname"})
}

func TestUnknownCoalescer(t *testing.T) {
	util.ExpectPanic(t, "Unknown target coalescer foobar (index 1) in filter <coalesce [ip foobar hostname barbaz]>", func() {
		Make("(coalesce ip foobar hostname barbaz)")
	})
}

func TestCoalesceOperation(t *testing.T) {
	f := Make("(coalesce host)").(*coalesce)
	cases := []struct {
		expected string
		target   target.Target
	}{
		{"foo", target.Target{Host: "foo", IP: "0.0.0.0", Hostname: "notfoo"}},
		{"bar", target.Target{Host: "bar"}},
	}
	for _, c := range cases {
		target := f.Filter([]target.Target{c.target})[0]
		assert.Equal(t, []string{"host"}, target.CoalesceOrder)
		assert.Equal(t, c.expected, target.SSHTarget())
	}
}
