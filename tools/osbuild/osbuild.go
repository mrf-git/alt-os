package main

import (
	"alt-os/exe"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const _COMMENT_OSBUILD = "//osbuild:" // Comment string that must be at the start of all relevant C source files.
const _MAX_NUM_OBJECTS = 8192         // Maximum number of object files to be able to handle at once.

// osbuild orchestrates the actual build process.
func osbuild(ctxt *OsbuildContext) {

	// Initialize high-level architecture specifics.
	switch ctxt.Profile.Arch {
	case "amd64":
		ctxt.ArchEFI = "X64"
	case "aarch64":
		ctxt.ArchEFI = "AA64"
	default:
		exe.Fatal("initializing build", errors.New("unrecognized profile arch: "+ctxt.Profile.Arch), ctxt.ExeContext)
	}

	// Initialize and display information about this current build session.
	toolinit(ctxt)
	infoinit(ctxt)
	if ctxt.IsVerbose {
		ctxt.Logger.WithFields(exe.Fields{
			"Profile":   ctxt.Profile.Name,
			"OsName":    ctxt.Conf.OsName,
			"OsVersion": ctxt.BuildInfo.VersionStr,
			"OsVersionNum": fmt.Sprintf("(%d, %d, %d)", ctxt.BuildInfo.VersionMajorNum,
				ctxt.BuildInfo.VersionMinorNum, ctxt.BuildInfo.RevisionNum),
			"OsArch":            ctxt.Profile.Arch,
			"BuildId":           ctxt.BuildInfo.Uuid,
			"BuildTime":         time.Unix(int64(ctxt.BuildInfo.Timestamp), 0).UTC().String(),
			"User":              ctxt.BuildInfo.Username,
			"ScmAuthorName":     ctxt.BuildInfo.Scm.AuthorName,
			"ScmAuthorEmail":    ctxt.BuildInfo.Scm.AuthorEmail,
			"BuildHostOs":       ctxt.BuildInfo.HostOs,
			"BuildHostArch":     ctxt.BuildInfo.HostArch,
			"BuildGoVersion":    ctxt.BuildInfo.GoVersionStr,
			"ScmCommitHash":     ctxt.BuildInfo.Scm.CommitHash,
			"ScmBranchIsClean":  strconv.FormatBool(ctxt.BuildInfo.Scm.IsClean),
			"ScmCommitIsSigned": strconv.FormatBool(ctxt.BuildInfo.Scm.IsSigned),
			"ScmAuthorTime":     time.Unix(int64(ctxt.BuildInfo.Scm.AuthorTime), 0).UTC().String(),
			"ScmRemoteUrl":      ctxt.BuildInfo.Scm.RemoteUrl,
			"ScmBranchRemote":   ctxt.BuildInfo.Scm.BranchRemote,
			"ScmBranchLocal":    ctxt.BuildInfo.Scm.BranchLocal,
			"Edk2GitUrl":        ctxt.Conf.Deps.Edk2.GitUrl,
			"Edk2Tag":           ctxt.Conf.Deps.Edk2.Tag,
			"AcpicaGitUrl":      ctxt.Conf.Deps.Acpica.GitUrl,
			"AcpicaTag":         ctxt.Conf.Deps.Acpica.Tag,
		}).Info("Initializing OS build")
	}

	// Verify that we can continue with the build in accordance with configuration.
	if !ctxt.Profile.AllowUnclean && !ctxt.BuildInfo.Scm.IsClean {
		exe.Fatal("initializing build", errors.New("unclean source repository not allowed for profile"), ctxt.ExeContext)
	}
	if !ctxt.Profile.AllowUnsigned && !ctxt.BuildInfo.Scm.IsSigned {
		exe.Fatal("initializing build", errors.New("unsigned source commit not allowed for profile"), ctxt.ExeContext)
	}

	ctxt.Logger.Info("Starting OS build")
	initCache(ctxt)

	// Clear the source build cache if the profile changed or user requested it.
	if ctxt.WipeSource {
		clearsourcecache(ctxt)
	} else if prevProfile, ok := ctxt.CacheManifest["profile"]; ok {
		if prevProfile != ctxt.Profile.Name {
			clearsourcecache(ctxt)
		}
	}
	ctxt.CacheManifest["profile"] = ctxt.Profile.Name

	// Acquire dependencies and perform the build.
	depbuild(ctxt)
	bootbuild(ctxt)
	finalize(ctxt)

}

// clearsourcecache clears the source cache.
func clearsourcecache(ctxt *OsbuildContext) {
	os.RemoveAll(path.Join(ctxt.CacheDir, "src"))
	for key := range ctxt.CacheManifest {
		if strings.HasPrefix(key, "boot-mtime-") || key == "bootHash" {
			delete(ctxt.CacheManifest, key)
		}
	}
}

// depbuild initializes all dependencies required to build the OS. Dependencies are cached in the build cache directory.
func depbuild(ctxt *OsbuildContext) {
	depChanged := false
	if depEdk2(ctxt) {
		depChanged = true
	}
	if depChanged {
		clearsourcecache(ctxt)
	}
}

// finalize completes any last build steps and copies the output boot image into the target output directory.
func finalize(ctxt *OsbuildContext) {
	edk2ImagePath := path.Join(ctxt.Edk2WorkspacePath, "Build", "SysModuleOutput", "RELEASE_SYS", ctxt.ArchEFI, "SysBoot.efi")
	if data, err := os.ReadFile(edk2ImagePath); err != nil {
		exe.Fatal("reading image output", err, ctxt.ExeContext)
	} else {
		os.MkdirAll(path.Dir(ctxt.Outpath), 0755)
		if f, err := os.Create(ctxt.Outpath); err != nil {
			exe.Fatal("copying image output", err, ctxt.ExeContext)
		} else {
			f.Write(data)
			f.Close()
		}
	}
	ctxt.Logger.Info("Done building")
}
