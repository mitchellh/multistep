package multistep

import "sync"

// BasicRunner is a Runner that just runs the given slice of steps.
type BasicRunner struct {
	// Steps is a slice of steps to run. Once set, this should _not_ be
	// modified.
	Steps []Step

	cancelDone chan struct{}
	runState   runState
	// l protects runState
	l sync.Mutex
}

type runState int

const (
	stateInitial = iota
	stateRunning
	stateCancelling
)

func (b *BasicRunner) Run(state map[string]interface{}) {
	b.l.Lock()
	// Make sure we only run one instance at a time
	if b.runState != stateInitial {
		panic("already running")
	}
	b.cancelDone = make(chan struct{})
	b.runState = stateRunning
	b.l.Unlock()

	// This runs after all of the cleanup steps so that we can notify any
	// waiting Cancel callers and transition the state back to initial
	defer func() {
		b.l.Lock()
		b.runState = stateInitial
		b.l.Unlock()
		close(b.cancelDone)
	}()

	for _, step := range b.Steps {
		b.l.Lock()
		if b.runState != stateRunning {
			// We've been cancelled, update the state bag and abort
			b.l.Unlock()
			state[StateCancelled] = true
			break
		}
		b.l.Unlock()

		action := step.Run(state)
		defer step.Cleanup(state)

		if action == ActionHalt {
			state[StateHalted] = true
			break
		}
	}
}

func (b *BasicRunner) Cancel() {
	b.l.Lock()
	// No-op if we're not running
	if b.runState == stateInitial {
		b.l.Unlock()
		return
	}
	// Transition state from running to cancelling
	b.runState = stateCancelling
	b.l.Unlock()

	// Wait until all of the cleanup hooks have been run
	<-b.cancelDone
}
