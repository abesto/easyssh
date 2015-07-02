package discoverers

import (
	"github.com/abesto/easyssh/fromsexp"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/util"
	"sort"
)

func Make(input string) interfaces.Discoverer {
	return fromsexp.MakeFromString(input, sexpTransforms, makeByName).(interfaces.Discoverer)
}

func SupportedDiscovererNames() []string {
	names := make([]string, len(discovererMakerMap)+len(sexpTransforms))

	// Normal discoverers
	for i := 0; i < len(sexpTransforms); i++ {
		names[i] = sexpTransforms[i].Name
	}

	// Aliases
	i := len(sexpTransforms)
	for key := range discovererMakerMap {
		names[i] = key
		i++
	}

	sort.Strings(names)
	return names
}

func makeFromSExp(data []interface{}) interfaces.Discoverer {
	return fromsexp.Make(data, sexpTransforms, makeByName).(interfaces.Discoverer)
}

const (
	nameCommaSeparated = "comma-separated"
	nameKnife          = "knife"
	nameKnifeHostname  = "knife-hostname"
	nameFirstMatching  = "first-matching"
	nameFixed          = "fixed"
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
	nameFixed:         func() interfaces.Discoverer { return &fixed{} },
}

var sexpTransforms = []fromsexp.SexpTransform{
	fromsexp.Alias("const", nameFixed),
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
