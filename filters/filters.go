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
	"strings"
	"io/ioutil"
	"os"
)

func Make(input string) interfaces.TargetFilter {
	return from_sexp.MakeFromString(input, nil, makeByName).(interfaces.TargetFilter)
}

func SupportedFilterNames() []string {
	var keys = make([]string, len(filterMakerMap))
	var i = 0
	for key := range filterMakerMap {
		keys[i] = key
		i += 1
	}
	return keys
}

func makeFromSExp(data []interface{}) interfaces.TargetFilter {
	return from_sexp.Make(data, nil, makeByName).(interfaces.TargetFilter)
}

const (
	nameEc2InstanceId = "ec2-instance-id"
	nameList          = "list"
	nameId            = "id"
	nameFirst         = "first"
	nameExternal      = "external"
)

var filterMakerMap = map[string]func() interfaces.TargetFilter{
	nameEc2InstanceId: func() interfaces.TargetFilter { return &ec2InstanceIdLookup{} },
	nameList:          func() interfaces.TargetFilter { return &list{} },
	nameId:            func() interfaces.TargetFilter { return &id{} },
	nameFirst:         func() interfaces.TargetFilter { return &first{} },
	nameExternal:      func() interfaces.TargetFilter { return &external{} },
}

func makeByName(name string) interface{} {
	var d interfaces.TargetFilter
	for key, maker := range filterMakerMap {
		if key == name {
			d = maker()
		}
	}
	if d == nil {
		util.Abort("filter \"%s\" is not known", name)
	}
	return d
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
	return fmt.Sprintf("<%s %s>", nameEc2InstanceId, f.region)
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
	return fmt.Sprintf("<%s %s>", nameList, f.children)
}

type id struct{}

func (f *id) Filter(targets []target.Target) []target.Target {
	return targets
}
func (f *id) SetArgs(args []interface{}) {
	util.RequireNoArguments(f, args)
}
func (f *id) String() string {
	return fmt.Sprintf("<%s>", nameId)
}

type first struct{}

func (f *first) Filter(targets []target.Target) []target.Target {
	if len(targets) > 0 {
		return targets[0:1]
	}
	return targets
}
func (f *first) SetArgs(args []interface{}) {
	util.RequireNoArguments(f, args)
}
func (f *first) String() string {
	return fmt.Sprintf("<%s>", nameFirst)
}

type external struct{
	Command exec.Cmd
}

func (f *external) Filter(targets []target.Target) []target.Target {
	var err error
	var output []byte
	var tmpFile *os.File
	tmpFile, err = ioutil.TempFile("", "easyssh")
	defer os.Remove(tmpFile.Name())
	if err != nil {
		util.Abort(err.Error())
	}
	tmpFile.Write([]byte(strings.Join(target.TargetStrings(targets), "\n")))
	f.Command.Args = append(f.Command.Args, tmpFile.Name())
	output, err = f.Command.Output()
	if err != nil {
		util.Abort("%s failed: %s", f, err)
	}
	var lines = strings.Split(strings.TrimSpace(string(output)), "\n")
	var newTargets = make([]target.Target, len(lines))
	var i int
	for i = 0; i < len(lines); i++ {
		newTargets[i] = target.FromString(lines[i])
	}
	return newTargets
}
func (f *external) SetArgs(args []interface{}) {
	util.RequireArguments(f, 1, args)
	var strArgs = make([]string, len(args))
	var i int
	for i = 0; i < len(args); i++ {
		strArgs[i] = string(args[i].([]uint8))
	}
	f.Command = *exec.Command(strArgs[0], strArgs[1:]...)
	f.Command.Stdin = os.Stdin
}
func (f *external) String() string {
	return fmt.Sprintf("<%s %s>", nameExternal, f.Command)
}