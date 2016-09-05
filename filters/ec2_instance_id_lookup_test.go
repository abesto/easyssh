package filters

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

type dummyEc2InstanceIdParser struct {
	shouldMatch bool
}

func (p dummyEc2InstanceIdParser) Parse(input string) string {
	if p.shouldMatch {
		return input + ".instanceid"
	}
	return ""
}

func TestEc2InstanceIdLookupStringViaMake(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		input := "(ec2-instance-id test-region)"
		structs := "[ec2-instance-id test-region]"
		final := "<ec2-instance-id test-region>"
		l.ExpectDebugf("MakeFromString %s -> %s", input, structs)
		l.ExpectDebugf("Make %s -> %s", structs, final)
		Make(input)
	})
}

func TestEc2InstanceIdLookupMakeWithoutArgument(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(ec2-instance-id)", "[ec2-instance-id]")
		util.ExpectPanic(t, "<ec2-instance-id > requires exactly 1 argument(s), got 0: []",
			func() { Make("(ec2-instance-id)") })
	})
}

func TestEc2InstanceIdLookupFilterWithoutSetArgs(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		util.ExpectPanic(t, "<ec2-instance-id > requires exactly 1 argument(s), got 0: []",
			func() { (&ec2InstanceIdLookup{}).Filter([]target.Target{}) })
	})
}

func TestEc2InstanceIdSetTooManyArgs(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(ec2-instance-id foo bar)", "[ec2-instance-id foo bar]")
		util.ExpectPanic(t, "<ec2-instance-id > requires exactly 1 argument(s), got 2: [foo bar]",
			func() { Make("(ec2-instance-id foo bar)") })
	})
}

func TestEc2InstanceIdSetArgs(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(ec2-instance-id foo)", "[ec2-instance-id foo]").Times(1)
		l.ExpectDebugf("Make %s -> %s", "[ec2-instance-id foo]", "<ec2-instance-id foo>").Times(1)
		f := Make("(ec2-instance-id foo)").(*ec2InstanceIdLookup)
		if f.region != "foo" {
			t.Errorf("Expected region to be foo, was %s", f.region)
		}
		if len(f.args) != 1 || fmt.Sprintf("%s", f.args[0]) != "foo" {
			t.Error(len(f.args), f.args)
		}
	})
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

func givenAnEc2InstanceIdLookupWithMockedParserAndRunner(shouldMatch bool) (*util.MockCommandRunner, *ec2InstanceIdLookup) {
	r := &util.MockCommandRunner{}
	idParser := dummyEc2InstanceIdParser{shouldMatch}
	f := &ec2InstanceIdLookup{idParser: idParser, commandRunner: r, region: "dummy-region", args: []interface{}{"dummy-region"}}
	return r, f
}

func awsReturns(r *util.MockCommandRunner, instanceIds []string, region string, output string, err error) *mock.Call {
	return r.On("Outputs", "aws", append([]string{"ec2", "describe-instances", "--region", region, "--instance-ids"}, instanceIds...)).Return(
		util.CommandRunnerOutputs{Combined: []byte(output), Error: err})
}

func assertFilterResults(t *testing.T, f *ec2InstanceIdLookup, input []target.Target, expectedOutput []target.Target) {
	actualOutput := f.Filter(input)
	if len(input) != len(actualOutput) {
		t.Fail()
	}
	for i := 0; i < len(input); i++ {
		if expectedOutput[0].IP != actualOutput[0].IP {
			t.Errorf("IP of Target %d was expected to be %s, found %s", i, expectedOutput[0].IP, actualOutput[0].IP)
		}
	}
}

func TestEc2InstanceIdLookupFails(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		r, f := givenAnEc2InstanceIdLookupWithMockedParserAndRunner(true)
		msg := "A client error (InvalidInstanceID.NotFound) occurred when calling the DescribeInstances operation: The instance ID 'i-deadbeef' does not exist"
		host := "dummy-instance-id"
		instanceId := host + ".instanceid"
		targets := target.FromStrings(host, host)
		l.ExpectInfof("EC2 Instance lookup: %s in %s", "[dummy-instance-id.instanceid dummy-instance-id.instanceid]", f.region)
		l.ExpectDebugf("Response from AWS API: %s", msg)
		l.ExpectInfof("EC2 Instance lookup failed in region %s (aws command failed): %s", f.region, msg)
		// On the aws cli tool fails
		awsReturns(r, []string{instanceId, instanceId}, f.region, msg, util.DummyError{Msg: "test fails aws"})
		// Filtering doesn't touch the target list
		assertFilterResults(t, f, targets, targets)
		r.AssertExpectations(t)
		// And no panic happened on JSON parsing, even though the CLI tools output was not valid JSON, because we don't even try to parse the output.
	})
}

