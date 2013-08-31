package multistep

import "sync"

// BasicRunner is a Runner that just runs the given slice of steps.
type BasicRunner struct {
	// Steps is a slice of steps to run. Once set, this should _not_ be
	// modified.
	Steps []Step

	cancelCond *sync.Cond
	cancelChs  []chan<- bool
	running    bool
	l          sync.Mutex
}

func (b *BasicRunner) Run(state StateBag) {
	// Make sure we only run one at a time
	b.l.Lock()
	if b.running {
		panic("already running")
	}
	b.cancelChs = nil
	b.cancelCond = sync.NewCond(&sync.Mutex{})
	b.running = true
	b.l.Unlock()

	// cancelReady is used to signal that the cancellation goroutine
	// started and is waiting. The cancelEnded channel is used to
	// signal the goroutine actually ended.
	cancelReady := make(chan bool, 1)
	cancelEnded := make(chan bool)
	go func() {
		b.cancelCond.L.Lock()
		cancelReady <- true
		b.cancelCond.Wait()
		b.cancelCond.L.Unlock()

		if b.cancelChs != nil {
			state.Put(StateCancelled, true)
		}

		cancelEnded <- true
	}()

	// Create the channel that we'll say we're done on in the case of
	// interrupts here. We do this here so that this deferred statement
	// runs last, so all the Cleanup methods are able to run.
	defer func() {
		b.l.Lock()
		defer b.l.Unlock()

		// Make sure the cancellation goroutine cleans up properly. This
		// is a bit complicated. Basically, we first wait until the goroutine
		// waiting for cancellation is actually waiting. Then we broadcast
		// to it so it can unlock. Then we wait for it to tell us it finished.
		<-cancelReady
		b.cancelCond.L.Lock()
		b.cancelCond.Broadcast()
		b.cancelCond.L.Unlock()
		<-cancelEnded

		if b.cancelChs != nil {
			for _, doneCh := range b.cancelChs {
				doneCh <- true
			}
		}

		b.running = false
	}()

	for _, step := range b.Steps {
		// We also check for cancellation here since we can't be sure
		// the goroutine that is running to set it actually ran.
		if b.cancelChs != nil {
			state.Put(StateCancelled, true)
			break
		}

		action := step.Run(state)
		defer step.Cleanup(state)

		if _, ok := state.GetOk(StateCancelled); ok {
			break
		}

		if action == ActionHalt {
			state.Put(StateHalted, true)
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
	b.cancelCond.Broadcast()
	b.l.Unlock()

	<-done
}
