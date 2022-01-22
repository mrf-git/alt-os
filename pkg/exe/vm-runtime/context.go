package main

import (
	"alt-os/exe"
)

// VmRuntimeContext holds context information for vm-runtime.
type VmRuntimeContext struct {
	*exe.ExeContext
	// Stores the root directory of all image subdirectories.
	imageDir string
	// The maximum number of virtual machines to allow at once.
	maxMachines int
	// Maps VM id strings to their environment.
	vmEnvs map[string]VmEnvironment
	// Maps VM ids to their kill signal channels.
	vmSigChs map[string]chan<- int
	// Maps VM ids to their return code channels.
	vmRetChs map[string]<-chan int
}
