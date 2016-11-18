package filters

import (
	"encoding/json"
	"fmt"
	"regexp"

	"strings"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

var ec2InstanceIdRegex = regexp.MustCompile("i-[0-9a-f]{8}")
var longEc2InstanceIdRegex = regexp.MustCompile("i-[0-9a-f]{17}")

type ec2InstanceIdParser interface {
	Parse(input string) string
}
type realEc2InstanceIdParser struct{}

func (p realEc2InstanceIdParser) Parse(input string) string {
	longID := longEc2InstanceIdRegex.FindString(input)
	if longID == "" {
		return ec2InstanceIdRegex.FindString(input)
	}
	return longID
}

type ec2Instance struct {
	InstanceId      string
	PublicDnsName   string
	PublicIpAddress string
}

type ec2Reservation struct {
	Instances []ec2Instance
}

type ec2DescribeInstanceApiResponse struct {
	Reservations []ec2Reservation
}

type ec2InstanceIdLookup struct {
	args          []interface{}
	region        string
	commandRunner util.CommandRunner
	idParser      ec2InstanceIdParser
}

func (f *ec2InstanceIdLookup) Filter(targets []target.Target) []target.Target {
	util.RequireArguments(f, 1, f.args)

	if len(targets) == 0 {
		util.Logger.Debugf("%s received no targets, skipping lookup", f)
		return targets
	}

	idToIndex := map[string]int{}
	ids := make([]string, 0, len(targets))
	for idx, t := range targets {
		instanceID := f.idParser.Parse(t.Host)
		if len(instanceID) > 0 {
			ids = append(ids, instanceID)
			idToIndex[instanceID] = idx
		} else {
			util.Logger.Debugf("Target %s looks like it doesn't have EC2 instance ID, skipping lookup for region %s", t.FriendlyName(), f.region)
		}
	}

	if len(ids) == 0 {
		util.Logger.Debugf("%s received no targets that look like they have EC2 instance IDs", f)
		return targets
	}

	util.Logger.Infof("EC2 Instance lookup: %s in %s", ids, f.region)

	outputs := f.commandRunner.Outputs("aws", append([]string{"ec2", "describe-instances", "--region", f.region, "--instance-ids"}, ids...))
	util.Logger.Debugf("Response from AWS API: %s", outputs.Combined)
	if outputs.Error != nil {
		util.Logger.Infof("EC2 Instance lookup failed in region %s (aws command failed): %s", f.region, strings.TrimSpace(string(outputs.Combined)))
		return targets
	}

	var data ec2DescribeInstanceApiResponse
	if err := json.Unmarshal(outputs.Combined, &data); err != nil {
		panic(fmt.Sprintf("Invalid JSON returned by AWS API.\nError: %s\nJSON follows this line\n%s", err.Error(), outputs.Combined))
	}

	if data.Reservations == nil || len(data.Reservations) == 0 {
		util.Logger.Infof("EC2 instance lookup failed in region %s (Reservations is empty in the received JSON)", f.region)
		return targets
	}

	for _, reservation := range data.Reservations {
		for _, instance := range reservation.Instances {
			id := instance.InstanceId
			idx := idToIndex[id]
			inputTargetName := targets[idx].Host
			targets[idx].IP = instance.PublicIpAddress
			targets[idx].Host = instance.PublicDnsName
			util.Logger.Infof("AWS API returned PublicIpAddress=%s PublicDnsName=%s for %s (%s)", targets[idx].IP, targets[idx].Host, inputTargetName, id)
		}
	}

	return targets
}
func (f *ec2InstanceIdLookup) SetArgs(args []interface{}) {
	util.RequireArguments(f, 1, args)
	f.args = args
	f.region = string(args[0].([]byte))
	util.RequireOnPath(f, "aws")
}
func (f *ec2InstanceIdLookup) String() string {
	return fmt.Sprintf("<%s %s>", nameEc2InstanceId, f.region)
}
