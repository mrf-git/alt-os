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

	if homedir, err := os.UserHomeDir(); err != nil {
		fmt.Fprintf(os.Stderr, "could not get homedir: %s\n", err.Error())
		os.Exit(1)
	} else {
		defaultCachedir = path.Join(homedir, ".cache", "os-build-cache")
	}

	// Parse command line.
	var profilename, confname, cachedir string
	var wipesrc, wipeall, verbose bool
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", EXE_USAGE)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
	}
	flag.StringVar(&profilename, "profile", defaultProfile, "The build profile to use")
	flag.StringVar(&cachedir, "cache", defaultCachedir, "The directory to use for caching")
	flag.StringVar(&confname, "i", "", "The build configuration file to use")
	flag.BoolVar(&wipesrc, "ws", false, "Wipes the cached OS build files under the cache directory before building")
	flag.BoolVar(&wipeall, "wa", false, "Wipes all files under the cache directory before building (implies -ws)")
	flag.BoolVar(&wipesrc, "w", false, "Alias for -ws")
	flag.BoolVar(&verbose, "v", false, "Whether to print additional build details")
	flag.Parse()
	if confname == "" {
		flag.Usage()
		os.Exit(1)
	}
	if wipeall {
		wipesrc = true
	}
	if strings.ContainsAny(profilename, " \n\r\t") {
		fmt.Fprintf(os.Stderr, "spaces in profile name\n")
		os.Exit(1)
	}

	// Initialize context and wipe directories if needed.
	ctxt := initContext(profilename, confname, cachedir, verbose)
	if wipeall {
		os.RemoveAll(cachedir)
	} else if wipesrc {
		os.RemoveAll(path.Join(cachedir, "src"))
	}

	osbuild(ctxt)

	exe.Success(ctxt.ExeContext)
}
