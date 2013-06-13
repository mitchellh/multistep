package multistep

import (
	"reflect"
	"testing"
)

func TestDebugRunner_Impl(t *testing.T) {
	var raw interface{}
	raw = &DebugRunner{}
	if _, ok := raw.(Runner); !ok {
		t.Fatal("DebugRunner must be a runner.")
	}
}

func TestDebugRunner_Run(t *testing.T) {
	data := make(map[string]interface{})
	stepA := &TestStepAcc{Data: "a"}
	stepB := &TestStepAcc{Data: "b"}

	pauseFn := func(loc DebugLocation, name string, state map[string]interface{}) {
		key := "data"
		if loc == DebugLocationBeforeCleanup {
			key = "cleanup"
		}

		if _, ok := state[key]; !ok {
			state[key] = make([]string, 0, 5)
		}

		data := state[key].([]string)
		state[key] = append(data, name)
	}

	r := &DebugRunner{
		Steps:   []Step{stepA, stepB},
		PauseFn: pauseFn,
	}

	r.Run(data)

	// Test data
	expected := []string{"a", "TestStepAcc", "b", "TestStepAcc"}
	results := data["data"].([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected results: %#v", results)
	}

	// Test cleanup
	expected = []string{"TestStepAcc", "b", "TestStepAcc", "a"}
	results = data["cleanup"].([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected results: %#v", results)
	}
}
