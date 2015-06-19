package filters

import (
	"testing"

	"github.com/abesto/easyssh/target"
)

func TestEc2InstanceIdLookupString(t *testing.T) {
	f := Make("(ec2-instance-id test-region)")
	expected := "<ec2-instance-id test-region>"
	actual := f.String()
	if actual != expected {
		t.Errorf("String() output was expected to be %s, was %s", expected, actual)
	}
}

func expectPanic(t *testing.T, expectedErr interface{}, f func()) {
	defer func() {
		actualErr := recover()
		if actualErr == nil {
			t.Errorf("Expected panic(%s), got no panic", expectedErr)
		}
		if actualErr != expectedErr {
			t.Errorf("Expected panic(%s), got panic(%s) instead", expectedErr, actualErr)
		}
	}()
	f()
}

func TestEc2InstanceIdLookupMakeWithoutArgument(t *testing.T) {
	expectPanic(t, "ec2-instance-id requires exactly one argument, the region name to use for looking up instances",
		func() { Make("(ec2-instance-id)") })
}

func TestEc2InstanceIdLookupFilterWithoutSetArgs(t *testing.T) {
	expectPanic(t, "ec2-instance-id requires exactly one argument, the region name to use for looking up instances",
		func() { (&ec2InstanceIdLookup{}).Filter([]target.Target{}) })
}

func TestEc2InstanceIdLookupFilterSetTooManyArgs(t *testing.T) {
	expectPanic(t, "ec2-instance-id requires exactly one argument, the region name to use for looking up instances",
		func() { Make("(ec2-instance-id foo bar)") })
}

func TestEc2InstanceIdLookupFilterSetArgs(t *testing.T) {
	f := Make("(ec2-instance-id foo)").(*ec2InstanceIdLookup)
	if f.region != "foo" {
		t.Errorf("Expected region to be foo, was %s", f.region)
	}
}
