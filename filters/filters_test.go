package filters

import (
	"sort"
	"testing"

	"github.com/abesto/easyssh/util"
)

func TestSupportedFilterNames(t *testing.T) {
	expectedNames := []string{"first", "external", "ec2-instance-id", "list", "id"}
	actualNames := SupportedFilterNames()

	sort.Strings(expectedNames)
	sort.Strings(actualNames)

	if len(expectedNames) != len(actualNames) {
		t.Error("len")
	}
	for i := 0; i < len(expectedNames); i++ {
		if expectedNames[i] != actualNames[i] {
			t.Error(i, expectedNames[i], actualNames[i], expectedNames, actualNames)
		}
	}
}

func TestMakeFilterWrongName(t *testing.T) {
	util.ExpectPanic(t, "filter \"foo-bar\" is not known", func() {
		Make("(list (foo-bar))")
	})
}
