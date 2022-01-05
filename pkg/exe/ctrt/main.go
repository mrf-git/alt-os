package main

import (
	"alt-os/api"
	"alt-os/exe"
	"flag"
	"fmt"
	"os"
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
	Messages []*api.ApiProtoMessage
}

// main is the entry point.
func main() {
	// Parse command line.
	var infile, format string
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", _USAGE)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
	}
	flag.StringVar(&infile, "i", "", "The input object(s) file to use for container runtime changes")
	flag.StringVar(&format, "f", "", "Input file format (yaml,json), inferred from extension if not specified")
	flag.Parse()

	if infile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Read input file and initialize context.
	ctxt := &CtrtContext{ExeContext: &exe.ExeContext{}}
	if messages, err := api.UnmarshalApiProtoMessages(infile, format); err != nil {
		exe.Fatal("unmarshaling proto messages", err, ctxt.ExeContext)
	} else {
		ctxt.Messages = messages
	}

	handleApiMessages(ctxt)

	exe.Success(ctxt.ExeContext)
}
