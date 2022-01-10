#include "vmx.h"

// Call the CPUID instruction.
static inline void cpuid(int *a_in_out_val, int *b_out_val, int *c_in_out_val, int *d_out_val) {
    int a_val, b_val, c_val, d_val;
    asm("cpuid"
        : "=a"(a_val),"=b"(b_val),"=c"(c_val),"=d"(d_val)
        : "a"(*a_in_out_val),"c"(*c_in_out_val)
    );
    *a_in_out_val = a_val;
    *b_out_val = b_val;
    *c_in_out_val = c_val;
    *d_out_val = d_val;
}

// exe_vmm_vmx_has_feature
int exe_vmm_vmx_has_feature(){
    int a_val = 1;
    int b_val = 0;
    int c_val = 0;
    int d_val = 0;
    cpuid(&a_val, &b_val, &c_val, &d_val);
    return (c_val & (1 << 5)) >> 5;
}
