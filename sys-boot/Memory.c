#include "Common.h"
#include "Memory.h"


void * SYSABI Sys_Memory_AllocPages(IN OUT SYS_MEMORY_TABLE *MemoryTable, IN const UINTN NumPages,
    IN const BOOLEAN IsLow, OUT SYS_MEMORY_REGION **ReservedRegion, IN const UINTN DebugPort) {
    
    SYS_MEMORY_REGION *Region = NULL;
    UINTN MinGap = (UINTN) -1;
    for (UINTN i=0; i < MemoryTable->NumRegions; i++) {
        if (IsLow && MemoryTable->Regions[i].Addr > 0xFFFFFFFF) {
            break;
        }
        UINTN AvailablePages = MemoryTable->Regions[i].NumPages - MemoryTable->Regions[i].UsedPages;
        if (AvailablePages >= NumPages && !MemoryTable->Regions[i].IsPersistent && !MemoryTable->Regions[i].IsSpecial) {
            UINTN Gap = AvailablePages - NumPages;
            if (Gap < MinGap) {
                Region = &MemoryTable->Regions[i];
                MinGap = Gap;
            }
        }
    }
    if (Region == NULL) {
        PANIC_EXIT("memory: failed to alloc memory region", SYS_STATUS_FAIL, DebugPort);
        return NULL;
    }
    UINTN Addr = (UINTN) Region->Addr + Region->UsedPages * SYS_MEMORY_PAGE_SIZE;
    Region->UsedPages += NumPages;

    if (ReservedRegion != NULL) {
        *ReservedRegion = Region;
    }

    return (void*) Addr;
}


void * SYSABI Sys_Memory_AllocRoot(IN OUT SYS_MEMORY_TABLE *MemoryTable, IN const UINTN Size, IN const UINTN Align,
    IN const UINTN DebugPort) {
    UINTN AlignedTipAddr = ALIGN_VALUE(MemoryTable->RootStackTipAddr, Align);
    if (AlignedTipAddr + Size > MemoryTable->RootStackEndAddr){
        PANIC_EXIT("memory: failed to alloc root stack memory", SYS_STATUS_FAIL, DebugPort);
        return NULL;
    }
    MemoryTable->RootStackTipAddr = AlignedTipAddr + Size;
    return (void*) AlignedTipAddr;
}
