package exe

// ExeContext stores information about the currently-running executable.
type ExeContext struct {
	CleanupFuncs []func() // Functions that must run at exit.
}
