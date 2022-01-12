package main

import (
	"alt-os/api"
	"alt-os/exe"
	"regexp"
)

const EXE_USAGE = `vmm-runtime
-----------
Virtual machine monitor runtime service executable.

Manages virtual machine runtimes for the OS.
`

// VmmRuntimeContext holds context information for vmm-runtime.
type VmmRuntimeContext struct {
	*exe.ExeContext
	imageDir    string                   // Stores the root directory of all image subdirectories.
	maxMachines int                      // The maximum number of virtual machines to allow at once.
	vmEnvs      map[string]VmEnvironment // Maps VM id strings to their environment.
}

// main is the entry point.
func main() {
	allowedKindRe := regexp.MustCompile(`os.machine.runtime.[[:word:]]`)
	allowedVersionRe := regexp.MustCompile(`v0`)
	ctxt := &VmmRuntimeContext{
		vmEnvs: make(map[string]VmEnvironment),
	}
	kindImplMap := map[string]interface{}{
		"os.machine.runtime.VmmRuntimeService/v0": newVmmRuntimeServiceServerImpl(ctxt),
	}
	respHandlerMap := map[string]func(interface{}) error{}
	ctxt.ExeContext = exe.InitContext(EXE_USAGE, allowedKindRe, allowedVersionRe,
		kindImplMap, respHandlerMap)
	if err := api.ServiceMessages(ctxt.ApiServiceContext); err != nil {
		exe.Fatal("servicing messages", err, ctxt.ExeContext)
	}
	exe.Success(ctxt.ExeContext)
}
