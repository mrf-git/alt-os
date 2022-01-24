package main

import (
	"alt-os/api"
	api_os_build_v0 "alt-os/api/os/build/v0"
	"alt-os/exe"
	"errors"
	"os"
	"path/filepath"
)

// OsbuildContext holds context information for osbuild.
type OsbuildContext struct {
	*exe.ExeContext
	SrcRootDir string
	CacheDir   string
	Outpath    string
	IsVerbose  bool
	Conf       *api_os_build_v0.BuildConfiguration
	Profile    *api_os_build_v0.BuildProfile
	BuildInfo  *api_os_build_v0.BuildInfo
	ArchEFI    string

	FlagsCCRuntime    []string // All clang cc flags for sys runtime.
	FlagsCCRes        []string // All clang cc flags for boot resources.
	FlagsCCBoot       []string // All clang cc flags for boot.
	FlagsCCBootArch   []string // Arch-specific clang cc flags for boot.
	FlagsLinkRuntime  []string // All clang linker flags for sys runtime.
	FlagsLinkBoot     []string // All clang linker flags for boot.
	FlagsLinkBootArch []string // Arch-specific clang linker flags for boot.
	FlagsRCBoot       []string // llvm-rc flags for boot.
	FlagsNasmBoot     []string // nasm for boot.
}

// initContext initializes a new osbuild context and quickly checks that the
// current working directory is the source root, then loads the configuration.
func initContext(profilename, confname, cachedir string, verbose bool) *OsbuildContext {
	ctxt := &OsbuildContext{ExeContext: &exe.ExeContext{}}

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
	ctxt.CacheDir, _ = filepath.Abs(filepath.Clean(cachedir))
	ctxt.IsVerbose = verbose

	fatalReadError := func(msg string) {
		exe.Fatal("reading build conf", errors.New(msg), ctxt.ExeContext)
	}

	if messages, err := api.UnmarshalApiProtoMessages(confname, ""); err != nil {
		exe.Fatal("unmarshaling build conf", err, ctxt.ExeContext)
	} else if len(messages) != 1 {
		fatalReadError("expected exactly 1 message from " + confname)
	} else if msg := messages[0]; msg.Kind+"/"+msg.Version != "os.build.BuildConfiguration/v0" {
		fatalReadError("got unexpected message kind")
	} else if confDef, ok := msg.Def.(*api_os_build_v0.BuildConfiguration); !ok {
		fatalReadError("message type error")
	} else {
		foundProfile := false
		for _, profile := range confDef.Profiles {
			if profile.Name == profilename {
				ctxt.Profile = profile
				foundProfile = true
				break
			}
		}
		if !foundProfile {
			fatalReadError("build profile not found: " + profilename)
		}
		ctxt.Conf = confDef
	}

	ctxt.Outpath, _ = filepath.Abs(filepath.Clean(ctxt.Profile.Artifact))

	return ctxt
}
