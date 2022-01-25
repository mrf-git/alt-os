#include "Common.h"
#include "Cpu.h"


//
// Structure definitions.
//

typedef struct {
    // Memory address of the TSS pages.
    UINTN TssAddr;
    // Memory size of the TSS pages.
    UINTN TssSize;
    // Pointers to the relevant fields of the TSS structure needed by the CPU.
    UINT64 *Stack0Field;
    UINT64 *Stack1Field;
    UINT64 *Stack2Field;
    UINT64 *IstFields[7];
    UINT16 *IoMapField;
    // Pointers to the actual memory described by the fields of the TSS structure.
    void *Stack0Memory;
    void *Stack1Memory;
    void *Stack2Memory;
    void *IstMemory[7];
    void *IoMapMemory;

} CPU_TSS;






//
// CPU-related definitions.
//

#define CPU_MAX_NUM_IO_PORTS                    65536

#if defined(__amd64__) || defined(__x86__)
    #define CPU_MAX_NUM_SEGMENTS                8192
    #define CPU_TSS_HEADER_SIZE                 32
    typedef UINT64 CPU_SEGMENT_DESCRIPTOR;

    // static inline UINT32 SYSABI Sys_Cpu_GdtGetLimit(IN const CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     return (((UINT32)(*Desc >> 48) & 0x0F) << 16) | ((UINT32) *Desc & 0xFFFF);
    // }
    static inline void SYSABI Sys_Cpu_GdtSetLimit(IN const UINT32 Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
        *Desc &= ~((CPU_SEGMENT_DESCRIPTOR) 0xFFFF);
        *Desc &= ~(((CPU_SEGMENT_DESCRIPTOR) 0x0F) << 48);
        *Desc |= (Val & 0xFFFF);
        *Desc |= ((CPU_SEGMENT_DESCRIPTOR)(Val >> 16) & 0x0F) << 48;
    }
    // static inline UINT32 SYSABI Sys_Cpu_GdtGetBase(IN const CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     return ((((UINT32)(*Desc >> 48) & 0xFF00) | ((UINT32)(*Desc >> 32) & 0xFF)) << 16) | ((UINT32)(*Desc >> 16) & 0xFFFF);
    // }
    static inline void SYSABI Sys_Cpu_GdtSetBase(IN const UINT32 Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
        *Desc &= ~(((CPU_SEGMENT_DESCRIPTOR) 0xFFFF) << 16);
        *Desc &= ~(((CPU_SEGMENT_DESCRIPTOR) 0xFF) << 32);
        *Desc &= ~(((CPU_SEGMENT_DESCRIPTOR) 0xFF) << 48);
        *Desc |= (CPU_SEGMENT_DESCRIPTOR)(Val & 0xFFFF) << 16;
        *Desc |= ((CPU_SEGMENT_DESCRIPTOR)(Val >> 16) & 0xFF) << 32;
        *Desc |= ((CPU_SEGMENT_DESCRIPTOR)(Val >> 24) & 0xFF) << 48;
    }
    // static inline UINT8 SYSABI Sys_Cpu_GdtGetType(IN const CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     return (UINT8)(*Desc >> 40) & 0x0F;
    // }
    static inline void SYSABI Sys_Cpu_GdtSetType(IN const UINT8 Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
        *Desc &= ~(((CPU_SEGMENT_DESCRIPTOR) 0x0F) << 40);
        *Desc |= (CPU_SEGMENT_DESCRIPTOR)(Val & 0x0F) << 40;
    }
    // static inline UINT8 SYSABI Sys_Cpu_GdtGetRing(IN const CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     return (UINT8)(*Desc >> 45) & 0x03;
    // }
    static inline void SYSABI Sys_Cpu_GdtSetRing(IN const UINT8 Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
        *Desc &= ~(((CPU_SEGMENT_DESCRIPTOR) 0x03) << 45);
        *Desc |= (CPU_SEGMENT_DESCRIPTOR)(Val & 0x03) << 45;
    }
    // static inline BOOLEAN SYSABI Sys_Cpu_GdtGetIsSys(IN const CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     return ((*Desc >> 44) & 1) == 0;
    // }
    static inline void SYSABI Sys_Cpu_GdtSetIsSys(IN const BOOLEAN Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
        *Desc &= ~(((CPU_SEGMENT_DESCRIPTOR) 1) << 44);
        *Desc |= (CPU_SEGMENT_DESCRIPTOR)(Val ? 0 : 1) << 44;
    }
    // static inline BOOLEAN SYSABI Sys_Cpu_GdtGetIsPresent(IN const CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     return ((*Desc >> 47) & 1) == 1;
    // }
    static inline void SYSABI Sys_Cpu_GdtSetIsPresent(IN const BOOLEAN Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
        *Desc &= ~(((CPU_SEGMENT_DESCRIPTOR) 1) << 47);
        *Desc |= (CPU_SEGMENT_DESCRIPTOR)(Val ? 1 : 0) << 47;
    }
    // static inline BOOLEAN SYSABI Sys_Cpu_GdtGetIsAvailable(IN const CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     return ((*Desc >> 52) & 1) == 1;
    // }
    static inline void SYSABI Sys_Cpu_GdtSetIsAvailable(IN const BOOLEAN Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
        *Desc &= ~(((CPU_SEGMENT_DESCRIPTOR) 1) << 52);
        *Desc |= (CPU_SEGMENT_DESCRIPTOR)(Val ? 1 : 0) << 52;
    }
    // static inline BOOLEAN SYSABI Sys_Cpu_GdtGetIsDbSet(IN const CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     return ((*Desc >> 54) & 1) == 1;
    // }
    static inline void SYSABI Sys_Cpu_GdtSetIsDbSet(IN const BOOLEAN Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
        *Desc &= ~(((CPU_SEGMENT_DESCRIPTOR) 1) << 54);
        *Desc |= (CPU_SEGMENT_DESCRIPTOR)(Val ? 1 : 0) << 54;
    }
    // static inline BOOLEAN SYSABI Sys_Cpu_GdtGetIsLongSet(IN const CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     return ((*Desc >> 53) & 1) == 1;
    // }
    static inline void SYSABI Sys_Cpu_GdtSetIsLongSet(IN const BOOLEAN Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
        *Desc &= ~(((CPU_SEGMENT_DESCRIPTOR) 1) << 53);
        *Desc |= (CPU_SEGMENT_DESCRIPTOR)(Val ? 1 : 0) << 53;
    }
    // static inline BOOLEAN SYSABI Sys_Cpu_GdtGetIs4kGranularity(IN const CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     return ((*Desc >> 55) & 1) == 1;
    // }
    static inline void SYSABI Sys_Cpu_GdtSetIs4kGranularity(IN const BOOLEAN Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
        *Desc &= ~(((CPU_SEGMENT_DESCRIPTOR) 1) << 55);
        *Desc |= (CPU_SEGMENT_DESCRIPTOR)(Val ? 1 : 0) << 55;
    }

