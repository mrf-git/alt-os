#include <Uefi.h>
#include <Uefi/UefiSpec.h>

#include "Conf.h"
#include "OsInit.h"
#include "Elf.h"
#include "Memory.h"
#include "Cpu.h"


#define OS_INIT_INTERP      "boot"


//
// Structure definitions.
//

typedef struct {
    SYS_ELF_LIB *LibCommon;
    const CHAR8 *LibCommonName;
    SYS_ELF_LIB *LibCpu;
    const CHAR8 *LibCpuName;
    SYS_ELF_LIB *LibFrameBuffer;
    const CHAR8 *LibFrameBufferName;
    SYS_ELF_LIB *LibL0;
    const CHAR8 *LibL0Name;
    SYS_ELF_LIB *LibL1;
    const CHAR8 *LibL1Name;
    SYS_ELF_LIB *LibL2;
    const CHAR8 *LibL2Name;
    SYS_ELF_LIB *LibL3;
    const CHAR8 *LibL3Name;
    SYS_ELF_LIB *LibLink;
    const CHAR8 *LibLinkName;
    SYS_ELF_LIB *LibMemory;
    const CHAR8 *LibMemoryName;
} INIT_ELF_TABLE;



//
// Forward declarations.
//

static UINTN SYSABI Sys_Init_FindRootStack(IN SYS_BOOT_CONTEXT *Context);
static void SYSABI Sys_Init_VirtualMemoryInit(IN const UINTN RootStackBaseAddr, IN OUT SYS_BOOT_CONTEXT *Context,
    OUT SYS_MEMORY_TABLE **MemoryTable);


//
// Macros.
//

// Allocate a new reference to Type from the boot stack.
#define NEW_S_BOOT(Type,Context) \
    ((Type*) Sys_Init_AllocBoot(Context, sizeof(Type), CPU_STACK_ALIGNMENT))

// Allocate a new reference to Type from the root stack.
#define NEW_S_ROOT(Type,MemoryTable,DebugPort) \
    ((Type*) Sys_Memory_AllocRoot(MemoryTable, sizeof(Type), CPU_STACK_ALIGNMENT, DebugPort))

