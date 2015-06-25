package discoverers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/abesto/easyssh/fromsexp"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/util"
)

func Make(input string) interfaces.Discoverer {
	return fromsexp.MakeFromString(input, nil, makeByName).(interfaces.Discoverer)
}

func SupportedDiscovererNames() []string {
	var keys = make([]string, len(discovererMakerMap))
	var i = 0
	for key := range discovererMakerMap {
		keys[i] = key
		i++
	}
	return keys
}

func makeFromSExp(data []interface{}) interfaces.Discoverer {
	return fromsexp.Make(data, nil, makeByName).(interfaces.Discoverer)
}

const (
	nameCommaSeparated = "comma-separated"
	nameKnife          = "knife"
	nameKnifeHostname  = "knife-hostname"
	nameFirstMatching  = "first-matching"
)

var discovererMakerMap = map[string]func() interfaces.Discoverer{
	nameCommaSeparated: func() interfaces.Discoverer { return &commaSeparated{} },
	nameKnife:          func() interfaces.Discoverer { return &knifeSearch{publicIp} },
	nameKnifeHostname:  func() interfaces.Discoverer { return &knifeSearch{publicHostname} },
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
		util.Panicf("Discoverer \"%s\" is not known", name)
	}
	return d
}

type knifeSearchResultType int

const (
	publicIp knifeSearchResultType = iota
	publicHostname
)

type knifeSearch struct {
	resultType knifeSearchResultType
}

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
	err := cmd.Run()
	if err != nil {
		fmt.Print(stderr.String())
		util.Panicf(err.Error())
	}

	var data map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &data)

	var ips = []string{}
	fieldName := "public_ipv4"
	if d.resultType == publicHostname {
		fieldName = "public_hostname"
	}
	for _, row := range data["rows"].([]interface{}) {
		var automatic = row.(map[string]interface{})["automatic"].(map[string]interface{})
		if cloudV2, ok := automatic["cloud_v2"]; ok && cloudV2 != nil {
			ips = append(ips, cloudV2.(map[string]interface{})[fieldName].(string))
		} else {
			ips = append(ips, automatic["ipaddress"].(string))
		}
	}

	return ips
}
func (d *knifeSearch) SetArgs(args []interface{}) {
	if len(args) > 0 {
		util.Panicf("%s takes no configuration, got %s", d, args)
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
func (d *firstMatching) SetArgs(args []interface{}) {
	d.discoverers = []interfaces.Discoverer{}
	for _, exp := range args {
		d.discoverers = append(d.discoverers, makeFromSExp(exp.([]interface{})))
	}
}
func (d *firstMatching) String() string {
	var strs = []string{}
	for _, child := range d.discoverers {
		strs = append(strs, child.String())
	}
	return fmt.Sprintf("<%s %s>", nameFirstMatching, strings.Join(strs, " "))
}