#else
    #error "cpu: platform not supported"
#endif


#if defined(__amd64__)
    
    #define CPU_SEGMENT_TYPE_CODE                   9   // Execute-only, accessed (to prevent infinite loops, see Intel 3.4 Vol 3A).
    #define CPU_SEGMENT_TYPE_DATA_RW                3   // accessed
    #define CPU_SEGMENT_TYPE_DATA_RO                1   // accessed
    #define CPU_SEGMENT_TYPE_DATA_STACK             7   // RW, accessed
    #define CPU_SEGMENT_TYPE_LDT                    2
    #define CPU_SEGMENT_TYPE_TSS_AVAILABLE          9
    #define CPU_SEGMENT_TYPE_TSS_BUSY               11
    #define CPU_SEGMENT_TYPE_CALL_GATE              12
    #define CPU_SEGMENT_TYPE_INTERRUPT_GATE         14
    #define CPU_SEGMENT_TYPE_TRAP_GATE              15

    
    // static inline void SYSABI Sys_Cpu_IdtGate(IN const UINT16 Selector, IN const UINT64 Offset, IN const UINT8 Ring,
    //     IN const UINT8 IstIndex, IN const BOOLEAN IsTrap, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     *Desc = 0;
    //     *(Desc+sizeof(CPU_SEGMENT_DESCRIPTOR)) = 0;
    //     *Desc |= (CPU_SEGMENT_DESCRIPTOR)(Selector) << 16;
    //     *Desc |= (CPU_SEGMENT_DESCRIPTOR)(Offset & 0xFFFF);
    //     *Desc |= (CPU_SEGMENT_DESCRIPTOR)((Offset >> 16) & 0xFFFF) << 48;
    //     *(Desc+sizeof(CPU_SEGMENT_DESCRIPTOR)) |= (CPU_SEGMENT_DESCRIPTOR)(Offset >> 32);
    //     *Desc |= (CPU_SEGMENT_DESCRIPTOR)(IstIndex & 0x07) << 32;
    //     Sys_Cpu_GdtSetIsPresent(TRUE, Desc);
    //     Sys_Cpu_GdtSetRing(Ring, Desc);
    //     if (IsTrap){
    //         Sys_Cpu_GdtSetType(CPU_SEGMENT_TYPE_TRAP_GATE, Desc);
    //     } else {
    //         Sys_Cpu_GdtSetType(CPU_SEGMENT_TYPE_INTERRUPT_GATE, Desc);
    //     }
    // }

    // static inline void SYSABI Sys_Cpu_CallGate(IN const UINT16 Selector, IN const UINT64 Offset, IN const UINT8 Ring,
    //     IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     *Desc = 0;
    //     *(Desc+sizeof(CPU_SEGMENT_DESCRIPTOR)) = 0;
    //     *Desc |= (CPU_SEGMENT_DESCRIPTOR)(Selector) << 16;
    //     *Desc |= (CPU_SEGMENT_DESCRIPTOR)(Offset & 0xFFFF);
    //     *Desc |= (CPU_SEGMENT_DESCRIPTOR)((Offset >> 16) & 0xFFFF) << 48;
    //     *(Desc+sizeof(CPU_SEGMENT_DESCRIPTOR)) |= (CPU_SEGMENT_DESCRIPTOR)(Offset >> 32);
    //     Sys_Cpu_GdtSetIsPresent(TRUE, Desc);
    //     Sys_Cpu_GdtSetRing(Ring, Desc);
    //     Sys_Cpu_GdtSetType(CPU_SEGMENT_TYPE_CALL_GATE, Desc);
    // }


    // static inline void SYSABI Sys_Cpu_LdtSetLimit(IN const UINT32 Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     Sys_Cpu_GdtSetLimit(Val, Desc);
    // }
    // static inline void SYSABI Sys_Cpu_LdtSetBase(IN const UINT64 Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     Sys_Cpu_GdtSetBase((UINT32)Val & 0xFFFFFFFF, Desc);
    //     *(Desc+sizeof(CPU_SEGMENT_DESCRIPTOR)) = 0;
    //     *(Desc+sizeof(CPU_SEGMENT_DESCRIPTOR)) |= (CPU_SEGMENT_DESCRIPTOR)(Val >> 32);
    // }
    // static inline void SYSABI Sys_Cpu_LdtSetType(IN const UINT8 Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     Sys_Cpu_GdtSetType(Val, Desc);
    // }
    // static inline void SYSABI Sys_Cpu_LdtSetRing(IN const UINT8 Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     Sys_Cpu_GdtSetRing(Val, Desc);
    // }
    // static inline void SYSABI Sys_Cpu_LdtSetIsPresent(IN const BOOLEAN Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     Sys_Cpu_GdtSetIsPresent(Val, Desc);
    // }
    // static inline void SYSABI Sys_Cpu_LdtSetIsAvailable(IN const BOOLEAN Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     Sys_Cpu_GdtSetIsAvailable(Val, Desc);
    // }
    // static inline void SYSABI Sys_Cpu_LdtSetSysFlags(IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     Sys_Cpu_GdtSetIsSys(TRUE, Desc);
    //     Sys_Cpu_GdtSetIsDbSet(FALSE, Desc);
    //     Sys_Cpu_GdtSetIsLongSet(FALSE, Desc);
    // }
    // static inline void SYSABI Sys_Cpu_LdtSetIs4kGranularity(IN const BOOLEAN Val, IN OUT CPU_SEGMENT_DESCRIPTOR *Desc) {
    //     Sys_Cpu_GdtSetIs4kGranularity(Val, Desc);
    // }



