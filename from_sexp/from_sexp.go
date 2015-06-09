package from_sexp

import (
	"bitbucket.org/shanehanna/sexp"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/util"
)

func MakeFromString(input string, makeByName func(name string) interface{}) interface{} {
	var data, error = sexp.Unmarshal([]byte(input))
	if error != nil {
		util.Abort(error.Error())
	}
	util.Logger.Debugf("MakeFromString %s -> %s", input, data)
	var result = Make(data, makeByName)
	return result
}

func Make(data []interface{}, makeByName func(name string) interface{}) interface{} {
	var o = makeByName(string(data[0].([]byte))).(interfaces.HasSetArgs)
	o.SetArgs(data[1:])
	util.Logger.Debugf("Make %s -> %s", data, o)
	return o
}
