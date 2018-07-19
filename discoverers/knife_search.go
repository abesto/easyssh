package discoverers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

type knifeSearch struct {
	extractor     knifeSearchResultRowExtractor
	commandRunner util.CommandRunner
}

type knifeSearchResult struct {
	Rows []knifeSearchResultRow
}

type knifeSearchResultRow struct {
	Name      string
	Automatic knifeSearchResultRowAutomatic
}

type knifeSearchResultRowAutomatic struct {
	CloudV2   *knifeSearchResultCloudV2 `json:"cloud_v2"`
	Ipaddress string
	Hostname  string
	Fqdn      string
}

type knifeSearchResultCloudV2 struct {
	PublicHostname string `json:"public_hostname"`
	PublicIpv4     string `json:"public_ipv4"`
	LocalHostname  string `json:"local_hostname"`
	LocalIpv4      string `json:"local_ipv4"`
}

type knifeSearchResultRowExtractor interface {
	Extract(knifeSearchResultRow) target.Target
}

type realKnifeSearchResultRowExtractor struct {
}

func (e realKnifeSearchResultRowExtractor) Extract(row knifeSearchResultRow) target.Target {
	var target target.Target
	if row.Automatic.CloudV2 != nil && row.Automatic.CloudV2.PublicHostname != "" {
		target.Host = row.Automatic.CloudV2.PublicHostname
		target.IP = row.Automatic.CloudV2.PublicIpv4
	} else if row.Automatic.CloudV2 != nil && row.Automatic.CloudV2.LocalHostname != "" {
		target.Host = row.Automatic.CloudV2.LocalHostname
		target.IP = row.Automatic.CloudV2.LocalIpv4
	} else {
		target.Host = row.Automatic.Fqdn
		target.IP = row.Automatic.Ipaddress
	}
	target.Hostname = row.Automatic.Hostname
	return target
}

func (d *knifeSearch) Discover(input string) []target.Target {
	var targets []target.Target

	if !strings.Contains(input, ":") {
		util.Logger.Debugf("Host lookup string doesn't contain ':', it won't match anything in a knife search node query")
		return targets
	}

	util.Logger.Infof("Looking up nodes with knife matching %s", input)
	outputs := d.commandRunner.Outputs("knife", []string{"search", "node", "-F", "json", input})
	if outputs.Error != nil {
		util.Panicf("Knife lookup failed: %s\nOutput:\n%s", outputs.Error, outputs.Combined)
	}

	data := knifeSearchResult{}
	if err := json.Unmarshal(outputs.Stdout, &data); err != nil {
		util.Panicf("Failed to parse knife search result: %s", err)
	}
	//	util.Logger.Debugf("Parsed result into data: %s", data)

	for _, row := range data.Rows {
		target := d.extractor.Extract(row)
		if target.IsEmpty() {
			util.Logger.Infof("Host %s doesn't have an IP address or public hostname, ignoring", row.Name)
		} else {
			targets = append(targets, target)
		}
	}

	return targets
}

func (d *knifeSearch) SetArgs(args []interface{}) {
	util.RequireNoArguments(d, args)
	util.RequireOnPath(d, "knife")
}

func (d *knifeSearch) String() string {
	return fmt.Sprintf("<%s>", nameKnife)
}
