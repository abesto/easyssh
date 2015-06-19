package from_sexp

import (
	"bitbucket.org/shanehanna/sexp"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/util"
)

func MakeFromString(input string, aliases Aliases, makeByName func(name string) interface{}) interface{} {
	var data, error = sexp.Unmarshal([]byte(input))
	if error != nil {
		util.Abort(error.Error())
	}
	util.Logger.Debugf("MakeFromString %s -> %s", input, data)
	var result = Make(data, aliases, makeByName)
	return result
}

func Make(data []interface{}, aliases Aliases, makeByName func(name string) interface{}) interface{} {
	var name = string(data[0].([]byte))
	// Apply any aliases
	if aliases != nil {
		for _, item := range aliases {
			if item.Alias == name {
				util.Logger.Debugf("Alias: %s -> %s", name, item.Name)
				name = item.Name
			}
		}
	}
	// Build using provided constructor
	var o = makeByName(name).(interfaces.HasSetArgs)
	o.SetArgs(data[1:])
	util.Logger.Debugf("Make %s -> %s", data, o)
	return o
}

type Alias struct {
	Name  string
	Alias string
}

type Aliases []Alias
