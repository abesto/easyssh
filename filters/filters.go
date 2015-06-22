package filters

import (
	"fmt"
	"github.com/abesto/easyssh/from_sexp"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
	"io/ioutil"
	"os"
	"strings"
)

func Make(input string) interfaces.TargetFilter {
	return from_sexp.MakeFromString(input, nil, makeByName).(interfaces.TargetFilter)
}

func SupportedFilterNames() []string {
	var keys = make([]string, len(filterMakerMap))
	var i = 0
	for key := range filterMakerMap {
		keys[i] = key
		i += 1
	}
	return keys
}

func makeFromSExp(data []interface{}) interfaces.TargetFilter {
	return from_sexp.Make(data, nil, makeByName).(interfaces.TargetFilter)
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
			commandRunner: util.RealCommandRunner{}}
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

type first struct{}

func (f *first) Filter(targets []target.Target) []target.Target {
	if len(targets) > 0 {
		return targets[0:1]
	}
	return targets
}
func (f *first) SetArgs(args []interface{}) {
	util.RequireNoArguments(f, args)
}
func (f *first) String() string {
	return fmt.Sprintf("<%s>", nameFirst)
}

type external struct {
	argv          []string
	commandRunner util.CommandRunner
}

func (f *external) Filter(targets []target.Target) []target.Target {
	var err error
	var output []byte
	var tmpFile *os.File
	tmpFile, err = ioutil.TempFile("", "easyssh")
	defer os.Remove(tmpFile.Name())
	if err != nil {
		util.Panicf(err.Error())
	}
	tmpFile.Write([]byte(strings.Join(target.TargetStrings(targets), "\n")))
	output = f.commandRunner.RunWithStdinGetOutputOrPanic(f.argv[0], append(f.argv[1:], tmpFile.Name()))
	var lines = strings.Split(strings.TrimSpace(string(output)), "\n")
	var newTargets = make([]target.Target, len(lines))
	var i int
	for i = 0; i < len(lines); i++ {
		newTargets[i] = target.FromString(lines[i])
	}
	return newTargets
}
func (f *external) SetArgs(args []interface{}) {
	util.RequireArguments(f, 1, args)
	f.argv = make([]string, len(args))
	for i := 0; i < len(args); i++ {
		f.argv[i] = string(args[i].([]uint8))
	}
}
func (f *external) String() string {
	return fmt.Sprintf("<%s %s>", nameExternal, f.argv)
}
