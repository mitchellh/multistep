package multistep

// A step for testing that accumuluates data into a string slice in the
// the state bag. It always uses the "data" key in the state bag, and will
// initialize it.
type TestStepAcc struct {
	// The data inserted into the state bag.
	Data string
}

func (s TestStepAcc) Run(state map[string]interface{}) StepAction {
	if _, ok := state["data"]; !ok {
		state["data"] = make([]string, 0, 5)
	}

	data := state["data"].([]string)
	data = append(data, s.Data)
	state["data"] = data

	return ActionContinue
}

func (s TestStepAcc) Cleanup(map[string]interface{}) {
}