func TestEc2InstanceIdLookupInvalidJson(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		invalidJson := "HAH! not a valid JSON"
		host := "dummy-instance-id"
		instanceId := host + ".instanceid"
		r, f := givenAnEc2InstanceIdLookupWithMockedParserAndRunner(true)
		l.ExpectDebugf("Response from AWS API: %s", invalidJson)
		l.ExpectInfof("EC2 Instance lookup: %s in %s", "[dummy-instance-id.instanceid]", f.region)
		// On the AWS API returns invalid JSON
		awsReturns(r, []string{instanceId}, f.region, invalidJson, nil).Times(1)
		// I get a fatal error for filtering
		util.ExpectPanic(t, fmt.Sprintf("Invalid JSON returned by AWS API.\nError: invalid character 'H' looking for beginning of value\nJSON follows this line\n%s", invalidJson),
			func() { f.Filter([]target.Target{target.FromString(host)}) })
		r.AssertExpectations(t)
	})
}

func jsonWithoutReservations() string {
	bytes, _ := json.Marshal(ec2DescribeInstanceApiResponse{Reservations: []ec2Reservation{}})
	return string(bytes)
}

func jsonWithIp(ip string, dnsName string, instanceId string) string {
	bytes, _ := json.Marshal(ec2DescribeInstanceApiResponse{Reservations: []ec2Reservation{{Instances: []ec2Instance{{PublicIpAddress: ip, PublicDnsName: dnsName, InstanceId: instanceId}}}}})
	return string(bytes)
}

type lookupCase struct {
	inputHost  string
	instanceId string
	publicIp   string
	publicDns  string
	json       string
}

func makeLookupCase(inputHost string, publicIp string, publicDns string) lookupCase {
	instanceId := inputHost + ".instanceid"
	c := lookupCase{inputHost: inputHost, instanceId: instanceId, publicDns: publicDns}
	if publicIp == "" {
		c.publicIp = inputHost
		c.json = jsonWithoutReservations()
	} else {
		c.publicIp = publicIp
		c.json = jsonWithIp(publicIp, publicDns, instanceId)
	}
	return c
}

func makeInputAndOutputTargets(cases []lookupCase, shouldRewrite bool) ([]target.Target, []target.Target) {
	inputTargets := make([]target.Target, len(cases))
	outputTargets := make([]target.Target, len(cases))
	for i, c := range cases {
		inputTargets[i] = target.FromString(c.inputHost)
		if shouldRewrite {
			outputTargets[i] = target.FromString(c.publicIp)
		} else {
			outputTargets[i] = target.FromString(c.inputHost)
		}
	}
	return inputTargets, outputTargets
}

func mergeJsonsOfCases(cases []lookupCase) string {
	mergedData := ec2DescribeInstanceApiResponse{}
	for _, c := range cases {
		var caseData ec2DescribeInstanceApiResponse
		json.Unmarshal([]byte(c.json), &caseData)
		mergedData.Reservations = append(mergedData.Reservations, caseData.Reservations...)
	}
	mergedJson, _ := json.Marshal(mergedData)
	return string(mergedJson)
}

func assertLookupCasesPass(t *testing.T, r *util.MockCommandRunner, f *ec2InstanceIdLookup, shouldRewrite bool, cases []lookupCase) {
	ids := make([]string, len(cases))
	for i, c := range cases {
		ids[i] = c.instanceId
	}
	awsReturns(r, ids, f.region, mergeJsonsOfCases(cases), nil).Times(1)
	inputTargets, expectedOutputTargets := makeInputAndOutputTargets(cases, shouldRewrite)
	assertFilterResults(t, f, inputTargets, expectedOutputTargets)
}

func TestEc2InstanceIdLookupDoesntLookLikeInstanceId(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		cases := []lookupCase{
			makeLookupCase("no-hits", "", ""),
			makeLookupCase("foo.i-deadbeef.bar", "1.1.1.1", "public-deadbeef"),
			makeLookupCase("i-12345678", "2.2.2.2", "public-12345678"),
		}
		r, f := givenAnEc2InstanceIdLookupWithMockedParserAndRunner(false)
		for _, c := range cases {
			l.ExpectDebugf("Target %s looks like it doesn't have EC2 instance ID, skipping lookup for region %s", c.inputHost, f.region)
		}
		l.ExpectDebugf("%s received no targets that look like they have EC2 instance IDs", f.String())
		assertLookupCasesPass(t, r, f, false, cases)
	})
}

func TestEc2InstanceIdLookupHappyPath(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		r, f := givenAnEc2InstanceIdLookupWithMockedParserAndRunner(true)
		cases := []lookupCase{
			makeLookupCase("foo.i-deadbeef.bar", "1.1.1.1", "public-deadbeef"),
			makeLookupCase("i-12345678", "2.2.2.2", "public-12345678"),
		}

		l.ExpectInfof("EC2 Instance lookup: %s in %s", "[foo.i-deadbeef.bar.instanceid i-12345678.instanceid]", f.region)
		l.ExpectDebugf("Response from AWS API: %s", mergeJsonsOfCases(cases))
		for _, c := range cases {
			l.ExpectInfof("AWS API returned PublicIpAddress=%s PublicDnsName=%s for %s (%s)", c.publicIp, c.publicDns, c.inputHost, c.instanceId)
		}

		assertLookupCasesPass(t, r, f, true, cases)
		r.AssertExpectations(t)
	})
}
