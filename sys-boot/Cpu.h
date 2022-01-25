#ifndef SYS_BOOT_CPU_H
#define SYS_BOOT_CPU_H

#include "Common.h"
#include "Memory.h"


//
// Type enums.
//

// System segment types corresponding to their index in the global descriptor table.
typedef enum {
    CpuSegNone,
    CpuSegSysCodeRing0,
    CpuSegSysDataRing0,
    CpuSegSysRoDataRing0,
    CpuSegSysStackRing0,
    CpuSegSysCodeRing1,
    CpuSegSysDataRing1,
    CpuSegSysRoDataRing1,
    CpuSegSysStackRing1,
    CpuSegSysCodeRing2,
    CpuSegSysDataRing2,
    CpuSegSysRoDataRing2,
    CpuSegSysStackRing2,
    CpuSegSysCodeRing3,
    CpuSegSysDataRing3,
    CpuSegSysRoDataRing3,
    CpuSegSysStackRing3,
    CpuSegSysTss,
    CpuSegSysCallGate,
    CpuSegSysInterruptGate,
    CpuSegSysInterruptTrapGate,
    CpuSegSysTaskGate,

    _CpuSegMax,
} SYS_CPU_SEGMENT_TYPE;



//
// Structure definitions.
//

// Stores CPU memory segment information and pointers used by the OS.
typedef struct {
    SYS_MEMORY_REGION *GdtRegion;
    SYS_MEMORY_REGION *LowMemoryRegion;
    
} SYS_CPU_MEMORY_SEGMENTS;

// Parameters for GDT initialization.
typedef struct {
    UINTN SysCodeSizeRing0;
    UINTN SysDataSizeRing0;
    UINTN SysRoDataSizeRing0;
    UINTN SysCodeSizeRing1;
    UINTN SysDataSizeRing1;
    UINTN SysRoDataSizeRing1;
    UINTN SysCodeSizeRing2;
    UINTN SysDataSizeRing2;
    UINTN SysRoDataSizeRing2;
    UINTN SysCodeSizeRing3;
    UINTN SysDataSizeRing3;
    UINTN SysRoDataSizeRing3;

} SYS_CPU_GDT_PARAMS;



//
// Disables interrupts.
//
SYSEXPORT SYS_STATUS SYSABI Sys_Cpu_Cli();


//
// Enables interrupts.
//
SYSEXPORT SYS_STATUS SYSABI Sys_Cpu_Sti();


//
// Initializes the descriptor tables for CPU memory segments of the system memory regions and populates the
// given segments structure with outputs. Panics on failure.
//
void SYSABI Sys_Cpu_InitSegmentDescriptors(IN OUT SYS_MEMORY_TABLE *MemoryTable, IN const SYS_CPU_GDT_PARAMS *GdtParams,
    OUT SYS_CPU_MEMORY_SEGMENTS *Segments, IN const UINTN DebugPort);


#endif // SYS_BOOT_CPU_H
