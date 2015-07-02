package fromsexp

import (
	"bitbucket.org/shanehanna/sexp"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/util"
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

func isFirstItem(input []interface{}, this string) bool {
	switch input[0].(type) {
	case []byte:
		return string(input[0].([]byte)) == this
	default:
		return false
	}
}

func Alias(from string, to string) SexpTransform {
	return SexpTransform{
		Name:    from,
		Matches: func(input []interface{}) bool { return isFirstItem(input, from) },
		Transform: func(input []interface{}) []interface{} {
			output := make([]interface{}, len(input))
			copy(output, input)
			output[0] = []byte(to)
			return output
		},
	}
}

func wrap(input []interface{}, with string) []interface{} {
	return []interface{}{[]byte(with), input}
}

func WrapAndReplaceHead(ifNameIs string, wrapWith []string, replaceHeadWith []string) SexpTransform {
	return SexpTransform{
		Name:    ifNameIs,
		Matches: func(input []interface{}) bool { return isFirstItem(input, ifNameIs) },
		Transform: func(data []interface{}) []interface{} {
			replaceHeadWithInterfaces := make([]interface{}, len(replaceHeadWith))
			for i, s := range replaceHeadWith {
				replaceHeadWithInterfaces[i] = []byte(s)
			}
			data = append(replaceHeadWithInterfaces, data[1:]...)
			for _, wrapper := range wrapWith {
				data = wrap(data, wrapper)
			}
			return data
		},
	}
}
