package main

import "alt-os/os/code"

// VmEnvironment encapsulates the runtime environment for a single
// virtual machine to run virtualized code.
type VmEnvironment interface {
	// RunExecutableCode creates a hardware-virtualized, privilege level 0
	// environment in a new goroutine and starts running the specified
	// executable code in it. Calls returnCodeCallback with the code
	// returned by main when the code exits.
	RunExecutableCode(exeCode code.ExecutableCode, returnCodeCallback func(int)) error
}

// newVmEnvironment returns a newly-instantiated VmEnvironment.
func newVmEnvironment(ctxt *VmmRuntimeContext) VmEnvironment {
	return &_VmEnvironment{
		ctxt: ctxt,
	}
}

type _VmEnvironment struct {
	ctxt *VmmRuntimeContext
}

func (vmEnv *_VmEnvironment) RunExecutableCode(exeCode code.ExecutableCode,
	returnCodeCallback func(int)) error {

	// TODO
	return nil
}
