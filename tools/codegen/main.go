package main

import (
	"alt-os/exe"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const _USAGE = `codegen
-------
Installs any necessary module dependencies and runs code
generation on the source tree.
`

// CodegenContext holds context information for codegen.
type CodegenContext struct {
	*exe.ExeContext
	SrcRootDir string
}

// main is the entry point.
func main() {
	// Parse command line.
	var pb bool
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", _USAGE)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
	}
	flag.BoolVar(&pb, "p", false, "Generate the protocol buffer code")
	flag.Parse()

	if !pb {
		flag.Usage()
		os.Exit(1)
	}

	// Initialize the context and get all module dependencies.
	ctxt := initContext()
	if stdOut, stdErr, err := exe.Doexec("", "go", "mod", "download", "all"); err != nil {
		exe.Fatal("getting modules", exe.ErrOutput(stdOut, stdErr, err), ctxt.ExeContext)
	}

	if pb {
		protogen(ctxt)
	}

	exe.Success(ctxt.ExeContext)
}

// initContext initializes a new codegen context and quickly checks that the current working
// directory is the source root.
func initContext() *CodegenContext {
	ctxt := &CodegenContext{ExeContext: &exe.ExeContext{}}

	wkToolsDir, _ := os.Executable()
	wkToolsDir = filepath.Dir(filepath.Clean(wkToolsDir))
	srcRootDir := filepath.Clean(filepath.Join(wkToolsDir, "..", ".."))

	if cwd, err := os.Getwd(); err != nil {
		exe.Fatal("getting working dir", err, ctxt.ExeContext)
	} else if filepath.Clean(cwd) != srcRootDir || srcRootDir == "" ||
		filepath.Clean(filepath.Join(srcRootDir, "workspace", "tools")) != wkToolsDir {

		exe.Fatal("getting working dir", errors.New("bad working directory or root dir"), ctxt.ExeContext)
	}

	ctxt.SrcRootDir = srcRootDir

	return ctxt
}
