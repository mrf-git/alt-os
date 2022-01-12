package code

import (
	"bytes"
	"errors"
	"os"
)

const _BUFFER_SIZE = 10000

var _ELF_PREFIX = []byte{0x7f, 'E', 'L', 'F'}

// FromFile reads and returns the specified filename as executable code.
func FromFile(filename string) (*ExecutableCode, error) {

	var openFile *os.File
	if f, err := os.Open(filename); err != nil {
		return nil, err
	} else {
		openFile = f
	}

	readBuf := make([]byte, _BUFFER_SIZE)
	exeCode := &ExecutableCode{}
	var retErr error

	if _, err := openFile.Read(readBuf[:len(_ELF_PREFIX)]); err != nil {
		retErr = err
		goto closeAndReturn
	} else if !bytes.HasPrefix(readBuf, _ELF_PREFIX) {
		retErr = errors.New("bad ELF prefix")
		goto closeAndReturn
	} else if err := readExecutableCodeFromElf(openFile, exeCode, readBuf); err != nil {
		retErr = err
		goto closeAndReturn
	}

closeAndReturn:
	openFile.Close()
	return exeCode, retErr
}

// readExecutableCodeFromElf reads executable code bytes from the specified open file
// assuming the ELF prefix has already been read.
func readExecutableCodeFromElf(openFile *os.File, exeCode *ExecutableCode,
	readBuf []byte) error {

	// TODO

	return nil
}
