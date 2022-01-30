package machine

import (
	api_os_machine_image_v0 "alt-os/api/os/machine/image/v0"
	"errors"
)

// ValidateVirtualMachine verifies that all values of the VirtualMachine
// are valid.
func ValidateVirtualMachine(def *api_os_machine_image_v0.VirtualMachine, virtualized bool) error {
	makeError := func(msg string) error {
		return errors.New(msg + " for validating os.machine.image.VirtualMachine")
	}
	if def == nil {
		return makeError("missing VirtualMachine")
	}
	if def.ImageDir == "" {
		return makeError("bad `VirtualMachine.imageDir`")
	}
	if def.Processors == 0 {
		return makeError("bad `VirtualMachine.processors`")
	}
	if def.Memory == 0 {
		return makeError("bad `VirtualMachine.memory`")
	}
	for _, serial := range def.Serial {
		if serial.Address != 0 && serial.Port != 0 {
			return makeError("bad `VirtualMachine.serial`: cannot have both port and address")
		}
	}
	if virtualized {
		if def.EfiPath == "" {
			return makeError("missing `VirtualMachine.efiPath`")
		}
		if def.BiosImage == "" {
			return makeError("missing `VirtualMachine.biosImage`")
		}
		if def.VarsImage == "" {
			return makeError("missing `VirtualMachine.varsImage`")
		}
	}
	return nil
}
