#include "vmx.h"
#include "error-codes.h"

#define FEATURE_CONTROL_MSR 0x3A


// cpuid calls the CPUID instruction and returns the register outputs.
static inline void cpuid(intn_t *a_in_out_val, intn_t *b_out_val, intn_t *c_in_out_val, intn_t *d_out_val) {
    intn_t a_val, b_val, c_val, d_val;
    asm("cpuid"
        : "=a"(a_val),"=b"(b_val),"=c"(c_val),"=d"(d_val)
        : "a"(*a_in_out_val),"c"(*c_in_out_val)
    );
    *a_in_out_val = a_val;
    *b_out_val = b_val;
    *c_in_out_val = c_val;
    *d_out_val = d_val;
}

// rdmsr reads and returns the current value of the specified machine-specific register.
// Requires privilege level 0.
static inline intn_t rdmsr(intn_t msr, intn_t *a_out_val, intn_t *d_out_val) {
    intn_t a_val, d_val;
    asm("rdmsr"
        : "=d"(d_val),"=a"(a_val)
        : "c"(msr)
    );
    *a_out_val = a_val;
    *d_out_val = d_val;
}

// exe_vmm_vmx_has_feature returns 1 if the cpu has the vmx feature, or 0 otherwise.
intn_t exe_vmm_vmx_has_feature(){
    intn_t a_val = 1;
    intn_t b_val = 0;
    intn_t c_val = 0;
    intn_t d_val = 0;
    cpuid(&a_val, &b_val, &c_val, &d_val);
    return (c_val & (1 << 5)) >> 5;
}

// exe_vmm_vmx_on enables vmx mode and begins virtualizing privilege level 0.
// Returns 0 on success or error code on failure.
intn_t exe_vmm_vmx_on(){
    // intn_t a_val = 1;
    // intn_t b_val = 0;
    // intn_t c_val = 0;
    // intn_t d_val = 0;
    // cpuid(&a_val, &b_val, &c_val, &d_val);
    // bool has_msr = ((d_val & (1 << 5)) >> 5) == 1;
    // if (has_msr) {

    // }

    // TODO

    return 0;
}


// exe_vmm_vmx_off disables vmx mode and exits the virtualization environment.
intn_t exe_vmm_vmx_off(){

    // TODO

    return 0;
}
