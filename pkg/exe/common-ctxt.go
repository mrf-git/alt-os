package exe

import "alt-os/api"

// ExeContext stores information about the currently-running executable.
type ExeContext struct {
	*api.ApiServiceContext
	CleanupFuncs []func() // Functions that must run at exit.
}