// Reads the ELF library with the specified name from the boot context resources into the ElfTable.
#define READ_ELF_LIB(ElfTable,LibName,Context) \
    ElfTable->LibName = Sys_Elf_ReadLib(Context->Resources->LibName, Context->Resources->LibName ## Size, Context); \
    ElfTable->LibName ## Name = Context->Resources->LibName ## LogicalName; \
    Sys_Init_PrintMemorySizes(ElfTable->LibName ## Name, ElfTable->LibName, Context)


//
// Exported functions.
//
   
void SYSABI Sys_Init(IN SYS_BOOT_CONTEXT *Context) {

    SYS_SERIAL_LOG("init: starting OsInit\n", Context->Conf->DebugPort);

    if (Context->Conf->IsGraphicsOff) {
        SYS_SERIAL_LOG("init: graphics off\n", Context->Conf->DebugPort);
    }

    // Find and reserve root stack space and initialize virtual memory there.
    SYS_MEMORY_TABLE *MemoryTable;
    UINTN RootStackAddr = Sys_Init_FindRootStack(Context);
    Sys_Init_VirtualMemoryInit(RootStackAddr, Context, &MemoryTable);

    // Reserve space on the root stack for the global symbol table hash table.
    SYS_HASH_TABLE *Htable = (SYS_HASH_TABLE*) Sys_Memory_AllocRoot(MemoryTable, Sys_Htable_Size(),
        CPU_STACK_ALIGNMENT, Context->Conf->DebugPort);
    if (Htable == NULL) {
        PANIC_EXIT("init: failed to alloc hash table", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return;
    }
    Sys_Htable_Init(Htable);




    // SYS_CPU_GDT_PARAMS *GdtParams = NEW_S_BOOT(SYS_CPU_GDT_PARAMS, Context);
    // // TODO fill in params


    // SYS_CPU_MEMORY_SEGMENTS *Segments = NEW_S_ROOT(SYS_CPU_MEMORY_SEGMENTS, MemoryTable, Context->Conf->DebugPort);
    // Sys_Cpu_InitSegmentDescriptors(MemoryTable, GdtParams, Segments, Context->Conf->DebugPort);

    
    SYS_SERIAL_LOG("init: everything ok\n", Context->Conf->DebugPort);
    


    PANIC_EXIT("OK", SYS_STATUS_SUCESS, Context->Conf->DebugPort);
}


void * SYSABI Sys_Init_AllocBoot(IN OUT SYS_BOOT_CONTEXT *Context, IN const UINTN Size, IN const UINTN Align) {
    UINTN AlignedTipAddr = ALIGN_VALUE(Context->BootStackTipAddr, Align);
    if (AlignedTipAddr + Size > Context->BootStackEndAddr){
        return NULL;
    }
    Context->BootStackTipAddr = AlignedTipAddr + Size;
    return (void*) AlignedTipAddr;
}


// Scans through the physical memory map reported by UEFI and identifies a page-aligned address of
// conventional memory to use as the root stack. The root stack grows upwards before entering the OS,
// and persists in memory along with the OS. Panics on failure.
static UINTN SYSABI Sys_Init_FindRootStack(IN SYS_BOOT_CONTEXT *Context) {
    
    EFI_PHYSICAL_ADDRESS RootStackStartAddress = 0;
    UINTN MinGap = (UINTN) -1;
    UINTN NumRootStackPages = ALIGN_VALUE(SYS_ROOT_STACK_SIZE, SYS_MEMORY_PAGE_SIZE) / SYS_MEMORY_PAGE_SIZE;
    for (UINTN Offset=0; Offset < Context->MemoryMapSize; Offset += Context->MemoryMapDescriptorSize) {
        EFI_MEMORY_DESCRIPTOR *RegionDesc = (EFI_MEMORY_DESCRIPTOR*) (Context->MemoryMap + Offset);
        if (!RegionDesc->NumberOfPages || RegionDesc->PhysicalStart > SYS_MEMORY_MAX_START) {
            PANIC_EXIT("init: bad memory map descriptor", SYS_STATUS_FAIL, Context->Conf->DebugPort);
            return 0;
        }

        BOOLEAN IsRootable = FALSE;
        switch (RegionDesc->Type) {
            case EfiBootServicesCode:
            case EfiBootServicesData:
            case EfiConventionalMemory:
                IsRootable = (RegionDesc->PhysicalStart > 0
                    && (RegionDesc->Attribute & EFI_MEMORY_RUNTIME) != EFI_MEMORY_RUNTIME
                    && (RegionDesc->Attribute & EFI_MEMORY_NV) != EFI_MEMORY_NV
                    && (RegionDesc->Attribute & EFI_MEMORY_SP) != EFI_MEMORY_SP);
                break;
            default:
                continue;
        }

        if (IsRootable && RegionDesc->NumberOfPages >= NumRootStackPages) {
            UINTN Gap = RegionDesc->NumberOfPages - NumRootStackPages;
            if (Gap < MinGap) {
                RootStackStartAddress = RegionDesc->PhysicalStart;
                MinGap = Gap;
            }
        }
    }

    if (!RootStackStartAddress) {
        PANIC_EXIT("init: no root stack available", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return 0;
    }

    return RootStackStartAddress;
}


// Initializes the virtual memory for the system to all identity-mapped pages, or panics on failure.
// Consumes from the root stack and updates the MemoryTable with current memory information.
static void SYSABI Sys_Init_VirtualMemoryInit(IN const UINTN RootStackBaseAddr, IN OUT SYS_BOOT_CONTEXT *Context,
    OUT SYS_MEMORY_TABLE **MemoryTable) {
    
    // Allocate space for the MemoryTable structrure before region data.
    SYS_MEMORY_TABLE *NewMemoryTable = (SYS_MEMORY_TABLE*) ALIGN_VALUE(RootStackBaseAddr, CPU_STACK_ALIGNMENT);
    UINTN RegionStartAddr = (UINTN) NewMemoryTable + sizeof(SYS_MEMORY_TABLE);

    // Initialize an array of SYS_MEMORY_REGION structures big enough to hold all regions that stores the regions
    // usable by the OS, and another array of pointers to regions that can be freed after the OS is done initializing,
    // with no space reserved yet.
    SYS_MEMORY_REGION *AllSysRegions = (SYS_MEMORY_REGION*) ALIGN_VALUE(RegionStartAddr, CPU_STACK_ALIGNMENT);
    UINTN RootStackTipAddr = (UINTN) AllSysRegions + sizeof(SYS_MEMORY_REGION) * (Context->MemoryMapSize / Context->MemoryMapDescriptorSize);
    SYS_MEMORY_REGION **FreeableSysRegions = (SYS_MEMORY_REGION**) ALIGN_VALUE(RootStackTipAddr, CPU_STACK_ALIGNMENT);
    UINTN NumAllSysRegions = 0;
    UINTN NumFreeableSysRegions = 0;

    // Loop over all EFI-reported regions, identity-map all pages, and find what can be used by the OS.
    SYS_MEMORY_REGION *SysRegionRoot = NULL;
    for (UINTN Offset=0; Offset < Context->MemoryMapSize; Offset += Context->MemoryMapDescriptorSize) {
        EFI_MEMORY_DESCRIPTOR *RegionDesc = (EFI_MEMORY_DESCRIPTOR*) (Context->MemoryMap + Offset);
        BOOLEAN IsPersistent = FALSE;
        SYS_MEMORY_REGION *CurSysRegion = NULL;
        UINTN UsedPages = 0;
        RegionDesc->VirtualStart = RegionDesc->PhysicalStart;

        switch (RegionDesc->Type) {
            default:
                SYS_SERIAL_LOG("init: memory type ", Context->Conf->DebugPort);
                SYS_SERIAL_LOG_INT(RegionDesc->Type, TRUE, Context->Conf->DebugPort);
                SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
                PANIC_EXIT("init: memory type unknown", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                return;
            
            case EfiReservedMemoryType:
            case EfiUnusableMemory:
                // Ignore.
                break;

            case EfiACPIMemoryNVS:
            case EfiMemoryMappedIO:
            case EfiMemoryMappedIOPortSpace:
            case EfiPalCode:
            case EfiRuntimeServicesCode:
            case EfiRuntimeServicesData:
                // Ignore.
                break;

            case EfiLoaderCode:
            case EfiLoaderData:
                CurSysRegion = &AllSysRegions[NumAllSysRegions++];
                UsedPages = RegionDesc->NumberOfPages;
                FreeableSysRegions[NumFreeableSysRegions++] = CurSysRegion;
                break;

            case EfiACPIReclaimMemory:
                CurSysRegion = &AllSysRegions[NumAllSysRegions++];
                UsedPages = RegionDesc->NumberOfPages;
                FreeableSysRegions[NumFreeableSysRegions++] = CurSysRegion;
                break;

            case EfiPersistentMemory: IsPersistent = TRUE; // Fall through.
            case EfiBootServicesCode:
            case EfiBootServicesData:
            case EfiConventionalMemory:
                if((RegionDesc->Attribute & EFI_MEMORY_RUNTIME) == EFI_MEMORY_RUNTIME) {
                    // Ignore (shouldn't happen).
                    break;
                }
                if((RegionDesc->Attribute & EFI_MEMORY_NV) == EFI_MEMORY_NV) {
                    IsPersistent = TRUE;
                }
                CurSysRegion = &AllSysRegions[NumAllSysRegions++];
                if (RegionDesc->PhysicalStart == RootStackBaseAddr){
                    SysRegionRoot = CurSysRegion;
                    UsedPages = ALIGN_VALUE(SYS_ROOT_STACK_SIZE, SYS_MEMORY_PAGE_SIZE) / SYS_MEMORY_PAGE_SIZE;
                }
                break;
        }

        if (CurSysRegion != NULL) {
            // Region is usable, so initialize it.
            CurSysRegion->Addr = RegionDesc->VirtualStart;
            CurSysRegion->NumPages = RegionDesc->NumberOfPages;
            CurSysRegion->UsedPages = UsedPages;
            CurSysRegion->IsPersistent = IsPersistent;
            CurSysRegion->IsSpecial = (RegionDesc->Attribute & EFI_MEMORY_SP) == EFI_MEMORY_SP;
            CurSysRegion->SupportsUncacheable = (RegionDesc->Attribute & EFI_MEMORY_UC) == EFI_MEMORY_UC;
            CurSysRegion->SupportsWriteCombining = (RegionDesc->Attribute & EFI_MEMORY_WC) == EFI_MEMORY_WC;
            CurSysRegion->SupportsWriteThrough = (RegionDesc->Attribute & EFI_MEMORY_WT) == EFI_MEMORY_WT;
            CurSysRegion->SupportsWriteBack = (RegionDesc->Attribute & EFI_MEMORY_WB) == EFI_MEMORY_WB;
            CurSysRegion->Index = NumAllSysRegions - 1;
            CurSysRegion->MapOffset = Offset;

            // Verify that this region doesn't overlap the previous region.
            if (NumAllSysRegions > 1) {
                SYS_MEMORY_REGION *PrevSysRegion = &AllSysRegions[NumAllSysRegions-2];
                if (CurSysRegion->Addr < PrevSysRegion->Addr + PrevSysRegion->NumPages * SYS_MEMORY_PAGE_SIZE) {
                    PANIC_EXIT("init: overlapping memory", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                    return;
                }
            }
        }
    }

    if (NumAllSysRegions == 0 || SysRegionRoot == NULL) {
        PANIC_EXIT("init: bad virtual memory", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return;
    }

    // Make sure the page at address 0 is reserved.
    if (AllSysRegions->Addr == 0 && AllSysRegions->UsedPages == 0) {
        AllSysRegions->UsedPages++;
    }

    // Enable virtual memory mode in EFI runtime.
    EFI_RUNTIME_SERVICES *RS = (EFI_RUNTIME_SERVICES*) Context->RS;
    EFI_STATUS EfiStatus = RS->SetVirtualAddressMap(Context->MemoryMapSize, Context->MemoryMapDescriptorSize,
        EFI_MEMORY_DESCRIPTOR_VERSION, Context->MemoryMap);
    if (EFI_ERROR(EfiStatus)) {
        PANIC_EXIT("init: failed to initialize virtual memory", EfiStatus, Context->Conf->DebugPort);
        return;
    }

    // Populate the output memory table and consume virtual memory structures from the root stack.
    NewMemoryTable->Regions = AllSysRegions;
    *((UINTN*)&NewMemoryTable->NumRegions) = NumAllSysRegions;
    NewMemoryTable->FreeableRegions = FreeableSysRegions;
    NewMemoryTable->NumFreeableRegions = NumFreeableSysRegions;
    *((UINTN*)&NewMemoryTable->RootStackStartAddr) = RootStackBaseAddr;
    *((UINTN*)&NewMemoryTable->RootStackEndAddr) = RootStackBaseAddr + SYS_ROOT_STACK_SIZE;
    NewMemoryTable->RootStackTipAddr = (UINTN) FreeableSysRegions + sizeof(SYS_MEMORY_REGION*) * NumFreeableSysRegions;
    *MemoryTable = NewMemoryTable;
    
    SYS_SERIAL_LOG("init: initialized virtual memory\n", Context->Conf->DebugPort);
}


