package main

import (
	"alt-os/os/code"
	"fmt"
)

// VmEnvironment encapsulates the runtime environment for a single
// virtual machine to run virtualized code.
type VmEnvironment interface {
	// Run creates a hardware-virtualized, privilege level 0
	// environment in a new goroutine and starts running the virtual
	// machine code in it. Sends the code returned by main to
	// returnCodeCh when the virtual machine exits.
	Run(returnCodeCh <-chan int) error
}

// newVmEnvironment returns a newly-instantiated VmEnvironment for the
// specified code to run in.
func newVmEnvironment(exeCode code.ExecutableCode, ctxt *VmmRuntimeContext) VmEnvironment {
	return &_VmEnvironment{
		ctxt:    ctxt,
		exeCode: exeCode,
	}
}

type _VmEnvironment struct {
	ctxt    *VmmRuntimeContext
	exeCode code.ExecutableCode
}

func (vmEnv *_VmEnvironment) Run(returnCodeCh <-chan int) error {

	fmt.Println("starting run")
	// TODO
	return nil
}
