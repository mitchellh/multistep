package multistep

// BasicRunner is a Runner that just runs the given slice of steps.
type BasicRunner struct {
	// Steps is a slice of steps to run. Once set, this should _not_ be
	// modified.
	Steps []Step
}

func (b *BasicRunner) Run(state map[string]interface{}) {
	for _, step := range b.Steps {
		action := step.Run(state)
		defer step.Cleanup(state)

		if action == ActionHalt {
			break
		}
	}
}

func (b *BasicRunner) Cancel() {
}
