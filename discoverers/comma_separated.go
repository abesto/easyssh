package discoverers

import (
	"fmt"
	"strings"

	"github.com/abesto/easyssh/util"
)

type commaSeparated struct{}

func (d *commaSeparated) isComma(r rune) bool {
	return r == ','
}
func (d *commaSeparated) Discover(input string) []string {
	return strings.FieldsFunc(input, d.isComma)
}
func (d *commaSeparated) SetArgs(args []interface{}) {
	util.RequireNoArguments(d, args)
}
func (d *commaSeparated) String() string {
	return fmt.Sprintf("<%s>", nameCommaSeparated)
}
