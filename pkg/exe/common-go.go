package exe

import (
	"bufio"
	"errors"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

// FindPackageModDir attempts to find the module directory of the specified installed Go package.
func FindPackageModDir(pkg string) (string, error) {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = build.Default.GOPATH
	}
	if goPath == "" {
		return "", errors.New("invalid GOPATH")
	}

	relDir := ""
	if stdOut, stdErr, err := Doexec("", "go", "mod", "graph"); err != nil {
		return "", ErrOutput(stdOut, stdErr, err)
	} else {
		scanner := bufio.NewScanner(strings.NewReader(stdOut))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, pkg) {
				continue
			}
			relDir = strings.SplitN(line, " ", 2)[0]
			break
		}
	}
	if relDir == "" {
		return "", errors.New("package not found")
	}

	pkgDir := filepath.Clean(filepath.Join(goPath, "pkg", "mod", relDir))
	if _, err := os.Stat(pkgDir); err != nil {
		return "", err
	}

	return pkgDir, nil
}
