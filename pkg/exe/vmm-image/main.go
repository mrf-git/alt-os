package main

import (
	"alt-os/api"
	"alt-os/exe"
	"regexp"
)

const EXE_USAGE = `vmm-image
---------
Virtual machine monitor image service executable.

Manages virtual machine images for the OS.
`

// VmmImageContext holds context information for vmm-image.
type VmmImageContext struct {
	*exe.ExeContext
	rootDir string // Stores the root directory of all image subdirectories.
}

// main is the entry point.
func main() {
	allowedKindRe := regexp.MustCompile(`os.machine.image.[[:word:]]`)
	allowedVersionRe := regexp.MustCompile(`v0`)
	ctxt := &VmmImageContext{}
	kindImplMap := map[string]interface{}{
		"os.machine.image.VmmImageService/v0": newVmmImageServiceServerImpl(ctxt),
	}
	respHandlerMap := map[string]func(interface{}) error{}
	ctxt.ExeContext = exe.InitContext(EXE_USAGE, allowedKindRe, allowedVersionRe,
		kindImplMap, respHandlerMap)
	if err := api.ServiceMessages(ctxt.ApiServiceContext); err != nil {
		exe.Fatal("servicing messages", err, ctxt.ExeContext)
	}
	exe.Success(ctxt.ExeContext)
}
