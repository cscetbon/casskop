package actionstep

type StepResult interface {
	HasError() bool
	Error() error
	ShouldBreakReconcileLoop() bool
}

func Error(err error) StepResult {
	return errorResult{err: err}
}

func Break() StepResult {
	return successfulBreakSingleton
}

func Pass() StepResult {
	return passSingleton
}

type errorResult struct {
	err error
}

var _ StepResult = errorResult{}

func (e errorResult) HasError() bool {
	return e.err != nil
}

func (e errorResult) Error() error {
	return e.err
}

func (e errorResult) ShouldBreakReconcileLoop() bool {
	return true
}

type successfulBreak struct{}

var successfulBreakSingleton StepResult = successfulBreak{}

func (b successfulBreak) HasError() bool {
	return false
}

func (b successfulBreak) Error() error {
	return nil
}

func (b successfulBreak) ShouldBreakReconcileLoop() bool {
	return true
}

type pass struct{}

var passSingleton StepResult = pass{}

func (b pass) HasError() bool {
	return false
}

func (b pass) Error() error {
	return nil
}

func (b pass) ShouldBreakReconcileLoop() bool {
	return false
}
