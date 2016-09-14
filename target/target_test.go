package target

import (
	"testing"

	"reflect"

	"github.com/abesto/easyssh/util"
)

func TestSSHTarget(t *testing.T) {
	cases := []struct {
		target   Target
		expected string
	}{
		{Target{Host: "host-1.test", Hostname: "host-1", IP: "1.1.1.1"}, "1.1.1.1"},
		{Target{Host: "host-2.test", Hostname: "host-2"}, "host-2.test"},
		{Target{Host: "host-3.test", Hostname: "host-3", IP: "3.3.3.3", User: "user-3"}, "user-3@3.3.3.3"},
		{Target{Host: "host-4.test", Hostname: "host-4", User: "user-4"}, "user-4@host-4.test"},
	}
	for _, item := range cases {
		actual := item.target.SSHTarget()
		if actual != item.expected {
			t.Errorf("Expected: %s. Actual: %s.", item.expected, actual)
		}
	}
	util.ExpectPanic(t, "At least one of Target.IP and Target.Host must be set", func() { _ = Target{User: "user-3"}.SSHTarget() })
}

func TestFriendlyName(t *testing.T) {
	cases := []struct {
		target   Target
		expected string
	}{
		{Target{Host: "host-1.test", Hostname: "host-1", IP: "1.1.1.1"}, "host-1"},
		{Target{Host: "host-2.test", IP: "2.2.2.2"}, "host-2.test"},
		{Target{IP: "3.3.3.3"}, "3.3.3.3"},
		{Target{Host: "host-4.test", Hostname: "host-4", IP: "4.4.4.4", User: "user-4"}, "user-4@host-4"},
		{Target{Host: "host-5.test", IP: "5.5.5.5", User: "user-5"}, "user-5@host-5.test"},
		{Target{IP: "6.6.6.6", User: "user-6"}, "user-6@6.6.6.6"},
	}
	for _, item := range cases {
		actual := item.target.FriendlyName()
		if actual != item.expected {
			t.Errorf("Expected: %s. Actual: %s.", item.expected, actual)
		}
	}
	util.ExpectPanic(t, "At least one of Target.IP and Target.Host must be set", func() { _ = Target{User: "user-3"}.FriendlyName() })
}

func TestFromString(t *testing.T) {
	happyCases := []struct {
		input    string
		expected Target
	}{
		//host
		{"host-1", Target{Host: "host-1"}},
		{"user-2@host-2", Target{Host: "host-2", User: "user-2"}},
		{"@host-3", Target{Host: "host-3"}},
		// IPv4
		{"4.4.4.4", Target{IP: "4.4.4.4"}},
		{"user-5@5.5.5.5", Target{IP: "5.5.5.5", User: "user-5"}},
		{"@6.6.6.6", Target{IP: "6.6.6.6"}},
		// IPv6
		{"::7", Target{IP: "::7"}},
		{"user-5@::8", Target{IP: "::8", User: "user-5"}},
		{"@::9", Target{IP: "::9"}},
	}
	sadCases := []struct {
		input    string
		panicMsg string
	}{
		{"", "FromString(str string) Target got an empty string"},
		{"@", "At least one of Target.IP and Target.Host must be set"},
		{"user-1@", "At least one of Target.IP and Target.Host must be set"},
		{"a@b@c", "FromString(str string) Target got a string containing more than one @ character"},
	}
	for _, happy := range happyCases {
		actual := FromString(happy.input)
		if !reflect.DeepEqual(actual, happy.expected) {
			t.Errorf("Actual: %s. Expected: %s.", actual, happy.expected)
		}
	}
	for _, sad := range sadCases {
		util.ExpectPanic(t, sad.panicMsg, func() { FromString(sad.input) })
	}
}
