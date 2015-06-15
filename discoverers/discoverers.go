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
	"github.com/eadmund/sexprs"
)

func Make(input string) interfaces.Discoverer {
	return from_sexp.MakeFromString(input, makeByName).(interfaces.Discoverer)
}

func SupportedDiscovererNames() []string {
	var keys = make([]string, len(discovererMakerMap))
	var i = 0
	for key := range discovererMakerMap {
		keys[i] = key
		i += 1
	}
	return keys
}

func makeFromSExp(data sexprs.Sexp) interfaces.Discoverer {
	return from_sexp.Make(data, makeByName).(interfaces.Discoverer)
}

const (
	nameCommaSeparated = "comma-separated"
	nameKnife          = "knife"
	nameFirstMatching  = "first-matching"
)

var discovererMakerMap = map[string]func() interfaces.Discoverer{
	nameCommaSeparated: func() interfaces.Discoverer { return &commaSeparated{} },
	nameKnife:          func() interfaces.Discoverer { return &knifeSearch{} },
	nameFirstMatching:  func() interfaces.Discoverer { return &firstMatching{} },
}

func makeByName(name string) interface{} {
	var d interfaces.Discoverer
	for key, maker := range discovererMakerMap {
		if key == name {
			d = maker()
		}
	}
	if d == nil {
		util.Abort("Discoverer \"%s\" is not known", name)
	}
	return d
}

type commaSeparated struct{}

func (d *commaSeparated) Discover(input string) []string {
	return strings.Split(input, ",")
}
func (d *commaSeparated) SetArgs(args []sexprs.Sexp) {
	if len(args) > 0 {
		util.Abort("%s takes no configuration, got %s", d, args)
	}
}
func (d *commaSeparated) String() string {
	return fmt.Sprintf("<%s>", nameCommaSeparated)
}

type knifeSearch struct{}

func (d *knifeSearch) Discover(input string) []string {
	if !strings.Contains(input, ":") {
		util.Logger.Debugf("Host lookup string doesn't contain ':', it won't match anything in a knife search node query")
		return []string{}
	}

	util.Logger.Infof("Looking up nodes with knife matching %s", input)

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
func (d *knifeSearch) SetArgs(args []sexprs.Sexp) {
	if len(args) > 0 {
		util.Abort("%s takes no configuration, got %s", d, args)
	}
}
func (d *knifeSearch) String() string {
	return fmt.Sprintf("<%s>", nameKnife)
}

type firstMatching struct {
	discoverers []interfaces.Discoverer
}

func (d *firstMatching) Discover(input string) []string {
	var hosts []string
	for _, discoverer := range d.discoverers {
		util.Logger.Debugf("Trying discoverer %s", discoverer)
		hosts = discoverer.Discover(input)
		if len(hosts) > 0 {
			return hosts
		}
	}
	return []string{}
}
func (d *firstMatching) SetArgs(args []sexprs.Sexp) {
	d.discoverers = []interfaces.Discoverer{}
	for _, exp := range args {
		d.discoverers = append(d.discoverers, makeFromSExp(exp))
	}
}
func (d *firstMatching) String() string {
	var strs = []string{}
	for _, child := range d.discoverers {
		strs = append(strs, child.String())
	}
	return fmt.Sprintf("<%s %s>", nameFirstMatching, strings.Join(strs, " "))
}
