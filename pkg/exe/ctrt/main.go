package main

import (
	"alt-os/api"
	api_os_container_runtime_v0 "alt-os/api/os/container/runtime/v0"
	"alt-os/exe"
	"regexp"
)

const EXE_USAGE = `ctrt
----
Container runtime.
Operates a container runtime in accordance with the OCI runtime specification:
https://github.com/opencontainers/runtime-spec/blob/v1.0.2/runtime.md

The program expects a mandatory input file containing one or more api
requests to apply to a runtime.
`

// CtrtContext holds context information for ctrt.
type CtrtContext struct {
	*exe.ExeContext
}

// main is the entry point.
func main() {
	allowedKindRe := regexp.MustCompile(`os.container.runtime.[[:word:]]`)
	allowedVersionRe := regexp.MustCompile(`v0`)
	ctxt := &CtrtContext{}
	kindImplMap := map[string]interface{}{
		"os.container.runtime.ContainerRuntimeService/v0": newContainerRuntimeServiceServerImpl(ctxt),
	}
	respHandlerMap := map[string]func(interface{}) error{
		"os.container.runtime.ContainerRuntimeService/v0.List": func(resp interface{}) error {
			return handleRespList(resp.(*api_os_container_runtime_v0.ListResponse))
		},
	}
	ctxt.ExeContext = exe.InitContext(EXE_USAGE, allowedKindRe, allowedVersionRe,
		kindImplMap, respHandlerMap)
	if err := api.ServiceMessages(ctxt.ApiServiceContext); err != nil {
		exe.Fatal("servicing messages", err, ctxt.ExeContext)
	}
	exe.Success(ctxt.ExeContext)
}
