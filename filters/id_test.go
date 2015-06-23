package filters

import (
	"testing"
)

func TestIdStringViaMake(t *testing.T) {
	withLogAssertions(t, func(l *mockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(id)", "[id]")
		l.ExpectDebugf("Make %s -> %s", "[id]", "<id>")
		Make("(id)")
	})
}

func TestIdMakeWithArgument(t *testing.T) {
	withLogAssertions(t, func(l *mockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(id foo)", "[id foo]")
		expectPanic(t, "<id> doesn't take any arguments, got 1: [foo]", func() { Make("(id foo)") })
	})
}

func TestIdOperation(t *testing.T) {
	f := Make("(id)").(*id)

	before := givenTargets("one", "two")
	after := f.Filter(before)
	if len(after) != len(before) || before[0] != after[0] || before[1] != after[1] {
		t.Error(before, after)
	}
}
