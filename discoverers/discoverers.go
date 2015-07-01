package discoverers

import (
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
	nameKnife: func() interfaces.Discoverer {
		return &knifeSearch{
			realKnifeSearchResultRowIpExtractor{publicIp}, util.RealCommandRunner{}}
	},
	nameKnifeHostname: func() interfaces.Discoverer {
		return &knifeSearch{
			realKnifeSearchResultRowIpExtractor{publicHostname}, util.RealCommandRunner{}}
	},
	nameFirstMatching: func() interfaces.Discoverer { return &firstMatching{} },
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
