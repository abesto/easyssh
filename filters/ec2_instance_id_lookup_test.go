package filters

import (
	"fmt"
	"github.com/abesto/easyssh/target"
	"github.com/maraino/go-mock"
	"testing"
)

func expectPanic(t *testing.T, expectedErr interface{}, f func()) {
	defer func() {
		actualErr := recover()
		if actualErr == nil {
			t.Errorf("Expected panic(\"%s\"), got no panic", expectedErr)
		}
		if actualErr != expectedErr {
			t.Errorf("Expected panic(\"%s\"), got panic(\"%s\") instead", expectedErr, actualErr)
		}
	}()
	f()
}

type dummyEc2InstanceIdParser struct{}

func (p dummyEc2InstanceIdParser) Parse(input string) string {
	return input
}

type mockCommandRunner struct {
	mock.Mock
}

func (r *mockCommandRunner) RunWithStdinGetOutputOrPanic(name string, args []string) []byte {
	ret := r.Called(name, args)
	return ret.Bytes(0)
}
func (r *mockCommandRunner) RunGetOutputOrPanic(name string, args []string) []byte {
	ret := r.Called(name, args)
	return ret.Bytes(0)
}
func (r *mockCommandRunner) RunGetOutput(name string, args []string) ([]byte, error) {
	ret := r.Called(name, args)
	return ret.Bytes(0), ret.Error(1)
}

type dummyError struct {
	msg string
}

func (e dummyError) Error() string {
	return e.msg
}

func TestEc2InstanceIdLookupString(t *testing.T) {
	f := Make("(ec2-instance-id test-region)")
	expected := "<ec2-instance-id test-region>"
	actual := f.String()
	if actual != expected {
		t.Errorf("String() output was expected to be %s, was %s", expected, actual)
	}
}

func TestEc2InstanceIdLookupMakeWithoutArgument(t *testing.T) {
	expectPanic(t, "ec2-instance-id requires exactly one argument, the region name to use for looking up instances",
		func() { Make("(ec2-instance-id)") })
}

func TestEc2InstanceIdLookupFilterWithoutSetArgs(t *testing.T) {
	expectPanic(t, "ec2-instance-id requires exactly one argument, the region name to use for looking up instances",
		func() { (&ec2InstanceIdLookup{}).Filter([]target.Target{}) })
}

func TestEc2InstanceIdLookupSetTooManyArgs(t *testing.T) {
	expectPanic(t, "ec2-instance-id requires exactly one argument, the region name to use for looking up instances",
		func() { Make("(ec2-instance-id foo bar)") })
}

func TestEc2InstanceIdLookupSetArgs(t *testing.T) {
	f := Make("(ec2-instance-id foo)").(*ec2InstanceIdLookup)
	if f.region != "foo" {
		t.Errorf("Expected region to be foo, was %s", f.region)
	}
}

func TestEc2InstanceIdParser(t *testing.T) {
	cases := map[string]string{
		"foo-i-deadbeef.subnet.private": "i-deadbeef",
		"i-foo":      "",
		"abesto.net": "",
	}
	parser := realEc2InstanceIdParser{}
	for input, expected := range cases {
		actual := parser.Parse(input)
		if actual != expected {
			t.Errorf("Parsed '%s' into EC2 instance id '%s', expected '%s'", input, actual, expected)
		}
	}
}

func TestEc2InstanceIdLookupFails(t *testing.T) {
	// Given an ec2InstanceIdLookup with mocked parser and runner
	instanceId := "dummy-instance-id"
	r := &mockCommandRunner{}
	f := ec2InstanceIdLookup{idParser: dummyEc2InstanceIdParser{}, commandRunner: r, region: "dummy-region"}
	// When the aws cli tool fails
	msg := "A client error (InvalidInstanceID.NotFound) occurred when calling the DescribeInstances operation: The instance ID 'i-deadbeef' does not exist"
	r.When("RunGetOutput", "aws", []string{"ec2", "describe-instances", "--instance-id", instanceId, "--region", f.region}).Return([]byte(msg), dummyError{"test fails aws"}).Times(2)
	// Filtering doesn't touch the target list
	targets := []target.Target{target.FromString(instanceId), target.FromString(instanceId)}
	actualTargets := f.Filter(targets)
	if len(targets) != len(actualTargets) {
		t.Fail()
	}
	if targets[0] != actualTargets[0] || targets[1] != actualTargets[1] {
		t.Fail()
	}
	if ok, msg := r.Verify(); !ok {
		t.Error(msg)
	}
	// And no panic happened on JSON parsing, even though the CLI tools output was not valid JSON, because we don't even try to parse the output.
}

func TestEc2InstanceIdLookupInvalidJson(t *testing.T) {
	// Given an ec2InstanceIdLookup with mocked parser and runner
	instanceId := "dummy-instance-id"
	r := &mockCommandRunner{}
	f := ec2InstanceIdLookup{idParser: dummyEc2InstanceIdParser{}, commandRunner: r, region: "dummy-region"}
	// When the AWS API returns invalid JSON
	invalidJson := "HAH! not a valid JSON"
	r.When("RunGetOutput", "aws", []string{"ec2", "describe-instances", "--instance-id", instanceId, "--region", f.region}).Return([]byte(invalidJson)).Times(1)
	// I get a fatal error for filtering
	expectPanic(t, fmt.Sprintf("Invalid JSON returned by AWS API.\nError: invalid character 'H' looking for beginning of value\nJSON follows this line\n%s", invalidJson),
		func() { f.Filter([]target.Target{target.FromString(instanceId)}) })
	if ok, msg := r.Verify(); !ok {
		t.Error(msg)
	}
}
