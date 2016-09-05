package target

import (
	"reflect"

	"github.com/stretchr/testify/mock"
)

func AssertTargetListEquals(t mock.TestingT, expected []Target, actual []Target) {
	if len(expected) != len(actual) {
		t.Errorf("len expected=%d actual=%d (expected=%s actual=%s)", len(expected), len(actual), expected, actual)
	}
	for i := 0; i < len(expected); i++ {
		if !reflect.DeepEqual(expected[i], actual[i]) {
			t.Errorf("Lists not equal, first diff at index %d\nExpected %s\nActual %s\nExpected list: %s\nActual list: %s",
				i, expected[i], actual[i], expected, actual)
		}
	}
}
