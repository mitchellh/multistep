package multistep

import (
	"reflect"
	"testing"
)

func TestBasicRunner_ImplRunner(t *testing.T) {
	var raw interface{}
	raw = &BasicRunner{}
	if _, ok := raw.(Runner); !ok {
		t.Fatalf("BasicRunner must be a Runner")
	}
}

func TestBasicRunner_Run(t *testing.T) {
	data := make(map[string]interface{})
	stepA := &TestStepAcc{Data: "a"}
	stepB := &TestStepAcc{Data: "b"}

	r := &BasicRunner{Steps: []Step{stepA, stepB}}
	r.Run(data)

	// Test run data
	expected := []string{"a", "b"}
	results := data["data"].([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected result: %#v", results)
	}

	// Test cleanup data
	expected = []string{"b", "a"}
	results = data["cleanup"].([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected result: %#v", results)
	}

	// Test no halted or cancelled
	if _, ok := data[StateCancelled]; ok {
		t.Errorf("cancelled should not be in state bag")
	}

	if _, ok := data[StateHalted]; ok {
		t.Errorf("halted should not be in state bag")
	}
}

func TestBasicRunner_Run_Halt(t *testing.T) {
	data := make(map[string]interface{})
	stepA := &TestStepAcc{Data: "a"}
	stepB := &TestStepAcc{Data: "b", Halt: true}
	stepC := &TestStepAcc{Data: "c"}

	r := &BasicRunner{Steps: []Step{stepA, stepB, stepC}}
	r.Run(data)

	// Test run data
	expected := []string{"a", "b"}
	results := data["data"].([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected result: %#v", results)
	}

	// Test cleanup data
	expected = []string{"b", "a"}
	results = data["cleanup"].([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected result: %#v", results)
	}

	// Test that it says it is halted
	halted := data[StateHalted].(bool)
	if !halted {
		t.Errorf("not halted")
	}
}

func TestBasicRunner_Cancel(t *testing.T) {
	data := make(map[string]interface{})
	stepA := &TestStepAcc{Data: "a"}
	stepB := &TestStepAcc{Data: "b"}
	sync := make(chan struct{})
	stepInt := &TestStepSync{C: sync}
	stepC := &TestStepAcc{Data: "c"}

	r := &BasicRunner{Steps: []Step{stepA, stepB, stepInt, stepC}}
	go r.Run(data)

	sync <- struct{}{} // continue stepInt
	r.Cancel()

	// Test run data
	expected := []string{"a", "b"}
	results := data["data"].([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected result: %#v", results)
	}

	// Test cleanup data
	expected = []string{"b", "a"}
	results = data["cleanup"].([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected result: %#v", results)
	}

	// Test that the sync cleanup had the cancelled flag
	if _, ok := data["sync_cancelled"]; !ok {
		t.Errorf("sync cleanup not cancelled")
	}

	// Test that it says it was cancelled
	if _, ok := data[StateCancelled]; !ok {
		t.Errorf("not cancelled")
	}
}
