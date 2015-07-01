package discoverers

import (
	"testing"

	"github.com/abesto/easyssh/util"
)

func TestKnifeSearchStringViaMake(t *testing.T) {
	cases := []struct {
		input   string
		structs string
		final   string
	}{
		{input: "(knife)", structs: "[knife]", final: "<knife>"},
		{input: "(knife-hostname)", structs: "[knife-hostname]", final: "<knife-hostname>"},
	}
	for _, c := range cases {
		util.WithLogAssertions(t, func(l *util.MockLogger) {
			l.ExpectDebugf("MakeFromString %s -> %s", c.input, c.structs)
			l.ExpectDebugf("Make %s -> %s", c.structs, c.final)
			d := Make(c.input)
			if d.String() != c.final {
				t.Error(d)
			}
		})
	}
}


