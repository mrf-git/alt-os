package main

import (
	"alt-os/exe"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// osbuild orchestrates the actual build process.
func osbuild(ctxt *OsbuildContext) {

	// Initialize high-level architecture specifics.
	switch ctxt.Profile.Arch {
	case "amd64":
		ctxt.ArchEFI = "X64"
	default:
		exe.Fatal("initializing build", errors.New("unrecognized profile arch: "+ctxt.Profile.Arch), ctxt.ExeContext)
	}

	// Initialize and display information about this current build session.
	toolinit(ctxt)
	infoinit(ctxt)
	if ctxt.IsVerbose {
		fmt.Println("-------------------------")
		fmt.Printf("%-24s: %s\n", "Profile", ctxt.Profile.Name)
		fmt.Printf("%-24s: %s\n", "OsName", ctxt.Conf.OsName)
		fmt.Printf("%-24s: %s\n", "OsVersion", ctxt.BuildInfo.VersionStr)
		fmt.Printf("%-24s: (%d, %d, %d)\n", "OsVersionNum", ctxt.BuildInfo.VersionMajorNum,
			ctxt.BuildInfo.VersionMinorNum, ctxt.BuildInfo.RevisionNum)
		fmt.Printf("%-24s: %s\n", "OsArch", ctxt.Profile.Arch)
		fmt.Printf("%-24s: %s\n", "BuildId", ctxt.BuildInfo.Uuid)
		fmt.Printf("%-24s: %s\n", "BuildTime", time.Unix(int64(ctxt.BuildInfo.Timestamp), 0).UTC().String())
		fmt.Println()
		fmt.Printf("%-24s: %s\n", "User", ctxt.BuildInfo.Username)
		fmt.Printf("%-24s: %s\n", "ScmAuthorName", ctxt.BuildInfo.Scm.AuthorName)
		fmt.Printf("%-24s: %s\n", "ScmAuthorEmail", ctxt.BuildInfo.Scm.AuthorEmail)
		fmt.Println()
		fmt.Printf("%-24s: %s\n", "BuildHostOs", ctxt.BuildInfo.HostOs)
		fmt.Printf("%-24s: %s\n", "BuildHostArch", ctxt.BuildInfo.HostArch)
		fmt.Printf("%-24s: %s\n", "BuildGoVersion", ctxt.BuildInfo.GoVersionStr)
		fmt.Println()
		fmt.Printf("%-24s: %s\n", "ScmCommitHash", ctxt.BuildInfo.Scm.CommitHash)
		fmt.Printf("%-24s: %s\n", "ScmBranchIsClean", strconv.FormatBool(ctxt.BuildInfo.Scm.IsClean))
		fmt.Printf("%-24s: %s\n", "ScmCommitIsSigned", strconv.FormatBool(ctxt.BuildInfo.Scm.IsSigned))
		fmt.Printf("%-24s: %s\n", "ScmAuthorTime", time.Unix(int64(ctxt.BuildInfo.Scm.AuthorTime), 0).UTC().String())
		fmt.Printf("%-24s: %s\n", "ScmRemoteUrl", ctxt.BuildInfo.Scm.RemoteUrl)
		fmt.Printf("%-24s: %s\n", "ScmBranchRemote", ctxt.BuildInfo.Scm.BranchRemote)
		fmt.Printf("%-24s: %s\n", "ScmBranchLocal", ctxt.BuildInfo.Scm.BranchLocal)
		fmt.Println()
		fmt.Printf("%-24s: %s\n", "Edk2GitUrl", ctxt.Conf.Deps.Edk2.GitUrl)
		fmt.Printf("%-24s: %s\n", "Edk2Tag", ctxt.Conf.Deps.Edk2.Tag)
		fmt.Printf("%-24s: %s\n", "AcpicaGitUrl", ctxt.Conf.Deps.Acpica.GitUrl)
		fmt.Printf("%-24s: %s\n", "AcpicaTag", ctxt.Conf.Deps.Acpica.Tag)
		fmt.Println("-------------------------")
	}

	// Verify that we can continue with the build in accordance with configuration.
	if !ctxt.Profile.AllowUnclean && !ctxt.BuildInfo.Scm.IsClean {
		exe.Fatal("initializing build", errors.New("unclean source repository not allowed for profile"), ctxt.ExeContext)
	}
	if !ctxt.Profile.AllowUnsigned && !ctxt.BuildInfo.Scm.IsSigned {
		exe.Fatal("initializing build", errors.New("unsigned source commit not allowed for profile"), ctxt.ExeContext)
	}

	fmt.Println("Starting OS build...")
	fmt.Println()

	// TODO complete build
}
