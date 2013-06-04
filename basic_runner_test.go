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

	r := &BasicRunner{[]Step{stepA, stepB}}
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

	r := &BasicRunner{[]Step{stepA, stepB, stepC}}
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
