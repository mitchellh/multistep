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
	stepA := &TestStepAcc{"a"}
	stepB := &TestStepAcc{"b"}

	r := &BasicRunner{[]Step{stepA, stepB}}
	r.Run(data)

	expected := []string{"a", "b"}
	results := data["data"].([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected result: %#v", results)
	}
}
