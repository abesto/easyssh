package util

import (
	"fmt"
	"io"

	"reflect"

	"github.com/alexcesaro/log"
	"github.com/stretchr/testify/mock"
)

func ExpectPanic(t mock.TestingT, expectedErr interface{}, f func()) {
	defer func() {
		actualErr := recover()
		if actualErr == nil {
			t.Errorf("Expected panic(\"%s\"), got no panic", expectedErr)
		}
		if actualErr != expectedErr && expectedErr != nil {
			t.Errorf("Expected panic(\"%s\"), got panic(\"%s\") instead", expectedErr, actualErr)
		}
	}()
	f()
}

type MockCommandRunner struct {
	mock.Mock
}

func (r *MockCommandRunner) CombinedOutputWithStdinOrPanic(stdin io.Reader, name string, args []string) []byte {
	ret := r.Called(stdin, name, args)
	return ret.Get(0).([]byte)
}
func (r *MockCommandRunner) CombinedOutputOrPanic(name string, args []string) []byte {
	ret := r.Called(name, args)
	return ret.Get(0).([]byte)
}
func (r *MockCommandRunner) Outputs(name string, args []string) CommandRunnerOutputs {
	ret := r.Called(name, args)
	return ret.Get(0).(CommandRunnerOutputs)
}

type MockInteractiveCommandRunner struct {
	mock.Mock
}

func (r *MockInteractiveCommandRunner) Run(job InteractiveCommandRunnerJob) {
	r.Called(job)
}

func (r *MockInteractiveCommandRunner) RunParallel(jobs []InteractiveCommandRunnerJob) {
	r.Called(jobs)
}

type MockLogger struct {
	mock.Mock
}

func toStrings(xs []interface{}) []interface{} {
	ss := make([]interface{}, len(xs))
	for i, x := range xs {
		ss[i] = fmt.Sprintf("%s", x)
	}
	return ss
}
func (m *MockLogger) Emergency(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m *MockLogger) Emergencyf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m *MockLogger) Alert(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m *MockLogger) Alertf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m *MockLogger) Critical(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m *MockLogger) Criticalf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m *MockLogger) Error(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m *MockLogger) Warning(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m *MockLogger) Warningf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m *MockLogger) Notice(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m *MockLogger) Noticef(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m *MockLogger) Info(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m *MockLogger) Debug(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m *MockLogger) Debugf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m *MockLogger) Log(level log.Level, args ...interface{}) {
	m.Called(append([]interface{}{level}, toStrings(args)...)...)
}
func (m *MockLogger) Logf(level log.Level, format string, args ...interface{}) {
	m.Called(append([]interface{}{level, format}, toStrings(args)...)...)
}
func (m *MockLogger) LogEmergency() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m *MockLogger) LogAlert() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m *MockLogger) LogCritical() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m *MockLogger) LogError() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m *MockLogger) LogWarning() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m *MockLogger) LogNotice() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m *MockLogger) LogInfo() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m *MockLogger) LogDebug() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m *MockLogger) LogLevel(level log.Level) bool {
	ret := m.Called(level)
	return ret.Bool(0)
}
func (m *MockLogger) Close() error {
	ret := m.Called()
	return ret.Error(0)
}

func (m *MockLogger) ExpectDebugf(format string, args ...interface{}) *mock.Call {
	return m.On("Debugf", append([]interface{}{format}, args...)...).Times(1)
}
func (m *MockLogger) ExpectInfof(format string, args ...interface{}) *mock.Call {
	return m.On("Infof", append([]interface{}{format}, args...)...).Times(1)
}

type DummyError struct {
	Msg string
}

func (e DummyError) Error() string {
	return e.Msg
}

func ExpectLogs(t mock.TestingT, setExpectedCalls func(*MockLogger)) func() {
	originalLogger := Logger
	Logger = &MockLogger{}
	l := Logger.(*MockLogger)
	setExpectedCalls(l)
	return func() {
		l.AssertExpectations(t)
		Logger = originalLogger
	}
}

func WithLogAssertions(t mock.TestingT, f func(*MockLogger)) {
	ExpectLogs(t, f)()
}

func AssertStringListEquals(t mock.TestingT, expected []string, actual []string) {
	expectedInterfaces := make([]interface{}, len(expected))
	actualInterfaces := make([]interface{}, len(actual))
	for i := 0; i < len(expected); i++ {
		expectedInterfaces[i] = expected[i]
	}
	for i := 0; i < len(actual); i++ {
		actualInterfaces[i] = actual[i]
	}
	AssertInterfaceListEquals(t, expectedInterfaces, actualInterfaces)
}

func AssertInterfaceListEquals(t mock.TestingT, expected []interface{}, actual []interface{}) {
	if len(expected) != len(actual) {
		t.Errorf("len expected=%d actual=%d", len(expected), len(actual))
	}
	for i := 0; i < len(expected); i++ {
		if !reflect.DeepEqual(expected[i], actual[i]) {
			t.Errorf("Lists not equal, first diff at index %d\nExpected %s\nActual %s\nExpected list: %s\nActual list: %s",
				i, expected[i], actual[i], expected, actual)
		}
	}
}
