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
}

// main is the entry point.
func main() {
	allowedKindRe := regexp.MustCompile(`os.machine.runtime.[[:word:]]`)
	allowedVersionRe := regexp.MustCompile(`v0`)
	ctxt := &VmmRuntimeContext{}
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
