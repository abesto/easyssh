package discoverers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/abesto/easyssh/from_sexp"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/util"
	"os/exec"
	"strings"
)

func Make(input string) interfaces.Discoverer {
	return from_sexp.MakeFromString(input, makeByName).(interfaces.Discoverer)
}

func makeFromSExp(data []interface{}) interfaces.Discoverer {
	return from_sexp.Make(data, makeByName).(interfaces.Discoverer)
}

func makeByName(name string) interface{} {
	var d interfaces.Discoverer
	switch name {
	case "comma-separated":
		d = &commaSeparated{}
	case "knife":
		d = &knifeSearch{}
	case "first-matching":
		d = &firstMatching{}
	default:
		util.Abort("Command \"%s\" is not known", name)
	}
	return d
}

type commaSeparated struct{}

func (d *commaSeparated) Discover(input string) []string {
	return strings.Split(input, ",")
}
func (d *commaSeparated) SetArgs(args []interface{}) {
	if len(args) > 0 {
		util.Abort("%s takes no configuration, got %s", d, args)
	}
}
func (d *commaSeparated) String() string {
	return "[comma-separated]"
}

type knifeSearch struct{}

func (d *knifeSearch) Discover(input string) []string {
	if !strings.Contains(input, ":") {
		fmt.Printf("Host lookup string doesn't contain ':', it won't match anything in a knife search node query\n")
		return []string{}
	}

	fmt.Printf("Looking up nodes with knife matching %s\n", input)

	var cmd = exec.Command("knife", "search", "node", "-F", "json", input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	var error = cmd.Run()
	if error != nil {
		fmt.Print(stderr.String())
		util.Abort(error.Error())
	}

	var data map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &data)

	var ips = []string{}
	for _, row := range data["rows"].([]interface{}) {
		var automatic = row.(map[string]interface{})["automatic"].(map[string]interface{})
		if cloud_v2, ok := automatic["cloud_v2"]; ok && cloud_v2 != nil {
			ips = append(ips, cloud_v2.(map[string]interface{})["public_ipv4"].(string))
		} else {
			ips = append(ips, automatic["ipaddress"].(string))
		}
	}

	return ips
}
func (d *knifeSearch) SetArgs(args []interface{}) {
	if len(args) > 0 {
		util.Abort("%s takes no configuration, got %s", d, args)
	}
}
func (d *knifeSearch) String() string {
	return "[knife]"
}

type firstMatching struct {
	discoverers []interfaces.Discoverer
}

func (d *firstMatching) Discover(input string) []string {
	var hosts []string
	for _, discoverer := range d.discoverers {
		fmt.Printf("Trying discoverer %s\n", discoverer)
		hosts = discoverer.Discover(input)
		if len(hosts) > 0 {
			return hosts
		}
	}
	return []string{}
}
func (d *firstMatching) SetArgs(args []interface{}) {
	d.discoverers = []interfaces.Discoverer{}
	for _, exp := range args {
		d.discoverers = append(d.discoverers, makeFromSExp(exp.([]interface{})))
	}
	fmt.Printf("Will use the first discoverer returning a non-empty host set from the discoverer list %s\n", d.discoverers)
}
func (d *firstMatching) String() string {
	var strs = []string{}
	for _, child := range d.discoverers {
		strs = append(strs, child.String())
	}
	return fmt.Sprintf("[first-matching %s]", strings.Join(strs, " "))
}
