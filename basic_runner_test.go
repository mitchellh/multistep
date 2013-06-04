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
}

func TestBasicRunner_Cancel(t *testing.T) {
	ch := make(chan chan bool)
	data := make(map[string]interface{})
	stepA := &TestStepAcc{Data: "a"}
	stepB := &TestStepAcc{Data: "b"}
	stepInt := &TestStepSync{ch}
	stepC := &TestStepAcc{Data: "c"}

	r := &BasicRunner{Steps: []Step{stepA, stepB, stepInt, stepC}}
	go r.Run(data)

	// Wait until we reach the sync point
	responseCh := <-ch

	// Cancel then continue chain
	cancelCh := make(chan bool)
	go func() {
		r.Cancel()
		cancelCh <- true
	}()

	responseCh <- true
	<-cancelCh

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
}
