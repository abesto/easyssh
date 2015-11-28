package fromsexp

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/abesto/sexp"
	"github.com/maraino/go-mock"

	"github.com/abesto/easyssh/util"
)

type MockWithMakeByName struct {
	mock.Mock
}

func (m *MockWithMakeByName) makeByName(name string) interface{} {
	return m.Called(name).Get(0)
}

type MockHasSetArgs struct {
	mock.Mock
}

func (s *MockHasSetArgs) SetArgs(args []interface{}) {
	s.Called(fmt.Sprintf("%s", args))
}

func TestAlias(t *testing.T) {
	transform := Alias("before", "after")
	cases := []struct {
		input    string
		expected string
	}{
		{"()", "()"},
		{"(foo)", "(foo)"},
		{"(before-stays x)", "(before-stays x)"},
		{"(before x)", "(after x)"},
	}
	for _, item := range cases {
		inputData, _ := sexp.Unmarshal([]byte(item.input))
		expectedData, _ := sexp.Unmarshal([]byte(item.expected))
		actualData := transform.TransformIfMatches(inputData)
		if !reflect.DeepEqual(expectedData, actualData) {
			t.Errorf("%v returned %s for input %s. Expected %s.", transform, actualData, inputData, expectedData)
		}
	}
}

func TestReplace(t *testing.T) {
	transform := Replace("(before (x))", "(after (y))")
	cases := []struct {
		input    string
		expected string
	}{
		{"()", "()"},
		{"(foo)", "(foo)"},
		{"(before)", "(before)"},
		{"(before-stays x)", "(before-stays x)"},
		{"(before x)", "(before x)"},
		{"(before (x))", "(after (y))"},
		{"(before foobar)", "(before foobar)"},
	}
	for _, item := range cases {
		inputData, _ := sexp.Unmarshal([]byte(item.input))
		expectedData, _ := sexp.Unmarshal([]byte(item.expected))
		actualData := transform.TransformIfMatches(inputData)
		if !reflect.DeepEqual(expectedData, actualData) {
			t.Errorf("%v returned %s for input %s. Expected %s.", transform, actualData, inputData, expectedData)
		}
	}
}

func TestMakeFromStringWithoutTransforms(t *testing.T) {
	m := &MockWithMakeByName{}
	input := "(foo bar baz)"
	expectedFoo := &MockHasSetArgs{}

	m.When("makeByName", "foo").Times(1).Return(expectedFoo)
	expectedFoo.When("SetArgs", "[bar baz]").Times(1)
	actualFoo := MakeFromString(input, []SexpTransform{}, m.makeByName)

	if actualFoo != expectedFoo {
		t.Errorf("MakeFromString returned %v, expected: %v", actualFoo, expectedFoo)
	}

	mock.AssertVerifyMocks(t, m, expectedFoo)
}

func TestMakeFromStringWithTransforms(t *testing.T) {
	m := &MockWithMakeByName{}
	input := "(aaa (xxx (yyy)))"
	expectedFoo := &MockHasSetArgs{}
	transforms := []SexpTransform{
		Alias("aaa", "say"),
		Replace("(say (xxx (yyy)))", "(say hello world)"),
	}

	m.When("makeByName", "say").Times(1).Return(expectedFoo)
	expectedFoo.When("SetArgs", "[hello world]").Times(1)
	actualFoo := MakeFromString(input, transforms, m.makeByName)

	if actualFoo != expectedFoo {
		t.Errorf("MakeFromString returned %v, expected: %v", actualFoo, expectedFoo)
	}

	mock.AssertVerifyMocks(t, m, expectedFoo)
}

func TestInvalidInputs(t *testing.T) {
	makeByName := func(s string) interface{} { return nil }
	transforms := []SexpTransform{}
	cases := [](func()){
		func() { MakeFromString("---qqqq", transforms, makeByName) },
		func() { Replace("---", "(x)") },
		func() { Replace("(x)", "---") },
	}
	for _, item := range cases {
		util.ExpectPanic(t, nil, item)
	}
}
