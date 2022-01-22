package main

import (
	"alt-os/api"
	"alt-os/exe"
	"regexp"
)

const EXE_USAGE = `vm-runtime
----------
Virtual machine monitor runtime service executable.

Manages virtual machine runtimes for the OS.
`

// main is the entry point.
func main() {
	allowedKindRe := regexp.MustCompile(`os.machine.runtime.[[:word:]]`)
	allowedVersionRe := regexp.MustCompile(`v0`)
	ctxt := &VmRuntimeContext{
		vmEnvs:   make(map[string]VmEnvironment),
		vmSigChs: make(map[string]chan<- int),
		vmRetChs: make(map[string]<-chan int),
	}
	kindImplMap := map[string]interface{}{
		"os.machine.runtime.VmRuntimeService/v0": newVmRuntimeServiceServerImpl(ctxt),
	}
	respHandlerMap := map[string]func(interface{}) error{}
	loggerConf := &exe.LoggerConf{
		Enabled:    true,
		Level:      "info",
		ExeTag:     "vm-runtime",
		FormatJson: false,
	}
	ctxt.ExeContext = exe.InitContext(EXE_USAGE, allowedKindRe, allowedVersionRe,
		kindImplMap, respHandlerMap, loggerConf)
	if err := api.ServiceMessages(ctxt.ApiServiceContext); err != nil {
		exe.Fatal("servicing messages", err, ctxt.ExeContext)
	}
	exe.Success(ctxt.ExeContext)
}
