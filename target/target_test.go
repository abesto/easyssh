package target

import (
	"testing"

	"github.com/abesto/easyssh/util"
)

func TestString(t *testing.T) {
	cases := []struct {
		target   Target
		expected string
	}{
		{Target{"host-1", ""}, "host-1"},
		{Target{"host-2", "user-2"}, "user-2@host-2"},
	}
	for _, item := range cases {
		actual := item.target.String()
		if actual != item.expected {
			t.Errorf("Expected: %s. Actual: %s.", item.expected, actual)
		}
	}
	util.ExpectPanic(t, "Target host cannot be empty", func() { _ = Target{"", "user-3"}.String() })
}

func TestFromString(t *testing.T) {
	happyCases := []struct {
		input    string
		expected Target
	}{
		{"host-1", Target{"host-1", ""}},
		{"user-2@host-2", Target{"host-2", "user-2"}},
		{"@host-3", Target{"host-3", ""}},
	}
	sadCases := []struct {
		input    string
		panicMsg string
	}{
		{"", "FromString(str string) Target got an empty string"},
		{"@", "Target host cannot be empty"},
		{"user-1@", "Target host cannot be empty"},
		{"a@b@c", "FromString(str string) Target got a string containing more than one @ character"},
	}
	for _, happy := range happyCases {
		actual := FromString(happy.input)
		if actual != happy.expected {
			t.Errorf("Actual: %s. Expected: %s.", actual, happy.expected)
		}
	}
	for _, sad := range sadCases {
		util.ExpectPanic(t, sad.panicMsg, func() { FromString(sad.input) })
	}
}
