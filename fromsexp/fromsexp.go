package fromsexp

import (
	"bitbucket.org/shanehanna/sexp"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/util"
	"reflect"
)

func MakeFromString(input string, transforms []SexpTransform, makeByName func(name string) interface{}) interface{} {
	data, err := sexp.Unmarshal([]byte(input))
	if err != nil {
		util.Panicf(err.Error())
	}
	util.Logger.Debugf("MakeFromString %s -> %s", input, data)
	var result = Make(data, transforms, makeByName)
	return result
}

func Make(data []interface{}, transforms []SexpTransform, makeByName func(name string) interface{}) interface{} {
	// Apply any transforms
	if transforms != nil {
		for _, item := range transforms {
			if item.Matches(data) {
				newData := item.Transform(data)
				util.Logger.Debugf("Transform: %s -> %s", data, newData)
				data = newData
			}
		}
	}
	// Build using provided constructor
	//	util.Logger.Debug(data)
	var name = string(data[0].([]byte))
	var o = makeByName(name).(interfaces.HasSetArgs)
	o.SetArgs(data[1:])
	util.Logger.Debugf("Make %s -> %s", data, o)
	return o
}

type SexpTransformMatcher func(input []interface{}) bool
type SexpTransformFunction func(input []interface{}) []interface{}

type SexpTransform struct {
	Name      string
	Matches   SexpTransformMatcher
	Transform SexpTransformFunction
}

func Replace(original string, replacement string) SexpTransform {
	originalData, err := sexp.Unmarshal([]byte(original))
	if err != nil {
		panic(err)
	}
	replacementData, err := sexp.Unmarshal([]byte(replacement))
	if err != nil {
		panic(err)
	}
	return SexpTransform{
		Name:    string(originalData[0].([]byte)),
		Matches: func(input []interface{}) bool { return reflect.DeepEqual(originalData, input) },
		Transform: func(input []interface{}) []interface{} {
			return replacementData
		},
	}
}
