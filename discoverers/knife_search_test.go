package discoverers

import (
	"testing"

	"encoding/json"

	"errors"

	"github.com/abesto/easyssh/util"
	"github.com/maraino/go-mock"
)

func TestKnifeStringViaMake(t *testing.T) {
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

func TestKnifeMakeWithArgument(t *testing.T) {
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(knife foo)", "[knife foo]")
		util.ExpectPanic(t, "<knife> doesn't take any arguments, got 1: [foo]", func() { Make("(knife foo)") })
	})
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("MakeFromString %s -> %s", "(knife-hostname foo)", "[knife-hostname foo]")
		util.ExpectPanic(t, "<knife-hostname> doesn't take any arguments, got 1: [foo]", func() { Make("(knife-hostname foo)") })
	})
}

type mockKnifeSearchResultRowIpExtractor struct {
	mock.Mock
}

func (e mockKnifeSearchResultRowIpExtractor) Extract(row knifeSearchResultRow) string {
	return e.Called(row).String(0)
}

func (e mockKnifeSearchResultRowIpExtractor) GetResultType() knifeSearchResultType {
	return e.Called().GetType(0, publicHostname).(knifeSearchResultType)
}

func givenAMockedKnifeSearch() (knifeSearch, *mockKnifeSearchResultRowIpExtractor, *util.MockCommandRunner) {
	commandRunner := util.MockCommandRunner{}
	ipExtractor := mockKnifeSearchResultRowIpExtractor{}
	return knifeSearch{&ipExtractor, &commandRunner}, &ipExtractor, &commandRunner
}

func givenKnifeSearchResultWithCloudV2(values ...string) knifeSearchResult {
	data := knifeSearchResult{}
	for _, value := range values {
		row := knifeSearchResultRow{}
		row.Automatic.CloudV2 = new(knifeSearchResultCloudV2)
		row.Automatic.CloudV2.PublicHostname = value + ".hostname"
		row.Automatic.CloudV2.PublicIpv4 = value + ".ipv4"
		data.Rows = append(data.Rows, row)
	}
	return data
}

func knifeReturnsWithCloudV2(r *util.MockCommandRunner, input string, data knifeSearchResult) {
	jsonData, _ := json.Marshal(data)
	outputs := util.CommandRunnerOutputs{
		Error:    nil,
		Stdout:   jsonData,
		Combined: []byte("Junk\n" + string(jsonData) + "\nMore junk\n"),
	}
	whenKnifeSearch(r, input).Return(outputs)
}

func whenKnifeSearch(r *util.MockCommandRunner, input string) *mock.MockFunction {
	return r.When("Outputs", "knife", []string{"search", "node", "-F", "json", input})
}

func TestKnifeNoColonInSearchString(t *testing.T) {
	s, e, r := givenAMockedKnifeSearch()
	input := "no colon at all"
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("Host lookup string doesn't contain ':', it won't match anything in a knife search node query")
		e.When("Extract", mock.Any).Times(0)
		r.When("Outputs", mock.Any, mock.Any).Times(0)
		if len(s.Discover(input)) != 0 {
			t.Fail()
		}
	})
	util.VerifyMocks(t, e, r)
}

func TestKnifeError(t *testing.T) {
	s, e, r := givenAMockedKnifeSearch()
	input := "name:whatever"
	err := "knife run failed"
	whenKnifeSearch(r, input).Return(util.CommandRunnerOutputs{Error: errors.New(err), Combined: []byte("Foo\nBar")}).Times(1)
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectInfof("Looking up nodes with knife matching %s", input)
		e.When("Extract", mock.Any).Times(0)
		util.ExpectPanic(t, "Knife lookup failed: knife run failed\nOutput:\nFoo\nBar", func() { s.Discover(input) })
	})
	util.VerifyMocks(t, e, r)
}

func TestKnifeInvalidJsonFromKnife(t *testing.T) {
	s, e, r := givenAMockedKnifeSearch()
	input := "foo:bar"
	whenKnifeSearch(r, input).Return(util.CommandRunnerOutputs{Stdout: []byte("Invalid JSON")}).Times(1)
	e.When("Extract", mock.Any).Times(0)
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectInfof("Looking up nodes with knife matching %s", input)
		util.ExpectPanic(t,
			"Failed to parse knife search result: invalid character 'I' looking for beginning of value",
			func() { s.Discover(input) })
	})
	util.VerifyMocks(t, e, r)
}

func TestKnifeHappyPath(t *testing.T) {
	s, e, r := givenAMockedKnifeSearch()
	input := "test:query"
	data := givenKnifeSearchResultWithCloudV2("alpha", "beta", "gamma")
	knifeReturnsWithCloudV2(r, input, data)
	for _, row := range data.Rows {
		e.When("Extract", row).Return(row.Automatic.CloudV2.PublicHostname).Times(1)
	}
	var actualIps []string
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectInfof("Looking up nodes with knife matching %s", input)
		actualIps = s.Discover(input)
	})
	expectedIps := []string{"alpha.hostname", "beta.hostname", "gamma.hostname"}
	util.AssertStringListEquals(t, expectedIps, actualIps)
	util.VerifyMocks(t, e, r)
}
