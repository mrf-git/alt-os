package main

import (
	"alt-os/api"
	"alt-os/exe"
	"regexp"
)

const EXE_USAGE = `vm-image
--------
Virtual machine monitor image service executable.

Manages virtual machine images for the OS.
`

// VmImageContext holds context information for vm-image.
type VmImageContext struct {
	*exe.ExeContext
	rootDir string // Stores the root directory of all image subdirectories.
}

// main is the entry point.
func main() {
	allowedKindRe := regexp.MustCompile(`os.machine.image.[[:word:]]`)
	allowedVersionRe := regexp.MustCompile(`v0`)
	ctxt := &VmImageContext{}
	kindImplMap := map[string]interface{}{
		"os.machine.image.VmImageService/v0": newVmImageServiceServerImpl(ctxt),
	}
	respHandlerMap := map[string]func(interface{}) error{}
	loggerConf := &exe.LoggerConf{
		Enabled:    true,
		Level:      "info",
		ExeTag:     "vm-image",
		FormatJson: false,
	}
	ctxt.ExeContext = exe.InitContext(EXE_USAGE, allowedKindRe, allowedVersionRe,
		kindImplMap, respHandlerMap, loggerConf)
	if err := api.ServiceMessages(ctxt.ApiServiceContext); err != nil {
		exe.Fatal("servicing messages", err, ctxt.ExeContext)
	}
	exe.Success(ctxt.ExeContext)
}
