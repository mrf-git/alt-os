package main

import "alt-os/exe"

// HwRuntimeContext holds context information for hw-runtime.
type HwRuntimeContext struct {
	*exe.ExeContext
	// Stores the root directory of all image subdirectories on the device.
	imageDir string
	// The maximum number of virtual machines to allow at once.
	maxMachines int
	// Maps hardware id strings to their environment.
	hwEnvs map[string]HwEnvironment
	// Maps VM ids to their kill signal channels.
	vmSigChs map[string]chan<- int
	// Maps VM ids to their return code channels.
	vmRetChs map[string]<-chan int
}
