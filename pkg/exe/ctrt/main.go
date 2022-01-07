package main

import (
	"alt-os/api"
	api_ctrt_v0 "alt-os/api/ctrt/v0"
	"alt-os/exe"
	"regexp"
)

const _USAGE = `ctrt
----
Container runtime.
Operates a container runtime in accordance with the OCI runtime
specification:
https://github.com/opencontainers/runtime-spec/blob/v1.0.2/runtime.md

The program expects a single mandatory parameter specifying one or
more objects in yaml format to apply to a runtime.
`

// CtrtContext holds context information for ctrt.
type CtrtContext struct {
	*exe.ExeContext
}

// main is the entry point.
func main() {
	allowedKindRe := regexp.MustCompile(`api.ctrt.[[:word:]]`)
	allowedVersionRe := regexp.MustCompile(`v0`)
	ctxt := &CtrtContext{}
	kindImplMap := map[string]interface{}{
		"api.ctrt.ContainerRuntime/v0": NewContainerRuntimeServerImpl(ctxt),
	}
	respHandlerMap := map[string]func(interface{}) error{
		"api.ctrt.ContainerRuntime/v0.List": func(resp interface{}) error {
			return handleRespList(resp.(*api_ctrt_v0.ListResponse))
		},
	}
	ctxt.ExeContext = exe.InitContext(_USAGE, allowedKindRe, allowedVersionRe,
		kindImplMap, respHandlerMap)
	if err := api.ServiceMessages(ctxt.ApiServiceContext); err != nil {
		exe.Fatal("servicing messages", err, ctxt.ExeContext)
	}
	exe.Success(ctxt.ExeContext)
}
