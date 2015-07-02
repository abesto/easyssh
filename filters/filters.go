package filters

import (
	"github.com/abesto/easyssh/fromsexp"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/util"
	"sort"
)

func Make(input string) interfaces.TargetFilter {
	return fromsexp.MakeFromString(input, nil, makeByName).(interfaces.TargetFilter)
}

func SupportedFilterNames() []string {
	keys := make([]string, len(filterMakerMap))
	i := 0
	for key := range filterMakerMap {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

func makeFromSExp(data []interface{}) interfaces.TargetFilter {
	return fromsexp.Make(data, nil, makeByName).(interfaces.TargetFilter)
}

const (
	nameEc2InstanceId = "ec2-instance-id"
	nameList          = "list"
	nameId            = "id"
	nameFirst         = "first"
	nameExternal      = "external"
)

var filterMakerMap = map[string]func() interfaces.TargetFilter{
	nameEc2InstanceId: func() interfaces.TargetFilter {
		return &ec2InstanceIdLookup{
			idParser: realEc2InstanceIdParser{}, commandRunner: util.RealCommandRunner{}}
	},
	nameList:  func() interfaces.TargetFilter { return &list{} },
	nameId:    func() interfaces.TargetFilter { return &id{} },
	nameFirst: func() interfaces.TargetFilter { return &first{} },
	nameExternal: func() interfaces.TargetFilter {
		return &external{
			commandRunner: util.RealCommandRunner{},
			tmpFileMaker:  &realTmpFileMaker{},
		}
	},
}

func makeByName(name string) interface{} {
	var d interfaces.TargetFilter
	for key, maker := range filterMakerMap {
		if key == name {
			d = maker()
		}
	}
	if d == nil {
		util.Panicf("filter \"%s\" is not known", name)
	}
	return d
}
