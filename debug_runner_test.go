package multistep

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func testingPauseFn(DebugLocation, string, StateBag) {
	return
}

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

// confirm that can't run twice
func TestDebugRunner_Run_Run(t *testing.T) {
	defer func() {
		recover()
	}()
	ch := make(chan chan bool)
	stepInt := &TestStepSync{ch}
	stepWait := &TestStepWaitForever{}
	r := &DebugRunner{Steps: []Step{stepInt, stepWait}, PauseFn: testingPauseFn}

	go r.Run(new(BasicStateBag))
	// wait until really running
	<-ch

	// now try to run aain
	r.Run(new(BasicStateBag))

	// should not get here in nominal codepath
	t.Errorf("Was able to run an already running DebugRunner")
}

func TestDebugRunner_Cancel(t *testing.T) {
	ch := make(chan chan bool)
	data := new(BasicStateBag)
	stepA := &TestStepAcc{Data: "a"}
	stepB := &TestStepAcc{Data: "b"}
	stepInt := &TestStepSync{ch}
	stepC := &TestStepAcc{Data: "c"}

	r := &DebugRunner{PauseFn: testingPauseFn}
	r.Steps = []Step{stepA, stepB, stepInt, stepC}

	// cancelling an idle Runner is a no-op
	r.Cancel()

	go r.Run(data)

	// Wait until we reach the sync point
	responseCh := <-ch

	// Cancel then continue chain
	cancelCh := make(chan bool)
	go func() {
		r.Cancel()
		cancelCh <- true
	}()

	for {
		if _, ok := data.GetOk(StateCancelled); ok {
			responseCh <- true
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	<-cancelCh

	// Test run data
	expected := []string{"a", "b"}
	results := data.Get("data").([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected result: %#v", results)
	}

	// Test cleanup data
	expected = []string{"b", "a"}
	results = data.Get("cleanup").([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected result: %#v", results)
	}

	// Test that it says it is cancelled
	cancelled := data.Get(StateCancelled).(bool)
	if !cancelled {
		t.Errorf("not cancelled")
	}
}

func TestDebugPauseDefault(t *testing.T) {
	now := time.Now()
	t.Logf("[%s] start", time.Since(now))

	// Start pausing
	complete := make(chan bool, 1)
	go func() {
		t.Logf("[%s] anon func", time.Since(now))
		dr := &DebugRunner{
			Steps: []Step{
				&TestStepAcc{Data: "a"},
			},
			PauseFn: func(DebugLocation, string, StateBag) {
				t.Logf("[%s] PauseFn start", time.Since(now))
				time.Sleep(50 * time.Millisecond)
				t.Logf("[%s] PauseFn stop", time.Since(now))
			},
		}

		t.Logf("[%s] anon func start dr.Run", time.Since(now))
		dr.Run(new(BasicStateBag))
		t.Logf("[%s] anon func end dr.Run", time.Since(now))

		t.Logf("[%s] anon func blocked", time.Since(now))
		complete <- true
		t.Logf("[%s] anon func unblocked", time.Since(now))
	}()

	t.Logf("[%s] complete blocked", time.Since(now))
	select {
	case <-complete:
		t.Logf("[%s] received complete", time.Since(now))
		t.Fatal("shouldn't have completed")
	case a := <-time.After(5 * time.Millisecond):
		t.Logf("[%s] received After [%s]", time.Since(now), time.Since(a))
	}
	t.Logf("[%s] complete unblocked", time.Since(now))

	t.Logf("[%s] select final start", time.Since(now))
	select {
	case <-complete:
		t.Logf("[%s] select final complete", time.Since(now))
	case a := <-time.After(200 * time.Millisecond):
		t.Logf("[%s] select final After [%s]", time.Since(now), time.Since(a))
		t.Fatal("didn't complete")
	}
}

func ExampleDebugLocationAfterRunString() {
	fmt.Println(DebugLocationAfterRun.String())
	// Output: after run of
}

func ExampleDebugLocationBeforeCleanupString() {
	fmt.Println(DebugLocationBeforeCleanup.String())
	// Output: before cleanup of
}
