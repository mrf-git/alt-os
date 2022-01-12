package main

/*
#include "error-codes.h"
#include "vmx.h"
*/
import "C"
import (
	"errors"
	"fmt"
)

// isVmxSupported returns true if CPU virtual machine extension
// features are available.
func isVmxSupported() bool {
	hasVmxFeature := int(C.exe_vmm_vmx_has_feature())
	return hasVmxFeature == 1
}

// vmxOn enables vmx mode and begins virtualizing privilege level 0.
func vmxOn() error {
	statusCode := int(C.exe_vmm_vmx_on())
	switch statusCode {
	default:
		return fmt.Errorf("unknown error code: %d", statusCode)
	case C.VMM_ERR_NONE:
		break
	case C.VMM_VMX_ERR_NO_MSR:
		return errors.New("host cpu does not support msr")
	}
	return nil
}

// vmxOff disables vmx mode and exits the virtualization environment.
func vmxOff() error {
	statusCode := int(C.exe_vmm_vmx_off())
	switch statusCode {
	default:
		return fmt.Errorf("unknown error code: %d", statusCode)
	case C.VMM_ERR_NONE:
		break
	}
	return nil
}
