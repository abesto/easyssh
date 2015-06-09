package filters

import (
	"encoding/json"
	"fmt"
	"github.com/abesto/easyssh/from_sexp"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
	"os/exec"
	"regexp"
)

func Make(input string) interfaces.TargetFilter {
	return from_sexp.MakeFromString(input, makeByName).(interfaces.TargetFilter)
}

func makeFromSExp(data []interface{}) interfaces.TargetFilter {
	return from_sexp.Make(data, makeByName).(interfaces.TargetFilter)
}

func makeByName(name string) interface{} {
	var f interfaces.TargetFilter
	switch name {
	case "ec2-instance-id":
		f = &ec2InstanceIdLookup{}
	case "list":
		f = &list{}
	case "id":
		f = &id{}
	case "first":
		f = &first{}
	default:
		util.Abort("Filter \"%s\" is not known", name)
	}
	return f
}

type ec2InstanceIdLookup struct {
	region string
}

func (f *ec2InstanceIdLookup) Filter(targets []target.Target) []target.Target {
	if f.region == "" {
		util.Abort("ec2-instance-id requires exactly one argument, the region name to use for looking up instances")
	}
	var re = regexp.MustCompile("i-[0-9a-f]{8}")
	for idx, t := range targets {
		var instanceId = re.FindString(t.Host)
		if len(instanceId) > 0 {
			var cmd = exec.Command("aws", "ec2", "describe-instances", "--instance-id", instanceId, "--region", f.region)
			util.Logger.Infof("EC2 Instance lookup: %s", cmd.Args)
			var output, _ = cmd.Output()
			var data map[string]interface{}
			json.Unmarshal(output, &data)

			var reservations = data["Reservations"]
			if reservations == nil {
				util.Logger.Infof("EC2 instance lookup failed for %s (%s) in region %s", t.Host, instanceId, f.region)
				continue
			}
			targets[idx].Host = reservations.([]interface{})[0].(map[string]interface{})["Instances"].([]interface{})[0].(map[string]interface{})["PublicIpAddress"].(string)
		} else {
			util.Logger.Debugf("Target %s looks like it doesn't have EC2 instance ID, skipping lookup for region %s", t, f.region)
		}
	}
	return targets
}
func (f *ec2InstanceIdLookup) SetArgs(args []interface{}) {
	if len(args) != 1 {
		util.Abort("ec2-instance-id requires exactly one argument, the region name to use for looking up instances")
	}
	f.region = string(args[0].([]byte))
}
func (f *ec2InstanceIdLookup) String() string {
	return fmt.Sprintf("<ec2-instance-id %s>", f.region)
}

type list struct {
	children []interfaces.TargetFilter
}

func (f *list) Filter(targets []target.Target) []target.Target {
	for _, child := range f.children {
		targets = child.Filter(targets)
		util.Logger.Debugf("Targets after filter %s: %s", child, targets)
	}
	return targets
}
func (f *list) SetArgs(args []interface{}) {
	for _, def := range args {
		f.children = append(f.children, makeFromSExp(def.([]interface{})))
	}
}
func (f *list) String() string {
	return fmt.Sprintf("<list %s>", f.children)
}

type id struct{}

func (f *id) Filter(targets []target.Target) []target.Target {
	return targets
}
func (f *id) SetArgs(args []interface{}) {
	util.RequireNoArguments(f, args)
}
func (f *id) String() string {
	return "<id>"
}

type first struct{}

func (f *first) Filter(targets []target.Target) []target.Target {
	if len(targets) > 0 {
		return targets[1:]
	}
	return targets
}
func (f *first) SetArgs(args []interface{}) {
	util.RequireNoArguments(f, args)
}
func (f *first) String() string {
	return "<first>"
}