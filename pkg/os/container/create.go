package container

import (
	"alt-os/api"
	api_os_container_bundle_v0 "alt-os/api/os/container/bundle/v0"
	api_os_container_process_v0 "alt-os/api/os/container/process/v0"
	api_os_machine_image_v0 "alt-os/api/os/machine/image/v0"
	"alt-os/os/machine"
	"encoding/json"
	"errors"
	"fmt"
	oci "oci/specs"
	"os"
	"path/filepath"
)

const _MIN_VERSION_ANNOTATION_KEY = "alt-os/oci-annotation/min-version"
const _MIN_VERSION_ANNOTATION_VALUE = "0"
const _ROOT_DIR_NAME = "rootfs"
const _CONFIG_JSON_NAME = "config.json"
const _RLIMIT_UNLIMITED_VALUE = uint64(0xFFFFFFFFFFFFFFFF)

// CreateBundle first calls ValidateBundle to ensure that the bundle
// definition is valid, then creates the bundle in the specified root
// directory according to the definition.
func CreateBundle(def *api_os_container_bundle_v0.Bundle, rootDir string) error {
	if err := ValidateBundle(def); err != nil {
		return err
	}
	makeError := func(msg string) error {
		return errors.New(msg + " for creating os.container.bundle.Bundle")
	}

	// Create/initialize the bundle output directory.
	bundleDir := filepath.Clean(filepath.Join(rootDir, def.BundleDir))
	if err := os.MkdirAll(filepath.Join(bundleDir, _ROOT_DIR_NAME), 0755); err != nil {
		return err
	}

	// Initialize the virtual machine if defined.
	var vm *api_os_machine_image_v0.VirtualMachine
	if def.VirtualMachine != nil {
		vm = def.VirtualMachine
	} else if def.VirtualMachineFile != "" {
		if messages, err := api.UnmarshalApiProtoMessages(def.VirtualMachineFile, ""); err != nil {
			return err
		} else if len(messages) != 1 {
			return makeError("expected exactly 1 message from " + def.VirtualMachineFile)
		} else if msg := messages[0]; msg.Kind+"/"+msg.Version != "os.machine.image.VirtualMachine/v0" {
			return makeError("got unexpected message kind")
		} else if typedDef, ok := msg.Def.(*api_os_machine_image_v0.VirtualMachine); !ok {
			return makeError("message type error")
		} else if err := machine.ValidateVirtualMachine(typedDef); err != nil {
			return err
		} else {
			vm = typedDef
		}
	}
	if vm != nil {
		// TODO create and initialize the defined virtual machine.
		fmt.Println(vm.GoString())
	}

	// Create and write the config.json file in the bundle output directory.
	makeEnvStrs := func(vars []*api_os_container_process_v0.EnvironmentVariable) (strs []string) {
		for _, envVar := range vars {
			strs = append(strs, envVar.Name+"="+envVar.Value)
		}
		return
	}
	makeCapStrs := func(caps []api_os_container_process_v0.Capability) (strs []string) {
		for _, cap := range caps {
			strs = append(strs, cap.String())
		}
		return
	}
	ociProcess := &oci.Process{
		Cwd: def.Process.Cwd,
		User: oci.User{
			UID:            def.Process.User.Uid,
			GID:            def.Process.User.Gid,
			Umask:          def.Process.User.Umask,
			AdditionalGids: def.Process.User.AdditionalGids,
		},
		Args: def.Process.Args,
		Env:  makeEnvStrs(def.Process.Env),
	}
	if def.Process.Terminal != nil && def.Process.Terminal.Enable {
		ociProcess.Terminal = true
		ociProcess.ConsoleSize = &oci.Box{
			Height: uint(def.Process.Terminal.Height),
			Width:  uint(def.Process.Terminal.Width),
		}
	}
	if def.Process.Capabilities != nil {
		ociProcess.Capabilities = &oci.LinuxCapabilities{
			Permitted:   makeCapStrs(def.Process.Capabilities.Permitted),
			Bounding:    makeCapStrs(def.Process.Capabilities.Bounding),
			Effective:   makeCapStrs(def.Process.Capabilities.Effective),
			Inheritable: makeCapStrs(def.Process.Capabilities.Inheritable),
			Ambient:     makeCapStrs(def.Process.Capabilities.Ambient),
		}
	}
	for _, rlimit := range def.Process.Rlimits {
		limitValue := func(isUnlimited bool, value uint64) uint64 {
			if isUnlimited {
				return _RLIMIT_UNLIMITED_VALUE
			}
			return value
		}
		ociProcess.Rlimits = append(ociProcess.Rlimits, oci.POSIXRlimit{
			Type: rlimit.Type.String(),
			Soft: limitValue(rlimit.SoftUnlimited, rlimit.SoftValue),
			Hard: limitValue(rlimit.HardUnlimited, rlimit.HardValue),
		})
	}
	ociConfig := &oci.Spec{
		Version:     oci.Version,
		Hostname:    def.Hostname,
		Root:        &oci.Root{Path: _ROOT_DIR_NAME},
		Process:     ociProcess,
		Annotations: map[string]string{_MIN_VERSION_ANNOTATION_KEY: _MIN_VERSION_ANNOTATION_VALUE},
	}
	for _, vol := range def.VolumeMounts {
		ociConfig.Mounts = append(ociConfig.Mounts, oci.Mount{
			Destination: vol.Destination,
			Source:      vol.Source,
			Type:        vol.Type,
			Options:     vol.Options,
		})
	}
	if ociJson, err := json.Marshal(ociConfig); err != nil {
		return err
	} else if err := os.WriteFile(filepath.Join(bundleDir, _CONFIG_JSON_NAME), ociJson, 0666); err != nil {
		return err
	} else {
		fmt.Println(string(ociJson))
	}

	// TODO

	return nil
}
