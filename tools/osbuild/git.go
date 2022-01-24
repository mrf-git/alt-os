package main

import (
	api_os_build_v0 "alt-os/api/os/build/v0"
	"path"
	"strings"
	"time"

	"alt-os/exe"
)

// gitinfo gets a current scm snapshot with information obtained from git.
func gitinfo(ctxt *OsbuildContext) *api_os_build_v0.ScmSnapshot {
	gitfatal := func(name, out, err string) {
		exe.Fatal("getting git "+name, exe.ErrOutput(out, err, nil), ctxt.ExeContext)
	}

	gitshow := func(name, format string) string {
		if stdOut, stdErr, err := exe.Doexec("", "git", "show", "-s", "--format="+format, "HEAD^{commit}"); err != nil {
			gitfatal("show "+name, stdOut, stdErr)
		} else {
			return strings.TrimSpace(stdOut)
		}
		return ""
	}

	authorName := gitshow("authorName", "%an")
	authorEmail := gitshow("authorName", "%ae")
	commitHash := gitshow("commitHash", "%H")
	commitSigned := gitshow("commitSigned", "%G?") == "G"
	var authorTime time.Time
	if t, err := time.Parse(time.RFC3339, gitshow("authorTime", "%aI")); err != nil {
		exe.Fatal("parsing git author time", err, ctxt.ExeContext)
	} else {
		authorTime = t
	}

	var gitUrl string
	if stdOut, stdErr, err := exe.Doexec("", "git", "remote", "get-url", "origin"); err != nil {
		gitfatal("remote url", stdOut, stdErr)
	} else {
		gitUrl = strings.TrimSpace(stdOut)
	}

	var gitBranchLocal string
	var gitBranchRemote string
	gitClean := false
	if stdOut, stdErr, err := exe.Doexec("", "git", "status"); err != nil {
		gitfatal("status", stdOut, stdErr)
	} else {
		lines := strings.Split(stdOut, "\n")
		if len(lines) < 2 {
			gitfatal("status", stdOut, stdErr)
		}
		branchline := strings.TrimSpace(lines[0])
		if !strings.HasPrefix(branchline, "On branch ") {
			gitfatal("status local branch", stdOut, stdErr)
		}
		gitBranchLocal = strings.TrimPrefix(branchline, "On branch ")
		branchline = strings.TrimSpace(lines[1])
		if !strings.HasPrefix(branchline, "Your branch is up to date with '") {
			gitfatal("status remote branch", stdOut, stdErr)
		}
		gitBranchRemote = strings.TrimPrefix(branchline, "Your branch is up to date with '")
		gitBranchRemote = strings.TrimSuffix(gitBranchRemote, "'.")
		for _, line := range lines[2:] {
			if line == "nothing to commit, working tree clean" {
				gitClean = true
				break
			}
		}
	}

	scmSnapshot := &api_os_build_v0.ScmSnapshot{
		CommitHash:   commitHash,
		AuthorName:   authorName,
		AuthorEmail:  authorEmail,
		AuthorTime:   uint64(authorTime.UTC().Unix()),
		BranchLocal:  gitBranchLocal,
		BranchRemote: gitBranchRemote,
		RemoteUrl:    gitUrl,
		IsSigned:     commitSigned,
		IsClean:      gitClean,
	}
	return scmSnapshot
}

// githash returns the git hash of the last commit on HEAD for the git repo at the specified path.
func githash(curpath string, ctxt *OsbuildContext) (string, error) {
	gitdir := path.Join(curpath, ".git")
	if stdOut, stdErr, err := exe.Doexec("", "git", "--git-dir="+gitdir, "--work-tree="+curpath, "show", "-s",
		"--format=%H", "HEAD^{commit}"); err != nil {
		return "", exe.ErrOutput(stdOut, stdErr, err)
	} else {
		return strings.TrimSpace(stdOut), nil
	}
}

// isgitclean returns true if there are no local changes in the git repo at the specified path.
func isgitclean(curpath string, ctxt *OsbuildContext) bool {
	gitdir := path.Join(curpath, ".git")
	if stdOut, stdErr, err := exe.Doexec("", "git", "--git-dir="+gitdir, "--work-tree="+curpath, "status"); err != nil {
		exe.Fatal("git status "+curpath, exe.ErrOutput(stdOut, stdErr, err), ctxt.ExeContext)
	} else {
		for _, line := range strings.Split(stdOut, "\n") {
			if strings.TrimSpace(line) == "nothing to commit, working tree clean" {
				return true
			}
		}
	}
	return false
}
