package main

/*
#include "vmx.h"
*/
import "C"

func isVmxSupported() bool {
	hasVmxFeature := int(C.exe_vmm_vmx_has_feature())
	return hasVmxFeature == 1
}
