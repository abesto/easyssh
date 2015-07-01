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
	extractor     knifeSearchResultRowIpExtractor
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
}

type knifeSearchResultCloudV2 struct {
	PublicHostname string `json:"public_hostname"`
	PublicIpv4     string `json:"public_ipv4"`
}

type knifeSearchResultRowIpExtractor interface {
	Extract(knifeSearchResultRow) string
	GetResultType() knifeSearchResultType
}

type realKnifeSearchResultRowIpExtractor struct {
	resultType knifeSearchResultType
}

func (e realKnifeSearchResultRowIpExtractor) Extract(row knifeSearchResultRow) string {
	if row.Automatic.CloudV2 != nil {
		if e.resultType == publicHostname {
			return row.Automatic.CloudV2.PublicHostname
		}
		return row.Automatic.CloudV2.PublicIpv4
	}
	return row.Automatic.Ipaddress
}

func (e realKnifeSearchResultRowIpExtractor) GetResultType() knifeSearchResultType {
	return e.resultType
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

	data := knifeSearchResult{}
	if err := json.Unmarshal(outputs.Stdout, &data); err != nil {
		util.Panicf("Failed to parse knife search result: %s", err)
	}
	//	util.Logger.Debugf("Parsed result into data: %s", data)

	var ips = []string{}
	for _, row := range data.Rows {
		ips = append(ips, d.extractor.Extract(row))
	}

	return ips
}
func (d *knifeSearch) SetArgs(args []interface{}) {
	util.RequireNoArguments(d, args)
}

func (d *knifeSearch) String() string {
	var rawName string
	if d.extractor.GetResultType() == publicHostname {
		rawName = nameKnifeHostname
	} else {
		rawName = nameKnife
	}
	return fmt.Sprintf("<%s>", rawName)
}
