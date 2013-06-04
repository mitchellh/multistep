# multistep

multistep is a Go library for building up complex actions using discrete,
individual "steps." These steps are strung together and run in sequence
to achieve a more complex goal. The runner handles cleanup, cancelling, etc.
if necessary.
