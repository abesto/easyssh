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
	return Make(data, makeByName)
}

func Make(data []interface{}, makeByName func(name string) interface{}) interface{} {
	var o = makeByName(string(data[0].([]byte))).(interfaces.HasSetArgs)
	o.SetArgs(data[1:])
	return o
}
