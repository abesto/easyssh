package discoverers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/abesto/easyssh/util"
)

type knifeSearchResultType int

const (
	publicIp knifeSearchResultType = iota
	publicHostname
)

type knifeSearch struct {
	resultType    knifeSearchResultType
	commandRunner util.CommandRunner
}

type searchResult struct {
	Rows []struct {
		Name      string
		Automatic *struct {
			CloudV2 *struct {
				PublicHostname string `json:"public_hostname"`
				PublicIpv4     string `json:"public_ipv4"`
			} `json:"cloud_v2"`
			Ipaddress string
		}
	}
}

func (d *knifeSearch) Discover(input string) []string {
	if !strings.Contains(input, ":") {
		util.Logger.Debugf("Host lookup string doesn't contain ':', it won't match anything in a knife search node query")
		return []string{}
	}

	util.Logger.Infof("Looking up nodes with knife matching %s", input)
	outputs := d.commandRunner.Outputs("knife", []string{"search", "node", "-F", "json", input})
	if outputs.Error != nil {
		util.Panicf("Knife lookup failed: %s\nOutput:\n%s", outputs.Error, outputs.Combined)
	}

	data := searchResult{}
	if err := json.Unmarshal(outputs.Stdout, &data); err != nil {
		util.Panicf("Failed to parse knife search result: %s", err)
	}
	//	util.Logger.Debugf("Parsed result into data: %s", data)

	var ips = []string{}
	for _, row := range data.Rows {
		var thisIp string = ""
		if row.Automatic.CloudV2 != nil {
			if d.resultType == publicHostname {
				thisIp = row.Automatic.CloudV2.PublicHostname
			} else {
				thisIp = row.Automatic.CloudV2.PublicIpv4
			}
		} else {
			thisIp = row.Automatic.Ipaddress
		}
		ips = append(ips, thisIp)
	}

	return ips
}
func (d *knifeSearch) SetArgs(args []interface{}) {
	if len(args) > 0 {
		util.Panicf("%s takes no configuration, got %s", d, args)
	}
}
func (d *knifeSearch) String() string {
	var rawName string
	if d.resultType == publicHostname {
		rawName = nameKnifeHostname
	} else {
		rawName = nameKnife
	}
	return fmt.Sprintf("<%s>", rawName)
}
