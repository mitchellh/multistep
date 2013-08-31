package multistep

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestDebugRunner_Impl(t *testing.T) {
	var raw interface{}
	raw = &DebugRunner{}
	if _, ok := raw.(Runner); !ok {
		t.Fatal("DebugRunner must be a runner.")
	}
}

func TestDebugRunner_Run(t *testing.T) {
	data := new(BasicStateBag)
	stepA := &TestStepAcc{Data: "a"}
	stepB := &TestStepAcc{Data: "b"}

	pauseFn := func(loc DebugLocation, name string, state StateBag) {
		key := "data"
		if loc == DebugLocationBeforeCleanup {
			key = "cleanup"
		}

		if _, ok := state.GetOk(key); !ok {
			state.Put(key, make([]string, 0, 5))
		}

		data := state.Get(key).([]string)
		state.Put(key, append(data, name))
	}

	r := &DebugRunner{
		Steps:   []Step{stepA, stepB},
		PauseFn: pauseFn,
	}

	r.Run(data)

	// Test data
	expected := []string{"a", "TestStepAcc", "b", "TestStepAcc"}
	results := data.Get("data").([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected results: %#v", results)
	}

	// Test cleanup
	expected = []string{"TestStepAcc", "b", "TestStepAcc", "a"}
	results = data.Get("cleanup").([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected results: %#v", results)
	}
}

func TestDebugPauseDefault(t *testing.T) {
	loc := DebugLocationAfterRun
	name := "foo"
	state := new(BasicStateBag)

	// Create a pipe pair so that writes/reads are blocked until we do it
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Set stdin so we can control it
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Start pausing
	complete := make(chan bool, 1)
	go func() {
		DebugPauseDefault(loc, name, state)
		complete <- true
	}()

	select {
	case <-complete:
		t.Fatal("shouldn't have completed")
	case <-time.After(100 * time.Millisecond):
	}

	w.Write([]byte("\n"))

	select {
	case <-complete:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("didn't complete")
	}
}
