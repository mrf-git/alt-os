package machine

import (
	api_os_machine_image_v0 "alt-os/api/os/machine/image/v0"
	"fmt"
	"os"
	"path/filepath"
)

// CreateImage first calls ValidateVirtualMachine to ensure that the virtual
// machine definition is valid, then creates the image in the specified root
// directory according to the definition.
func CreateImage(def *api_os_machine_image_v0.VirtualMachine, rootDir string) error {
	if err := ValidateVirtualMachine(def); err != nil {
		return err
	}
	// makeError := func(msg string) error {
	// 	return errors.New(msg + " for creating os.machine.image.VirtualMachine")
	// }

	// Create/initialize the bundle output directory.
	imageDir := filepath.Clean(filepath.Join(rootDir, def.ImageDir))
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return err
	}

	// TODO
	fmt.Println(def.GoString())

	return nil
}
