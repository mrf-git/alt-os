package main

import (
	"alt-os/api"
	api_os_container_runtime_v0 "alt-os/api/os/container/runtime/v0"
	"alt-os/exe"
	"regexp"
)

const EXE_USAGE = `ct-runtime
----------
Container runtime service executable.

Operates a container runtime in accordance with the OCI runtime specification:
https://github.com/opencontainers/runtime-spec/blob/v1.0.2/runtime.md
`

// CtRuntimeContext holds context information for ct-runtime.
type CtRuntimeContext struct {
	*exe.ExeContext
}

// main is the entry point.
func main() {
	allowedKindRe := regexp.MustCompile(`os.container.runtime.[[:word:]]`)
	allowedVersionRe := regexp.MustCompile(`v0`)
	ctxt := &CtRuntimeContext{}
	kindImplMap := map[string]interface{}{
		"os.container.runtime.ContainerRuntimeService/v0": newContainerRuntimeServiceServerImpl(ctxt),
	}
	respHandlerMap := map[string]func(interface{}) error{
		"os.container.runtime.ContainerRuntimeService/v0.List": func(resp interface{}) error {
			return handleRespList(resp.(*api_os_container_runtime_v0.ListResponse))
		},
	}
	loggerConf := &exe.LoggerConf{
		Enabled:    true,
		Level:      "info",
		ExeTag:     "ct-runtime",
		FormatJson: false,
	}
	ctxt.ExeContext = exe.InitContext(EXE_USAGE, allowedKindRe, allowedVersionRe,
		kindImplMap, respHandlerMap, loggerConf)
	if err := api.ServiceMessages(ctxt.ApiServiceContext); err != nil {
		exe.Fatal("servicing messages", err, ctxt.ExeContext)
	}
	exe.Success(ctxt.ExeContext)
}
