package exe

import (
	"fmt"
	"os"
)

// Fatal prints a fatal status error message to stderr and exits with error status.
func Fatal(status string, err error, ctxt *ExeContext) {
	fmt.Fprintf(os.Stderr, "[ERROR] +++ %s +++ : %s\n", status, err.Error())
	// Cleanup after failure.
	for _, fn := range ctxt.CleanupFuncs {
		fn()
	}
	os.Exit(1)
}

// Success exits with success status.
func Success(ctxt *ExeContext) {
	// Cleanup after success.
	for _, fn := range ctxt.CleanupFuncs {
		fn()
	}
	os.Exit(0)
}