#else
    #error "cpu: platform not supported"
#endif


//
// Forward declarations.
//

static SYS_STATUS SYSABI Sys_Cpu_InitTss(IN OUT SYS_MEMORY_TABLE *MemoryTable,
    IN const BOOLEAN EnableTaskIo, IN OUT CPU_TSS *Tss, IN const UINTN DebugPort);
static SYS_STATUS SYSABI Sys_Cpu_InitGdt(IN OUT SYS_MEMORY_TABLE *MemoryTable, IN const UINTN GdtAddr,
    IN const SYS_CPU_GDT_PARAMS *Params, IN const UINTN DebugPort);

//
// Exported functions.
//

SYS_STATUS SYSABI Sys_Cpu_Cli() {
#if defined(__amd64__) || defined(__x86__)
    __asm__("cli");
    return SYS_STATUS_SUCESS;

#else
    return SYS_STATUS_FAIL;
#endif
}


SYS_STATUS SYSABI Sys_Cpu_Sti() {
#if defined(__amd64__) || defined(__x86__)
    __asm__("sti");
    return SYS_STATUS_SUCESS;

#else
    return SYS_STATUS_FAIL;
#endif
}


static inline const UINTN SYSABI Sys_Cpu_GetGdtSize() {
    return CPU_MAX_NUM_SEGMENTS * sizeof(CPU_SEGMENT_DESCRIPTOR);
}



