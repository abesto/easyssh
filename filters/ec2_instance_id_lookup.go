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

type ec2InstanceIdParser interface {
	Parse(input string) string
}
type realEc2InstanceIdParser struct{}

func (p realEc2InstanceIdParser) Parse(input string) string {
	return ec2InstanceIdRegex.FindString(input)
}

type ec2Instance struct {
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
	// TODO: build map by parsed id, do a single query using --instance-ids
	for idx, t := range targets {
		instanceId := f.idParser.Parse(t.Host)
		if len(instanceId) > 0 {
			util.Logger.Infof("EC2 Instance lookup: %s in %s", instanceId, f.region)

			outputs := f.commandRunner.Outputs("aws", []string{"ec2", "describe-instances", "--instance-id", instanceId, "--region", f.region})
			util.Logger.Debugf("Response from AWS API: %s", outputs.Combined)
			if outputs.Error != nil {
				util.Logger.Infof("EC2 Instance lookup failed for %s (%s) in region %s (aws command failed): %s", t.Host, instanceId, f.region, strings.TrimSpace(string(outputs.Combined)))
				continue
			}

			var data ec2DescribeInstanceApiResponse
			if err := json.Unmarshal(outputs.Combined, &data); err != nil {
				panic(fmt.Sprintf("Invalid JSON returned by AWS API.\nError: %s\nJSON follows this line\n%s", err.Error(), outputs.Combined))
			}

			if data.Reservations == nil || len(data.Reservations) == 0 {
				util.Logger.Infof("EC2 instance lookup failed for %s (%s) in region %s (Reservations is empty in the received JSON)", t.Host, instanceId, f.region)
				continue
			}

			ip := data.Reservations[0].Instances[0].PublicIpAddress
			util.Logger.Infof("AWS API returned PublicIpAddress %s for %s (%s)", ip, targets[idx].Host, instanceId)
			targets[idx].Host = ip
		} else {
			util.Logger.Debugf("Target %s looks like it doesn't have EC2 instance ID, skipping lookup for region %s", t, f.region)
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
