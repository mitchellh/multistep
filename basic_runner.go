package multistep

import "sync"

// BasicRunner is a Runner that just runs the given slice of steps.
type BasicRunner struct {
	// Steps is a slice of steps to run. Once set, this should _not_ be
	// modified.
	Steps []Step

	cancelCh chan chan bool
	running  bool
	l        sync.Mutex
}

func (b *BasicRunner) Run(state map[string]interface{}) {
	// Make sure we only run one at a time
	b.l.Lock()
	if b.running {
		panic("already running")
	}
	b.running = true
	b.l.Unlock()

	// Create the channel that we'll say we're done on in the case of
	// interrupts here. We do this here so that this deferred statement
	// runs last, so all the Cleanup methods are able to run.
	var doneCh chan bool
	defer func() {
		if doneCh != nil {
			doneCh <- true
		}
	}()

StepLoop:
	for _, step := range b.Steps {
		// If we got a cancel notification, then set the done channel
		// and just exit the loop now.
		select {
		case doneCh = <-b.cancelCh:
			state["cancelled"] = true
			break StepLoop
		default:
		}

		action := step.Run(state)
		defer step.Cleanup(state)

		if action == ActionHalt {
			break
		}
	}

	b.running = false
}

func (b *BasicRunner) Cancel() {
	b.l.Lock()
	defer b.l.Unlock()

	if !b.running {
		return
	}

	done := make(chan bool)
	b.cancelCh = make(chan chan bool)
	b.cancelCh <- done
	<-done
}
