#ifndef EXE_VMM_RUNTIME_VMX_H
#define EXE_VMM_RUNTIME_VMX_H

typedef __INTPTR_TYPE__ intn_t;

// exe_vmm_vmx_has_feature returns 1 if the cpu has the vmx feature, or 0 otherwise.
intn_t exe_vmm_vmx_has_feature();

// exe_vmm_vmx_on enables vmx mode and begins virtualizing privilege level 0.
// Returns 0 on success or error code on failure.
intn_t exe_vmm_vmx_on();

// exe_vmm_vmx_off disables vmx mode and exits the virtualization environment.
intn_t exe_vmm_vmx_off();

#endif // EXE_VMM_RUNTIME_VMX_H
