package container

import (
	api_os_container_bundle_v0 "alt-os/api/os/container/bundle/v0"
	api_os_container_machine_v0 "alt-os/api/os/container/machine/v0"
	api_os_container_process_v0 "alt-os/api/os/container/process/v0"
	"errors"
)

// ValidateBundle verifies that all values of the Bundle are valid
// according to OCI specifications. Process is required under the assumption
// that ContainerRuntimeService.Start will be called on the created bundle.
func ValidateBundle(def *api_os_container_bundle_v0.Bundle) error {
	makeError := func(msg string) error {
		return errors.New(msg + " for validating os.container.bundle.Bundle")
	}
	if def == nil {
		return makeError("missing Bundle")
	}
	for _, vol := range def.VolumeMounts {
		if vol.Destination == "" {
			return makeError("missing definition for `volumeMounts.destination`")
		}
	}
	if def.Process == nil {
		return makeError("missing definition for `process`")
	}
	if def.Process.Terminal != nil && def.Process.Terminal.Enable {
		if def.Process.Terminal.Height == 0 || def.Process.Terminal.Width == 0 {
			return makeError("bad terminal size")
		}
	}
	if def.Process.Cwd == "" {
		return makeError("missing definition for `process.cwd`")
	}
	for _, envVar := range def.Process.Env {
		if envVar.Name == "" {
			return makeError("missing name for `process.env` var")
		}
	}
	sawRlimit := map[string]struct{}{}
	for _, rlimit := range def.Process.Rlimits {
		if rlimit.Type == api_os_container_process_v0.ResourceLimitType_RLIMIT_NONE {
			return makeError("null `process.rlimits.type`")
		}
		if !rlimit.HardUnlimited && rlimit.HardValue == _RLIMIT_UNLIMITED_VALUE {
			return makeError("bad `process.rlimits.hardValue`")
		}
		if !rlimit.SoftUnlimited && rlimit.SoftValue == _RLIMIT_UNLIMITED_VALUE {
			return makeError("bad `process.rlimits.softValue`")
		}
		if _, ok := sawRlimit[rlimit.Type.String()]; ok {
			return makeError("duplicate `process.rlimits.type` " + rlimit.Type.String())
		}
		sawRlimit[rlimit.Type.String()] = struct{}{}
	}
	if def.Process.User == nil {
		return makeError("missing definition for `process.user`")
	}
	if def.VirtualMachine != nil {
		if def.VirtualMachineFile != "" {
			return errors.New("multiple virtual machine definition fields")
		}
		if err := ValidateContainerMachine(def.VirtualMachine); err != nil {
			return nil
		}
	}

	return nil
}

// ValidateContainerMachine verifies that all values of the ContainerMachine
// are valid.
func ValidateContainerMachine(def *api_os_container_machine_v0.ContainerMachine) error {
	makeError := func(msg string) error {
		return errors.New(msg + " for validating os.container.machine.ContainerMachine")
	}
	if def == nil {
		return makeError("missing ContainerMachine")
	}
	if def.Processors == 0 {
		return makeError("bad `ContainerMachine.processors`")
	}
	if def.Memory == 0 {
		return makeError("bad `ContainerMachine.memory`")
	}
	return nil
}
