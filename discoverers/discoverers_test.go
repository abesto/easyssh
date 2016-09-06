package discoverers

import (
	"testing"

	"github.com/abesto/easyssh/util"
)

func TestSupportedDiscovererNames(t *testing.T) {
	util.AssertStringListEquals(t,
		[]string{"comma-separated", "const", "first-matching", "fixed", "knife", "separated-by"},
		SupportedDiscovererNames())
}
