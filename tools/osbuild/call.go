package main

import (
	"alt-os/exe"
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// ccompile compiles the specified filename to a relocatable object file.
func ccompile(filename string, print, isUnchecked bool, ctxt *OsbuildContext) (objname string) {

	// Determine the sys type from the file path name and prepare the output.
	systype := strings.TrimPrefix(filename, filepath.Clean(path.Join(ctxt.SrcRootDir, "sys")))
	for {
		if d, f := filepath.Split(systype); d == "/" || d == "\\" || d == "" || f == systype || d == systype {
			systype = filepath.Clean(f)
			break
		} else {
			systype = filepath.Clean(d)
		}
	}

	var relname string
	if strings.HasPrefix(filename, ctxt.CacheDir) {
		relname = strings.TrimPrefix(filename, ctxt.CacheDir)
		objname = ctxt.CacheDir
	} else if strings.HasPrefix(filename, ctxt.SrcRootDir) {
		relname = strings.TrimPrefix(filename, ctxt.SrcRootDir)
		objname = path.Join(ctxt.CacheDir, "src")
	} else {
		exe.Fatal("compiling CXX", errors.New("unrecognized source file prefix"), ctxt.ExeContext)
	}
	if d, f := filepath.Split(relname); isUnchecked {
		objname = path.Join(objname, d, "unchecked", strings.TrimSuffix(f, path.Ext(f))+".o")
	} else {
		objname = path.Join(objname, d, "checked", strings.TrimSuffix(f, path.Ext(f))+".o")
	}
	objname = filepath.Clean(objname)
	if err := os.MkdirAll(filepath.Dir(objname), 0755); err != nil {
		exe.Fatal("creating "+filepath.Dir(objname), err, ctxt.ExeContext)
	}

	// Set the compiler command and flags.
	languageFlag := ""
	if ext := path.Ext(filename); ext == ".c" {
		languageFlag = "-xc"
	}

	preprocessArgs := []string{languageFlag}
	compileArgs := []string{languageFlag}
	ccCmd := ""
	switch systype {
	default:
		exe.Fatal("compiling CXX", errors.New("unrecognized systype: "+systype), ctxt.ExeContext)
	case "runtime":
		preprocessArgs = append(preprocessArgs, ctxt.FlagsCCRuntime...)
		compileArgs = append(compileArgs, ctxt.FlagsCCRuntime...)
		ccCmd = path.Join(ctxt.LlvmDir, "clang")
	}
	if isUnchecked {
		preprocessArgs = append(preprocessArgs, "-fno-stack-protector")
		compileArgs = append(compileArgs, "-fno-stack-protector")
	}
	preprocessArgs = append(preprocessArgs, "-frewrite-includes", "-M", "-E", filename)
	compileArgs = append(compileArgs, "-c", "-o"+objname, filename)

	if stdOut, stdErr, err := exe.Doexec("", ccCmd, compileArgs...); err != nil {
		exe.Fatal("compiling CXX "+filename, exe.ErrOutput(stdOut, stdErr, err), ctxt.ExeContext)
	}

	return objname
}

// objlink links the specified ELF object files into the specified output file using the specified toolpath.
func objlink(outname, outFilename, toolpath string, flagArgs, objpaths []string, isUnchecked bool, ctxt *OsbuildContext) {
	ctxt.Logger.WithFields(exe.Fields{"outname": outname}).Info("Linking")
	lds := path.Join(ctxt.SrcRootDir, "sys", "linker-scripts", "runtime.lds")
	linkArgs := []string{"-Wl,--script=" + lds}
	linkArgs = append(linkArgs, flagArgs...)
	if isUnchecked {
		linkArgs = append(linkArgs, "-fno-stack-protector")
	}
	linkArgs = append(linkArgs, "-o"+outFilename)
	linkArgs = append(linkArgs, objpaths...)
	if err := os.MkdirAll(filepath.Dir(outFilename), 0755); err != nil {
		exe.Fatal("creating "+filepath.Dir(outFilename), err, ctxt.ExeContext)
	}
	if stdOut, stdErr, err := exe.Doexec("", path.Join(toolpath, "clang"), linkArgs...); err != nil {
		exe.Fatal("linking "+outFilename, exe.ErrOutput(stdOut, stdErr, err), ctxt.ExeContext)
	}
	stripArgs := []string{"--strip-unneeded", outFilename}
	if stdOut, stdErr, err := exe.Doexec("", path.Join(toolpath, "llvm-strip"), stripArgs...); err != nil {
		exe.Fatal("stripping "+outFilename, exe.ErrOutput(stdOut, stdErr, err), ctxt.ExeContext)
	}
}

// srcscan recursively scans the specified root path for files to build for the specified stage. Any
// discovered paths are sent to the channel, and wg is signaled when done.
func srcscan(rootPath, ext, stage string, ch chan<- string, wg *sync.WaitGroup, ctxt *OsbuildContext) {
	getByStage := func(path string, wg *sync.WaitGroup) {
		if stage == "" {
			ch <- path
		} else {
			if f, err := os.Open(path); err != nil {
				exe.Fatal("scanning sources: "+path, err, ctxt.ExeContext)
			} else {
				defer f.Close()
				scanner := bufio.NewScanner(f)
				if scanner.Scan() {
					if strings.HasPrefix(scanner.Text(), _COMMENT_OSBUILD) {
						comment := strings.TrimPrefix(scanner.Text(), _COMMENT_OSBUILD)
						for _, target := range strings.Split(comment, ",") {
							if target == stage {
								ch <- path
								break
							}
						}
					}
				}
			}
		}
		wg.Done()
	}
	wg2 := &sync.WaitGroup{}
	filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if !d.Type().IsRegular() || !strings.HasSuffix(path, ext) {
			return nil
		}
		wg2.Add(1)
		go getByStage(filepath.Clean(path), wg2)
		return nil
	})
	wg2.Wait()
	wg.Done()
}
