package filters

import (
	"bitbucket.org/shanehanna/sexp"
	"encoding/json"
	"fmt"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
	"os/exec"
	"regexp"
)

type Ec2InstanceIdFilter struct {
	region string
}

func (d *Ec2InstanceIdFilter) Filter(targets []target.Target) []target.Target {
	if d.region == "" {
		util.Abort("ec2-instance-id requires exactly one argument, the region name to use for looking up instances")
	}
	var re = regexp.MustCompile("i-[0-9a-f]{8}")
	for idx, t := range targets {
		var instanceId = re.FindString(t.Host)
		if len(instanceId) > 0 {
			var cmd = exec.Command("aws", "ec2", "describe-instances", "--instance-id", instanceId, "--region", d.region)
			fmt.Printf("EC2 Instance lookup: %s\n", cmd.Args)
			var output, _ = cmd.Output()
			var data map[string]interface{}
			json.Unmarshal(output, &data)

			var reservations = data["Reservations"]
			if reservations == nil {
				fmt.Printf("EC2 instance lookup failed for %s (%s) in region %s\n", t.Host, instanceId, d.region)
				continue
			}
			targets[idx].Host = reservations.([]interface{})[0].(map[string]interface{})["Instances"].([]interface{})[0].(map[string]interface{})["PublicIpAddress"].(string)
		}
	}
	return targets
}
func (d *Ec2InstanceIdFilter) SetArgs(args []interface{}) {
	if len(args) != 1 {
		util.Abort("ec2-instance-id requires exactly one argument, the region name to use for looking up instances")
	}
	d.region = string(args[0].([]byte))
}
func (d *Ec2InstanceIdFilter) String() string {
	return fmt.Sprintf("[ec2-instance-id %s]", d.region)
}

var filterMap = map[string]interfaces.TargetFilter{
	"ec2-instance-id": &Ec2InstanceIdFilter{},
}

func ApplyFilters(filterDefinition string, targets []target.Target) []target.Target {
	var filterExprs, error = sexp.Unmarshal([]byte(filterDefinition))
	if error != nil {
		util.Abort(error.Error())
	}

	for _, filterExpr := range filterExprs {
		var parts = filterExpr.([]interface{})
		var filter = MakeFilterByName(string(parts[0].([]byte)))
		if len(parts) > 1 {
			filter.SetArgs(parts[1:])
		}

		targets = filter.Filter(targets)
		fmt.Printf("Targets after filter %s: %s\n", filter, targets)
	}

	return targets
}

func MakeFilterByName(name string) interfaces.TargetFilter {
	var filter, ok = filterMap[name]
	if !ok {
		util.Abort("Unknown filter %s", name)
	}
	return filter
}
