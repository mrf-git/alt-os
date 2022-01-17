package main

import "alt-os/exe"

// VmmRuntimeContext holds context information for vmm-runtime.
type VmmRuntimeContext struct {
	*exe.ExeContext
	imageDir    string                   // Stores the root directory of all image subdirectories.
	maxMachines int                      // The maximum number of virtual machines to allow at once.
	vmEnvs      map[string]VmEnvironment // Maps VM id strings to their environment.
	vmChs       map[string]chan<- int    // Maps VM ids to their return code channels.
}
