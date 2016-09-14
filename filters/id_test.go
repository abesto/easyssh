package filters

import (
	"reflect"
	"testing"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

func TestIdStringViaMake(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(id)", "[id]")
		l.ExpectDebugf("Make %s -> %s", "[id]", "<id>")
		Make("(id)")
	})
}

func TestIdMakeWithArgument(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(id foo)", "[id foo]")
		util.ExpectPanic(t, "<id> doesn't take any arguments, got 1: [foo]", func() { Make("(id foo)") })
	})
}

func TestIdOperation(t *testing.T) {
	f := Make("(id)").(*id)

	before := target.FromStrings("one", "two")
	after := f.Filter(before)
	if !reflect.DeepEqual(before, after) {
		t.Error(before, after)
	}
}
