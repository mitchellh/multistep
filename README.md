# multistep

multistep is a Go library for building up complex actions using discrete,
individual "steps." These steps are strung together and run in sequence
to achieve a more complex goal. The runner handles cleanup, cancelling, etc.
if necessary.

## Basic Example

Make a step to perform some action. The step can access your "state",
which is passed between steps by the runner.

```go
type stepAdd struct{}

func (s *stepAdd) Run(state map[string]interface{}) multistep.StepAction {
    // Read our value and assert that it is they type we want
    value := state["value"].(int)

    fmt.Printf("Value is %d\n", value)

    state["value"] = value + 1

    return multistep.ActionContinue
}

func (s *stepAdd) Cleanup(map[string]interface{}) {}
```

Make a runner and call your array of Steps.

```go
func main() {
    // Our "bag of state" that we read the value from
    state := make(map[string]interface{})

    // Our intial value
    state["value"] = 0

    steps := []multistep.Step{
        &stepAdd{},
        &stepAdd{},
        &stepAdd{},
    }

    runner := &multistep.BasicRunner{Steps: steps}

    // Executes the steps
    runner.Run(state)
}
```

This will produce:

```
Value is 1
Value is 2
Value is 3
```
