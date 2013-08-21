package multistep

import (
	"runtime"
)

// A step for testing that accumuluates data into a string slice in the
// the state bag. It always uses the "data" key in the state bag, and will
// initialize it.
type TestStepAcc struct {
	// The data inserted into the state bag.
	Data string

	// If true, it will halt at the step when it is run
	Halt bool
}

func (s TestStepAcc) Run(state map[string]interface{}) StepAction {
	s.insertData(state, "data")

	if s.Halt {
		return ActionHalt
	}

	return ActionContinue
}

func (s TestStepAcc) Cleanup(state map[string]interface{}) {
	s.insertData(state, "cleanup")
}

func (s TestStepAcc) insertData(state map[string]interface{}, key string) {
	if _, ok := state[key]; !ok {
		state[key] = make([]string, 0, 5)
	}

	data := state[key].([]string)
	data = append(data, s.Data)
	state[key] = data
}

type TestStepSync struct {
	C <-chan struct{}
}

func (s TestStepSync) Run(map[string]interface{}) StepAction {
	<-s.C
	runtime.Gosched() // ensure that any calls to cancel have a chance to run
	return ActionContinue
}

func (s TestStepSync) Cleanup(state map[string]interface{}) {
	if _, ok := state[StateCancelled]; ok {
		state["sync_cancelled"] = true
	}
}