void SYSABI Sys_Cpu_InitSegmentDescriptors(IN OUT SYS_MEMORY_TABLE *MemoryTable, IN const SYS_CPU_GDT_PARAMS *GdtParams,
    OUT SYS_CPU_MEMORY_SEGMENTS *Segments, IN const UINTN DebugPort) {
    
    // Find space for the Gdt within the given regions and get its address.
    UINTN GdtSize = Sys_Cpu_GetGdtSize();
    UINTN NumGdtPages = ALIGN_VALUE(GdtSize, SYS_MEMORY_PAGE_SIZE) / SYS_MEMORY_PAGE_SIZE;
    void *GdtMem = Sys_Memory_AllocPages(MemoryTable, NumGdtPages, FALSE, NULL, DebugPort);
    if (GdtMem == NULL) {
        PANIC_EXIT("cpu: failed to alloc gdt memory", SYS_STATUS_FAIL, DebugPort);
        return;
    }

    SYS_STATUS Status;

    CPU_TSS Tss;
    Status = Sys_Cpu_InitTss(MemoryTable, TRUE, &Tss, DebugPort);
    if (SYS_IS_ERROR(Status)) {
        PANIC_EXIT("cpu: failed to init task segments", SYS_STATUS_FAIL, DebugPort);
        return;
    }
    SYS_SERIAL_LOG("cpu: initialized task segments\n", DebugPort);

    
    // Initialize the descriptor tables.
    Status = Sys_Cpu_InitGdt(MemoryTable, (UINTN) GdtMem, GdtParams, DebugPort);
    if (SYS_IS_ERROR(Status)) {
        PANIC_EXIT("cpu: failed to init gdt", SYS_STATUS_FAIL, DebugPort);
        return;
    }

}

