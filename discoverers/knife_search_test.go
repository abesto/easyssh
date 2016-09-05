package discoverers

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

func TestKnifeStringViaMake(t *testing.T) {
	cases := []struct {
		input   string
		structs string
		final   string
	}{
		{input: "(knife)", structs: "[knife]", final: "<knife>"},
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
}

type mockKnifeSearchResultRowExtractor struct {
	mock.Mock
}

func (e *mockKnifeSearchResultRowExtractor) Extract(row knifeSearchResultRow) target.Target {
	r := e.Called(row)
	if len(r) > 0 {
		return r[0].(target.Target)
	}
	return target.Target{}
}

func givenAMockedKnifeSearch() (knifeSearch, *mockKnifeSearchResultRowExtractor, *util.MockCommandRunner) {
	commandRunner := util.MockCommandRunner{}
	ipExtractor := mockKnifeSearchResultRowExtractor{}
	return knifeSearch{&ipExtractor, &commandRunner}, &ipExtractor, &commandRunner
}

func givenKnifeSearchResultWithCloudV2(values ...string) knifeSearchResult {
	data := knifeSearchResult{}
	for _, value := range values {
		row := knifeSearchResultRow{Name: value}
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

func whenKnifeSearch(r *util.MockCommandRunner, input string) *mock.Call {
	return r.On("Outputs", "knife", []string{"search", "node", "-F", "json", input})
}

func TestKnifeNoColonInSearchString(t *testing.T) {
	s, e, r := givenAMockedKnifeSearch()
	input := "no colon at all"
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectDebugf("Host lookup string doesn't contain ':', it won't match anything in a knife search node query")
		if len(s.Discover(input)) != 0 {
			t.Fail()
		}
	})
	e.AssertExpectations(t)
	r.AssertExpectations(t)
}

func TestKnifeError(t *testing.T) {
	s, e, r := givenAMockedKnifeSearch()
	input := "name:whatever"
	err := "knife run failed"
	whenKnifeSearch(r, input).Return(util.CommandRunnerOutputs{Error: errors.New(err), Combined: []byte("Foo\nBar")}).Times(1)
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectInfof("Looking up nodes with knife matching %s", input)
		util.ExpectPanic(t, "Knife lookup failed: knife run failed\nOutput:\nFoo\nBar", func() { s.Discover(input) })
	})
	e.AssertExpectations(t)
	r.AssertExpectations(t)
}

func TestKnifeInvalidJsonFromKnife(t *testing.T) {
	s, e, r := givenAMockedKnifeSearch()
	input := "foo:bar"
	whenKnifeSearch(r, input).Return(util.CommandRunnerOutputs{Stdout: []byte("Invalid JSON")}).Times(1)
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectInfof("Looking up nodes with knife matching %s", input)
		util.ExpectPanic(t,
			"Failed to parse knife search result: invalid character 'I' looking for beginning of value",
			func() { s.Discover(input) })
	})
	e.AssertExpectations(t)
	r.AssertExpectations(t)
}

func TestKnifeHappyPath(t *testing.T) {
	s, e, r := givenAMockedKnifeSearch()
	input := "test:query"
	data := givenKnifeSearchResultWithCloudV2("alpha", "beta", "gamma")
	knifeReturnsWithCloudV2(r, input, data)
	for _, row := range data.Rows {
		e.On("Extract", row).Return(target.Target{Host: row.Automatic.CloudV2.PublicHostname}).Times(1)
	}
	var actualTargets []target.Target
	util.WithLogAssertions(t, func(l *util.MockLogger) {
		l.ExpectInfof("Looking up nodes with knife matching %s", input)
		actualTargets = s.Discover(input)
	})
	expectedTargets := target.FromStrings("alpha.hostname", "beta.hostname", "gamma.hostname")
	target.AssertTargetListEquals(t, expectedTargets, actualTargets)
	e.AssertExpectations(t)
	r.AssertExpectations(t)
}

func TestKnifeExtractor(t *testing.T) {
	cases := []struct {
		input          knifeSearchResultRow
		expectedOutput target.Target
	}{
		{
			expectedOutput: target.Target{IP: "a.ip", Host: "a.host", Hostname: "a.hostname"},
			input: knifeSearchResultRow{
				Automatic: knifeSearchResultRowAutomatic{
					Hostname: "a.hostname",
					CloudV2: &knifeSearchResultCloudV2{
						PublicIpv4:     "a.ip",
						PublicHostname: "a.host",
					}},
			},
		},
		{
			expectedOutput: target.Target{IP: "b.noncloud-ip", Host: "b.fqdn", Hostname: "b.hostname"},
			input: knifeSearchResultRow{
				Automatic: knifeSearchResultRowAutomatic{
					Hostname:  "b.hostname",
					Ipaddress: "b.noncloud-ip",
					Fqdn:      "b.fqdn",
				}},
		},
	}
	for _, c := range cases {
		e := realKnifeSearchResultRowExtractor{}
		if actualOutput := e.Extract(c.input); actualOutput != c.expectedOutput {
			t.Error("output", c.expectedOutput, actualOutput)
		}
	}
}
