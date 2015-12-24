package executors

import (
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/go-mock"
)

type mockExecutor struct {
	mock.Mock
}

func (e *mockExecutor) Exec(targets []target.Target, args []string) {
	e.Called(targets, args)
}
func (e *mockExecutor) SetArgs(args []interface{}) {
	// We don't actually want to assert on this, so no call to e.Called
}
func (e *mockExecutor) String() string {
	return "<mock>"
}

func withMockInMakerMap(f func()) {
	executorMakerMap["mock"] = func() interfaces.Executor { return &mockExecutor{} }
	f()
	delete(executorMakerMap, "mock")
}