// Finds and reserves memory for a new TSS and initializes it, populating the Tss structure on success.
static SYS_STATUS SYSABI Sys_Cpu_InitTss(IN OUT SYS_MEMORY_TABLE *MemoryTable,
    IN const BOOLEAN EnableTaskIo, IN OUT CPU_TSS *Tss, IN const UINTN DebugPort) {

    // Compute the required TSS size.
    UINTN IoMapSize = (CPU_MAX_NUM_IO_PORTS >> 3) + 1; // Must have an extra all-1s byte at the end.
    UINTN RequiredTssSize = ALIGN_VALUE(CPU_TSS_HEADER_SIZE, CPU_STACK_ALIGNMENT) + IoMapSize;
    RequiredTssSize = ALIGN_VALUE(RequiredTssSize, SYS_MEMORY_PAGE_SIZE);
    RequiredTssSize += ALIGN_VALUE(SYS_RING0_STACK_SIZE, SYS_MEMORY_PAGE_SIZE);
    RequiredTssSize += ALIGN_VALUE(SYS_RING1_STACK_SIZE, SYS_MEMORY_PAGE_SIZE);
    RequiredTssSize += ALIGN_VALUE(SYS_RING2_STACK_SIZE, SYS_MEMORY_PAGE_SIZE);

    // Find space for the TSS within the regions and get its address.
    UINTN NumTssPages = RequiredTssSize / SYS_MEMORY_PAGE_SIZE;
    void *TssMem = Sys_Memory_AllocPages(MemoryTable, NumTssPages, FALSE, NULL, DebugPort);
    if (TssMem == NULL) {
        return SYS_STATUS_FAIL;
    }
    UINTN TssAddr = (UINTN) TssMem;

    UINTN IoMapAddr = ALIGN_VALUE(TssAddr + CPU_TSS_HEADER_SIZE, CPU_STACK_ALIGNMENT);
    UINTN Ring0StackAddr = ALIGN_VALUE(IoMapAddr + IoMapSize, SYS_MEMORY_PAGE_SIZE);
    UINTN Ring1StackAddr = ALIGN_VALUE(Ring0StackAddr + SYS_RING0_STACK_SIZE, SYS_MEMORY_PAGE_SIZE);
    UINTN Ring2StackAddr = ALIGN_VALUE(Ring1StackAddr + SYS_RING1_STACK_SIZE, SYS_MEMORY_PAGE_SIZE);

    // Populate the TSS structure (Intel 7.7 Vol 3A) and assign Tss output parameter values.
    UINT8 *TssPtr;
    Tss->TssAddr = TssAddr;
    Tss->TssSize = RequiredTssSize;
    TssPtr = (UINT8*) TssAddr;
    *((UINT32*)TssPtr) = (UINT32) 0;
    TssPtr += 4;
    Tss->Stack0Field = (UINT64*)TssPtr;
    TssPtr += 8;
    *Tss->Stack0Field = (UINT64) Ring0StackAddr;
    Tss->Stack0Memory = (void*) Ring0StackAddr;
    Tss->Stack1Field = (UINT64*)TssPtr;
    TssPtr += 8;
    *Tss->Stack1Field = (UINT64) Ring1StackAddr;
    Tss->Stack1Memory = (void*) Ring1StackAddr;
    Tss->Stack2Field = (UINT64*)TssPtr;
    TssPtr += 8;
    *Tss->Stack2Field = (UINT64) Ring2StackAddr;
    Tss->Stack2Memory = (void*) Ring2StackAddr;
    *((UINT64*)TssPtr) = (UINT64) 0;
    TssPtr += 8;
    for (UINTN i=0; i < 7; i++) {  // Zero all 7 IST pointers.
        Tss->IstFields[i] = (UINT64*)TssPtr;
        TssPtr += 8;
        *Tss->IstFields[i] = (UINT64) 0;
        Tss->IstMemory[i] = NULL;
    }
    *((UINT64*)TssPtr) = (UINT64) 0;
    TssPtr += 8;
    *((UINT16*)TssPtr) = (UINT16) 0;
    TssPtr += 2;
    Tss->IoMapField = (UINT16*)TssPtr;
    *Tss->IoMapField = (UINT16) (IoMapAddr - TssAddr);
    Tss->IoMapMemory = (void*) IoMapAddr;

    // Zero or fill the IO permissions bitmap depending on whether IO is enabled.
    UINT8 IoVal = 0;
    if (EnableTaskIo) {
        IoVal = 0xFF;
    }
    TssPtr = (UINT8*) IoMapAddr;
    for (UINTN i=0; i < IoMapSize-1; i++, TssPtr++) {
        *TssPtr = IoVal;
    }
    *TssPtr = 0xFF;  // Last byte is always all 1s.

    return SYS_STATUS_SUCESS;
}




