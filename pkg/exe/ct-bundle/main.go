package main

import (
	"alt-os/api"
	"alt-os/exe"
	"regexp"
)

const EXE_USAGE = `ct-bundle
---------
Container bundle service executable.

Manages container bundles in accordance with the OCI bundle specification:
https://github.com/opencontainers/runtime-spec/blob/v1.0.2/bundle.md.
`

// CtBundleContext holds context information for ct-bundle.
type CtBundleContext struct {
	*exe.ExeContext
	rootDir string // Stores the root directory of all bundle subdirectories.
}

// main is the entry point.
func main() {
	allowedKindRe := regexp.MustCompile(`os.container.bundle.[[:word:]]`)
	allowedVersionRe := regexp.MustCompile(`v0`)
	ctxt := &CtBundleContext{}
	kindImplMap := map[string]interface{}{
		"os.container.bundle.ContainerBundleService/v0": newContainerBundleServiceServerImpl(ctxt),
	}
	respHandlerMap := map[string]func(interface{}) error{}
	loggerConf := &exe.LoggerConf{
		Enabled:    true,
		Level:      "info",
		ExeTag:     "ct-bundle",
		FormatJson: false,
	}
	ctxt.ExeContext = exe.InitContext(EXE_USAGE, allowedKindRe, allowedVersionRe,
		kindImplMap, respHandlerMap, loggerConf)
	if err := api.ServiceMessages(ctxt.ApiServiceContext); err != nil {
		exe.Fatal("servicing messages", err, ctxt.ExeContext)
	}
	exe.Success(ctxt.ExeContext)
}
