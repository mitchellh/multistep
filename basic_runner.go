package multistep

import "sync"

// BasicRunner is a Runner that just runs the given slice of steps.
type BasicRunner struct {
	// Steps is a slice of steps to run. Once set, this should _not_ be
	// modified.
	Steps []Step

	cancelChs []chan<- bool
	running   bool
	l         sync.Mutex
}

func (b *BasicRunner) Run(state map[string]interface{}) {
	// Make sure we only run one at a time
	b.l.Lock()
	if b.running {
		panic("already running")
	}
	b.cancelChs = nil
	b.running = true
	b.l.Unlock()

	// Create the channel that we'll say we're done on in the case of
	// interrupts here. We do this here so that this deferred statement
	// runs last, so all the Cleanup methods are able to run.
	defer func() {
		b.l.Lock()
		defer b.l.Unlock()

		for _, doneCh := range b.cancelChs {
			doneCh <- true
		}

		b.running = false
	}()

	for _, step := range b.Steps {
		// If we got a cancel notification, then set the done channel
		// and just exit the loop now.
		if b.cancelChs != nil {
			state["cancelled"] = true
			break
		}

		action := step.Run(state)
		defer step.Cleanup(state)

		if action == ActionHalt {
			break
		}
	}
}

func (b *BasicRunner) Cancel() {
	b.l.Lock()

	if !b.running {
		b.l.Unlock()
		return
	}

	if b.cancelChs == nil {
		b.cancelChs = make([]chan<- bool, 0, 5)
	}

	done := make(chan bool)
	b.cancelChs = append(b.cancelChs, done)
	b.l.Unlock()

	<-done
}
