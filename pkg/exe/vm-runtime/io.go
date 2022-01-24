package main

import (
	"alt-os/exe"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

const _IO_BUFFER_SIZE = 10000

// ioServiceParams holds parameters for the ioService method.
type ioServiceParams struct {
	comSocks [4]net.Listener
	vmEnv    *_VmEnvironment
}

// ioService services standard input, output, and error for the virtual
// machine specified in the parameters.
func ioService(params *ioServiceParams) {
	logger := params.vmEnv.logger

	// Accept all the socket listeners.
	comConns := [4]net.Conn{}
	for i, listener := range params.comSocks {
		if conn, err := listener.Accept(); err != nil {
			logger.WithFields(exe.Fields{
				"err": err.Error(),
				"com": i + 1,
			}).Error("failed to accept com socket")
			return
		} else {
			comConns[i] = conn
		}
	}

	// Route the IO.
	ioData := make([]byte, _IO_BUFFER_SIZE)
	for {
		comConns[0].SetDeadline(time.Now().Add(50 * time.Millisecond))
		if n, err := comConns[0].Read(ioData); err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
			logger.Error(err.Error())
		} else if n > 0 {
			// fmt.Print(string(ioData[:n]))
			// TODO put the data where it goes
		}
		comConns[1].SetDeadline(time.Now().Add(50 * time.Millisecond))
		if n, err := comConns[1].Read(ioData); err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
			logger.Error(err.Error())
		} else if n > 0 {
			// fmt.Println(string(ioData[:n]))
			// TODO put the data where it goes
		}
		comConns[2].SetDeadline(time.Now().Add(50 * time.Millisecond))
		if n, err := comConns[2].Read(ioData); err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
			logger.Error(err.Error())
		} else if n > 0 {
			// fmt.Println(string(ioData[:n]))
			// TODO put the data where it goes
		}
		comConns[3].SetDeadline(time.Now().Add(50 * time.Millisecond))
		if n, err := comConns[3].Read(ioData); err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
			logger.Error(err.Error())
		} else if n > 0 {
			fmt.Print(string(ioData[:n]))
			// TODO put the data where it goes
		}
		// time.Sleep(100 * time.Millisecond)
	}
}
