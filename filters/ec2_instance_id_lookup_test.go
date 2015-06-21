package filters

import (
	"encoding/json"
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

type hasVerify interface {
	Verify() (bool, error)
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

func givenAnEc2InstanceIdLookupWithMockedParserAndRunner() (*mockCommandRunner, *ec2InstanceIdLookup) {
	r := &mockCommandRunner{}
	f := &ec2InstanceIdLookup{idParser: dummyEc2InstanceIdParser{}, commandRunner: r, region: "dummy-region"}
	return r, f
}

func givenTargets(targetStrings ...string) []target.Target {
	targets := make([]target.Target, len(targetStrings))
	for i := 0; i < len(targetStrings); i++ {
		targets[i] = target.FromString(targetStrings[i])
	}
	return targets
}

func awsReturns(r *mockCommandRunner, instanceId string, region string, output string, err error) *mock.MockFunction {
	return r.When("RunGetOutput", "aws", []string{"ec2", "describe-instances", "--instance-id", instanceId, "--region", region}).Return([]byte(output), err)
}

func assertFilterResults(t *testing.T, f *ec2InstanceIdLookup, input []target.Target, expectedOutput []target.Target) {
	actualOutput := f.Filter(input)
	if len(input) != len(actualOutput) {
		t.Fail()
	}
	for i := 0; i < len(input); i++ {
		if expectedOutput[0] != actualOutput[0] {
			t.Errorf("Target %d was expected to be %s, found %s", i, expectedOutput[0], actualOutput[0])
		}
	}
}

func verifyMocks(t *testing.T, mocks ...hasVerify) {
	for _, m := range mocks {
		if ok, msg := m.Verify(); !ok {
			t.Error(msg)
		}
	}
}

func TestEc2InstanceIdLookupFails(t *testing.T) {
	instanceId := "dummy-instance-id"
	r, f := givenAnEc2InstanceIdLookupWithMockedParserAndRunner()
	// When the aws cli tool fails
	msg := "A client error (InvalidInstanceID.NotFound) occurred when calling the DescribeInstances operation: The instance ID 'i-deadbeef' does not exist"
	awsReturns(r, instanceId, f.region, msg, dummyError{"test fails aws"}).Times(2)
	// Filtering doesn't touch the target list
	targets := givenTargets(instanceId, instanceId)
	assertFilterResults(t, f, targets, targets)
	verifyMocks(t, r)
	// And no panic happened on JSON parsing, even though the CLI tools output was not valid JSON, because we don't even try to parse the output.
}

func TestEc2InstanceIdLookupInvalidJson(t *testing.T) {
	instanceId := "dummy-instance-id"
	r, f := givenAnEc2InstanceIdLookupWithMockedParserAndRunner()
	// When the AWS API returns invalid JSON
	invalidJson := "HAH! not a valid JSON"
	awsReturns(r, instanceId, f.region, invalidJson, nil).Times(1)
	// I get a fatal error for filtering
	expectPanic(t, fmt.Sprintf("Invalid JSON returned by AWS API.\nError: invalid character 'H' looking for beginning of value\nJSON follows this line\n%s", invalidJson),
		func() { f.Filter([]target.Target{target.FromString(instanceId)}) })
	verifyMocks(t, r)
	if ok, msg := r.Verify(); !ok {
		t.Error(msg)
	}
}

func jsonWithoutReservations() string {
	bytes, _ := json.Marshal(ec2DescribeInstanceApiResponse{Reservations: []ec2Reservation{}})
	return string(bytes)
}

func jsonWithIp(ip string) string {
	bytes, _ := json.Marshal(ec2DescribeInstanceApiResponse{Reservations: []ec2Reservation{ec2Reservation{Instances: []ec2Instance{ec2Instance{PublicIpAddress: ip}}}}})
	return string(bytes)
}

type lookupCase struct {
	inputHost string
	publicIp  string
	json      string
}

func makeLookupCase(inputHost string, publicIp string) lookupCase {
	c := lookupCase{inputHost: inputHost}
	if publicIp == "" {
		c.publicIp = inputHost
		c.json = jsonWithoutReservations()
	} else {
		c.publicIp = publicIp
		c.json = jsonWithIp(publicIp)
	}
	return c
}

func makeInputAndOutputTargets(cases []lookupCase) ([]target.Target, []target.Target) {
	inputTargets := make([]target.Target, len(cases))
	outputTargets := make([]target.Target, len(cases))
	for i, c := range cases {
		inputTargets[i] = target.FromString(c.inputHost)
		outputTargets[i] = target.FromString(c.publicIp)
	}
	return inputTargets, outputTargets
}

func TestEc2InstanceIdLookupHappyPath(t *testing.T) {
	r, f := givenAnEc2InstanceIdLookupWithMockedParserAndRunner()
	// When the AWS API returns IPs for some lookups, but not others
	cases := []lookupCase{
		makeLookupCase("no-hits", ""),
		makeLookupCase("foo.i-deadbeef.bar", "1.1.1.1"),
		makeLookupCase("i-12345678", "2.2.2.2"),
	}
	for _, c := range cases {
		awsReturns(r, c.inputHost, f.region, c.json, nil).Times(1)
	}
	// Filtering changes the targets which got results, but not the rest
	inputTargets, expectedOutputTargets := makeInputAndOutputTargets(cases)
	assertFilterResults(t, f, inputTargets, expectedOutputTargets)
	verifyMocks(t, r)
}
