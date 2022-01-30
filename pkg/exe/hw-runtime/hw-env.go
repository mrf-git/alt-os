package main

import (
	"alt-os/api"
	api_os_machine_image_v0 "alt-os/api/os/machine/image/v0"
	"alt-os/exe"
	"errors"
)

// HwEnvironment encapsulates the runtime environment for a single
// hardware device to run virtual machine containers.
type HwEnvironment interface {
	// Run creates a new privilege level 0 environment in the
	// hardware device and starts running virtual machine code in it.
	// Sends the code returned by main to returnCodeCh when the
	// virtual machine exits.
	Run(signalCh <-chan int, returnCodeCh chan<- int) error
}

// newHwEnvironment returns a newly-instantiated HwEnvironment.
func newHwEnvironment(hwDefFile string, ctxt *HwRuntimeContext) HwEnvironment {
	fatalReadError := func(msg string) {
		exe.Fatal("reading hardware definition file", errors.New(msg), ctxt.ExeContext)
	}
	if messages, err := api.UnmarshalApiProtoMessages(hwDefFile, ""); err != nil {
		exe.Fatal("unmarshaling proto messages", err, ctxt.ExeContext)
	} else if len(messages) != 1 {
		fatalReadError("expected exactly 1 message from " + hwDefFile)
	} else if msg := messages[0]; msg.Kind+"/"+msg.Version != "os.machine.image.VirtualMachine/v0" {
		fatalReadError("got unexpected message kind")
	} else if hwDef, ok := msg.Def.(*api_os_machine_image_v0.VirtualMachine); !ok {
		fatalReadError("message type error")
	} else {
		return &_HwEnvironment{
			logger:     exe.NewLogger(ctxt.ExeLoggerConf),
			ctxt:       ctxt,
			machineDef: hwDef,
		}
	}
	return nil
}

type _HwEnvironment struct {
	logger     exe.Logger
	ctxt       *HwRuntimeContext
	machineDef *api_os_machine_image_v0.VirtualMachine
}

func (hwEnv *_HwEnvironment) Run(signalCh <-chan int, returnCodeCh chan<- int) error {

	return nil
}
