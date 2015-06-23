package filters

import (
	"fmt"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
	"github.com/alexcesaro/log"
	"github.com/maraino/go-mock"
	"io"
	"testing"
)

func expectPanic(t *testing.T, expectedErr interface{}, f func()) {
	defer func() {
		actualErr := recover()
		if actualErr == nil {
			t.Errorf("Expected panic(\"%s\"), got no panic", expectedErr)
		}
		if actualErr != expectedErr {
			t.Errorf("Expected panic(\"%s\"), got panic(\"%s\") instead", expectedErr, actualErr)
		}
	}()
	f()
}

type mockCommandRunner struct {
	mock.Mock
}

func (r *mockCommandRunner) RunWithStdinGetOutputOrPanic(stdin io.Reader, name string, args []string) []byte {
	ret := r.Called(stdin, name, args)
	return ret.Bytes(0)
}
func (r *mockCommandRunner) RunGetOutputOrPanic(name string, args []string) []byte {
	ret := r.Called(name, args)
	return ret.Bytes(0)
}
func (r *mockCommandRunner) RunGetOutput(name string, args []string) ([]byte, error) {
	ret := r.Called(name, args)
	return ret.Bytes(0), ret.Error(1)
}

type mockLogger struct {
	mock.Mock
}

func toStrings(xs []interface{}) []interface{} {
	ss := make([]interface{}, len(xs))
	for i, x := range xs {
		ss[i] = fmt.Sprintf("%s", x)
	}
	return ss
}
func (m mockLogger) Emergency(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m mockLogger) Emergencyf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m mockLogger) Alert(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m mockLogger) Alertf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m mockLogger) Critical(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m mockLogger) Criticalf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m mockLogger) Error(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m mockLogger) Errorf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m mockLogger) Warning(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m mockLogger) Warningf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m mockLogger) Notice(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m mockLogger) Noticef(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m mockLogger) Info(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m mockLogger) Infof(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m mockLogger) Debug(args ...interface{}) {
	m.Called(toStrings(args)...)
}
func (m mockLogger) Debugf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, toStrings(args)...)...)
}
func (m mockLogger) Log(level log.Level, args ...interface{}) {
	m.Called(append([]interface{}{level}, toStrings(args)...)...)
}
func (m mockLogger) Logf(level log.Level, format string, args ...interface{}) {
	m.Called(append([]interface{}{level, format}, toStrings(args)...)...)
}
func (m mockLogger) LogEmergency() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m mockLogger) LogAlert() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m mockLogger) LogCritical() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m mockLogger) LogError() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m mockLogger) LogWarning() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m mockLogger) LogNotice() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m mockLogger) LogInfo() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m mockLogger) LogDebug() bool {
	ret := m.Called()
	return ret.Bool(0)
}
func (m mockLogger) LogLevel(level log.Level) bool {
	ret := m.Called(level)
	return ret.Bool(0)
}
func (m mockLogger) Close() error {
	ret := m.Called()
	return ret.Error(0)
}

func (m *mockLogger) ExpectDebugf(format string, args ...interface{}) *mock.MockFunction {
	return m.When("Debugf", append([]interface{}{format}, args...)...).Times(1)
}
func (m *mockLogger) ExpectInfof(format string, args ...interface{}) *mock.MockFunction {
	return m.When("Infof", append([]interface{}{format}, args...)...).Times(1)
}

type hasVerify interface {
	Verify() (bool, error)
}

type dummyError struct {
	msg string
}

func (e dummyError) Error() string {
	return e.msg
}

func expectLogs(t *testing.T, setExpectedCalls func(*mockLogger)) func() {
	originalLogger := util.Logger
	util.Logger = &mockLogger{}
	l := util.Logger.(*mockLogger)
	l.Reset()
	setExpectedCalls(l)
	return func() {
		verifyMocks(t, l)
		util.Logger = originalLogger
	}
}

func withLogAssertions(t *testing.T, f func(*mockLogger)) {
	expectLogs(t, f)()
}

func verifyMocks(t *testing.T, mocks ...hasVerify) {
	for _, m := range mocks {
		if ok, msg := m.Verify(); !ok {
			t.Error(msg)
		}
	}
}

func givenTargets(targetStrings ...string) []target.Target {
	targets := make([]target.Target, len(targetStrings))
	for i := 0; i < len(targetStrings); i++ {
		targets[i] = target.FromString(targetStrings[i])
	}
	return targets
}
