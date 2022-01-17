package code

import (
	"bytes"
	"debug/elf"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
)

var _ELF_PREFIX = []byte{0x7f, 'E', 'L', 'F'}
var _ELF_ABI_VERSION = 0
var _ELF_DYN_TYPE = 3
var _ELF_OSABI = 88
var _ELF_INTERP = "alt-os"
var _ELF_RELA_SIZE = 24

// FromFile reads and returns the ELF file from the specified filename as
// executable code. The entire program is read into memory, and if it is
// bigger than alt-os.os.limits.MAX_EXECUTABLE_SIZE, an error is returned.
func FromFile(filename string) (ExecutableCode, error) {

	var openFile *os.File
	if f, err := os.Open(filename); err != nil {
		return nil, err
	} else {
		openFile = f
	}

	prefixBuf := make([]byte, len(_ELF_PREFIX))
	exeCode := &_ExecutableCode{}
	var retErr error

	if _, err := openFile.Read(prefixBuf); err != nil {
		retErr = err
		goto closeAndReturn
	} else if !bytes.HasPrefix(prefixBuf, _ELF_PREFIX) {
		retErr = errors.New("bad ELF prefix")
		goto closeAndReturn
	} else {
		openFile.Seek(0, io.SeekStart)
		if err := readExecutableCodeFromElf(openFile, exeCode); err != nil {
			retErr = err
			goto closeAndReturn
		}
	}

closeAndReturn:
	openFile.Close()
	return exeCode, retErr
}

// alignVal returns the specified value aligned to the specified alignment.
func alignVal(val, alignment int) int {
	if alignment == 0 {
		return val
	}
	return val + ((alignment - val) & (alignment - 1))
}

// readExecutableCodeFromElf reads executable code bytes from the specified reader.
func readExecutableCodeFromElf(r io.ReaderAt, exeCode *_ExecutableCode) error {
	var elfFile *elf.File
	if f, err := elf.NewFile(r); err != nil {
		return err
	} else {
		elfFile = f
	}
	if elfFile.Version != elf.EV_CURRENT || elfFile.ABIVersion != uint8(_ELF_ABI_VERSION) {
		return fmt.Errorf("unexpected ELF version: %s, %d", elfFile.Version.String(), elfFile.ABIVersion)
	}
	if elfFile.Type != elf.Type(_ELF_DYN_TYPE) {
		return fmt.Errorf("bad ELF type: %s", elfFile.Type.String())
	}
	if elfFile.OSABI != elf.OSABI(_ELF_OSABI) {
		return fmt.Errorf("bad ELF OSABI: %s", elfFile.OSABI.String())
	}
	var err error
	switch runtime.GOARCH {
	default:
		err = fmt.Errorf("unhandled arch: %s", runtime.GOARCH)
	case "amd64":
		err = readAmd64CodeFromElf(r, elfFile, exeCode)
	}
	return err
}
