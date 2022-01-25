package main

import (
	"alt-os/exe"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
)

const EXE_USAGE = `osbuild
-------
Installs any necessary development dependencies and builds
operating system artifacts from source code. Requires a
build configuration file as input.
`

// main is the entry point.
func main() {

	// Default argument setup.
	defaultProfile := "dev"
	var defaultCachedir string
	var defaultLlvmdir string

	if homedir, err := os.UserHomeDir(); err != nil {
		fmt.Fprintf(os.Stderr, "could not get homedir: %s\n", err.Error())
		os.Exit(1)
	} else {
		defaultCachedir = path.Join(homedir, ".cache", "os-build-cache")
		defaultLlvmdir = path.Join(homedir, "llvm-alt")
	}

	// Parse command line.
	var profilename, confname, cachedir, llvmdir string
	var wipe, wipesrc, wipeall, verbose bool
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", EXE_USAGE)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
	}
	flag.StringVar(&profilename, "profile", defaultProfile, "The build profile to use")
	flag.StringVar(&cachedir, "cache", defaultCachedir, "The directory to use for caching")
	flag.StringVar(&llvmdir, "llvm", defaultLlvmdir, "The path to the root of llvm-alt")
	flag.StringVar(&confname, "i", "", "The build configuration file to use")
	flag.BoolVar(&wipesrc, "ws", false, "Wipes the cached OS build files under the cache directory before building")
	flag.BoolVar(&wipeall, "wa", false, "Wipes all files under the cache directory before building (implies -ws)")
	flag.BoolVar(&wipe, "w", false, "Alias for -ws")
	flag.BoolVar(&verbose, "v", false, "Whether to print additional build details")
	flag.Parse()
	if confname == "" {
		flag.Usage()
		os.Exit(1)
	}
	if wipeall || wipe {
		wipesrc = true
	}
	if strings.ContainsAny(profilename, " \n\r\t") {
		fmt.Fprintf(os.Stderr, "spaces in profile name\n")
		os.Exit(1)
	}

	// Initialize context and wipe directories if needed.
	loggerConf := &exe.LoggerConf{
		Enabled:    true,
		Level:      "info",
		ExeTag:     "osbuild",
		FormatJson: false,
	}
	ctxt := initContext(profilename, confname, cachedir, llvmdir, verbose, loggerConf)
	if wipeall {
		os.RemoveAll(cachedir)
	} else if wipesrc {
		ctxt.WipeSource = true
	}

	os.Setenv("PATH", ctxt.LlvmDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	os.Setenv("CC", "clang")
	os.Setenv("CXX", "clang++")

	osbuild(ctxt)

	exe.Success(ctxt.ExeContext)
}
