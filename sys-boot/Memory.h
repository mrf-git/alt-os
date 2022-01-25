#ifndef SYS_BOOT_MEMORY_H
#define SYS_BOOT_MEMORY_H

#include "Common.h"


//
// Memory definitions.
//

#define SYS_MEMORY_PAGE_SIZE                4096
#define SYS_MEMORY_MAX_START                0xFFFFFFFFFFFFF000

#define SYS_ROOT_STACK_SIZE                 0x8000000       // 128 MiB

#define SYS_RING0_STACK_SIZE                0x400000        // 4 MiB
#define SYS_RING1_STACK_SIZE                0x400000        // 4 MiB
#define SYS_RING2_STACK_SIZE                0x800000        // 8 MiB
#define SYS_RING3_STACK_SIZE                0x1000000       // 16 MiB


//
// Forward declarations.
//

typedef struct _SYS_MEMORY_REGION SYS_MEMORY_REGION;
typedef struct _SYS_MEMORY_TABLE SYS_MEMORY_TABLE;


//
// Function pointer type definitions.
//

// Sys_Memory_AllocPages
typedef void * SYSABI (*SYS_FN_MALLOC_PAGES)(IN OUT SYS_MEMORY_TABLE *MemoryTable, IN const UINTN NumPages,
    IN const BOOLEAN IsLow, OUT SYS_MEMORY_REGION **ReservedRegion, IN const UINTN DebugPort);

// Sys_Memory_AllocRoot
typedef void * SYSABI (*SYS_FN_MALLOC_ROOT)(IN OUT SYS_MEMORY_TABLE *MemoryTable, IN const UINTN Size,
    IN const UINTN Align, IN const UINTN DebugPort);


//
// Structure definitions.
//

// Represents a multi-page region of memory useable by the OS.
typedef struct _SYS_MEMORY_REGION {
    UINTN Addr;                             // The virtual/physical address of the start of the region.
    UINTN NumPages;                         // The number of pages in the region.
    UINTN UsedPages;                        // The number of pages currently used in the region.
    BOOLEAN IsPersistent;                   // If TRUE, the region is non-volatile.
    BOOLEAN IsSpecial;                      // If TRUE, the region is special-purpose memory.
    BOOLEAN SupportsUncacheable;            // If TRUE, supports a cache configuration of uncacheable.
    BOOLEAN SupportsWriteCombining;         // If TRUE, supports a cache configuration of write-combining.
    BOOLEAN SupportsWriteThrough;           // If TRUE, supports a cache configuration of write-through.
    BOOLEAN SupportsWriteBack;              // If TRUE, supports a cache configuration of write-back.
    UINTN Index;                            // The index of this region within the array of system regions.
    UINTN MapOffset;                        // The offset of the start of this region descriptor in the EFI-reported memory map.

} SYS_MEMORY_REGION;

// Table of pointers and information for OS memory, initialized by the boot loader.
typedef struct _SYS_MEMORY_TABLE {
    SYS_MEMORY_REGION *Regions;             // Array of all memory regions useable by the OS.
    SYS_MEMORY_REGION **FreeableRegions;    // Array of pointers to post-init-freeable regions within Regions.
    const UINTN NumRegions;                 // Number of regions in Regions.
    UINTN NumFreeableRegions;               // Number of pointers in FreeableRegions.
    const UINTN RootStackStartAddr;         // Start address of the root stack memory.
    const UINTN RootStackEndAddr;           // Final address of the root stack memory.
    UINTN RootStackTipAddr;                 // The address of the current root stack tip.

} SYS_MEMORY_TABLE;


//
// Finds a contiguous region of pages within the system regions such that the smallest
// region of at least NumPages is selected. The region is then updated to reflect the reserved
// space in its UsedPages. If ReservedRegion is not NULL, a pointer to the selected region is set in it.
// Persistent and special memory regions are not considered. If IsLow is TRUE, only the first 4 GiB of
// memory regions are considered. Returns a pointer to the reserved region or panics on failure.
//
SYSEXPORT void * SYSABI Sys_Memory_AllocPages(IN OUT SYS_MEMORY_TABLE *MemoryTable, IN const UINTN NumPages,
    IN const BOOLEAN IsLow, OUT SYS_MEMORY_REGION **ReservedRegion, IN const UINTN DebugPort);


//
// Allocates Size number of bytes on the root stack, with address aligned at Align bytes, and returns a
// pointer to the new memory. Panics on failure.
//
SYSEXPORT void * SYSABI Sys_Memory_AllocRoot(IN OUT SYS_MEMORY_TABLE *MemoryTable, IN const UINTN Size,
    IN const UINTN Align, IN const UINTN DebugPort);


#endif // SYS_BOOT_MEMORY_H
