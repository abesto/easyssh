package from_sexp

import (
	"github.com/eadmund/sexprs"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/util"
)

func MakeFromString(input string, makeByName func(name string) interface{}) interface{} {
	var data, _, error = sexprs.Parse([]byte(input))
	if error != nil {
		util.Abort(error.Error())
	}
	util.Logger.Debugf("MakeFromString %s -> %s", input, data)
	var result = Make(data, makeByName)
	return result
}

func Make(data sexprs.Sexp, makeByName func(name string) interface{}) interface{} {
	var o interfaces.HasSetArgs
	if sexprs.IsList(data) {
		var list = data.(sexprs.List)
		var nameExpr = list[0]
		if sexprs.IsList(nameExpr) {
			util.Abort("NOOOOOO")
		}
		var name = nameExpr.String()
		o = makeByName(name).(interfaces.HasSetArgs)

		var args = make([]sexprs.Sexp, len(list) - 1)
		for i := 1; i < len(list); i++ {
			args[i-1] = list[i]
		}
		o.SetArgs(args)
		util.Logger.Debugf("Make %s -> %s", data, o)
	} else {
		util.Abort("OH NO")
	}
	return o
}