static inline UINTN Sys_Cpu_GdtSegInitNone(CPU_SEGMENT_DESCRIPTOR *SegmentDesc) {
    Sys_Cpu_GdtSetBase(0, SegmentDesc);
    Sys_Cpu_GdtSetLimit(0, SegmentDesc);
    Sys_Cpu_GdtSetRing(0, SegmentDesc);
    Sys_Cpu_GdtSetType(0, SegmentDesc);
    Sys_Cpu_GdtSetIsSys(FALSE, SegmentDesc);
    Sys_Cpu_GdtSetIsPresent(FALSE, SegmentDesc);
    Sys_Cpu_GdtSetIsAvailable(FALSE, SegmentDesc);
    Sys_Cpu_GdtSetIsDbSet(FALSE, SegmentDesc);
    Sys_Cpu_GdtSetIsLongSet(FALSE, SegmentDesc);
    Sys_Cpu_GdtSetIs4kGranularity(FALSE, SegmentDesc);
    return sizeof(CPU_SEGMENT_DESCRIPTOR);
}

static inline UINTN Sys_Cpu_GdtSegInitSysCode(UINT8 Ring, UINT32 Addr, UINT32 NumPages, CPU_SEGMENT_DESCRIPTOR *SegmentDesc) {
    Sys_Cpu_GdtSetBase(Addr, SegmentDesc);
    Sys_Cpu_GdtSetLimit(NumPages, SegmentDesc);
    Sys_Cpu_GdtSetRing(Ring, SegmentDesc);
    Sys_Cpu_GdtSetType(CPU_SEGMENT_TYPE_CODE, SegmentDesc);
    Sys_Cpu_GdtSetIsSys(FALSE, SegmentDesc);
    Sys_Cpu_GdtSetIsPresent(TRUE, SegmentDesc);
    Sys_Cpu_GdtSetIsAvailable(TRUE, SegmentDesc);
    Sys_Cpu_GdtSetIsDbSet(FALSE, SegmentDesc);
    Sys_Cpu_GdtSetIsLongSet(TRUE, SegmentDesc);
    Sys_Cpu_GdtSetIs4kGranularity(TRUE, SegmentDesc);
    return sizeof(CPU_SEGMENT_DESCRIPTOR);
}


// Initializes the global descriptor table for the given system memory regions at the specified address.
static SYS_STATUS SYSABI Sys_Cpu_InitGdt(IN OUT SYS_MEMORY_TABLE *MemoryTable, IN const UINTN GdtAddr,
    IN const SYS_CPU_GDT_PARAMS *Params, IN const UINTN DebugPort) {

    
    CPU_SEGMENT_DESCRIPTOR *SegmentDesc = (CPU_SEGMENT_DESCRIPTOR*) GdtAddr;
    SegmentDesc += Sys_Cpu_GdtSegInitNone(SegmentDesc);

    // TODO
    SegmentDesc += Sys_Cpu_GdtSegInitSysCode(0, 0, 0, SegmentDesc);

    SYS_SERIAL_LOG("TODO...\n", DebugPort);





    // if (GdtAddr != (GdtAddr & 0xFFFFFFFF)) {
    //     // Gdt base addresses must be 32-bit.
    //     SYS_SERIAL_LOG("cpu: gdt region too high\n", DebugPort);
    //     return SYS_STATUS_FAIL;
    // }

    // for (UINTN i=0; i < CPU_MAX_NUM_SEGMENTS; i++, CurSegment += sizeof(CPU_SEGMENT_DESCRIPTOR)) {

    // }


    return SYS_STATUS_SUCESS;
}

