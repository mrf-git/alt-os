#ifndef SYS_BOOT_OS_INIT_H
#define SYS_BOOT_OS_INIT_H

#include "Conf.h"
#include "FrameBuffer.h"
#include "Resources.h"


//
// Structure for holding variables and references need to boot the 0 process.
//
typedef struct {

    // Boot.
    SYS_BOOT_CONF *Conf;                    // Pointer to the system configuration used to boot the system.
    SYS_BOOT_RESOURCES *Resources;          // Pointer to the respacked resources used to boot the system.

    // Memory.
    const UINTN BootStackStartAddr;         // Start address of the boot stack memory.
    const UINTN BootStackEndAddr;           // Final address of the boot stack memory.
    UINTN BootStackTipAddr;                 // The address of the current boot stack tip.
    void *MemoryMap;                        // Pointer to an array of EFI_MEMORY_DESCRIPTOR structures describing the system memory map at boot.
    UINTN MemoryMapSize;                    // The size in bytes of the memory map.
    UINTN MemoryMapDescriptorSize;          // The size in bytes of each memory map descriptor.

    // Graphics.
    SYS_FRAMEBUFFER *FrameBuffer;           // The graphical framebuffer for rendering, if graphics not off in system configuration.

    // SMBIOS.
    UINTN SmbiosVersion;                    // The supported SMBIOS version, either 1 or 3. 0 means SMBIOS not supported.
    void *SmbiosEntry;                      // Pointer to the SMBIOS entry structure if SmbiosVersion is 1, else the SMBIOS3 entry structure.

    // ACPI.        
    UINTN AcpiRevision;                     // The supported ACPI revision, either 1 or 2. 0 means ACPI not supported.
    void *Rsdp;                             // Pointer to the ACPI Root System Description Pointer structure.

    // UEFI.        
    const CHAR8 *SysConfBootPath;           // Pointer to the utf8-encoded path of the SYS.CONF sourced for BootConf, or NULL if not loaded.
    const CHAR8 *EfiBootPath;               // Pointer to the utf8-encoded path of the EFI image that was booted.
    const CHAR8 *FirmwareVendor;            // Pointer to the utf8-encoded vendor string reported by the firmware.
    UINTN FirmwareRevision;                 // The revision number reported by the firmware.
    void *FirmwareVendorGuid;               // Pointer to the vendor EFI_GUID structure reported by the firmware.
    void *RS;                               // Pointer to the UEFI Runtime Services table.
    void *BS;                               // Pointer to the UEFI Boot Services table, or NULL after Boot Services are exited.
    void *Lip;                              // Pointer to the UEFI Loaded Image Protocol, or NULL after Boot Services are exited.

} SYS_BOOT_CONTEXT;


//
// Allocates the requested memory from the boot stack in the specified boot context, or NULL on error.
//
void * SYSABI Sys_Init_AllocBoot(IN OUT SYS_BOOT_CONTEXT *Context, IN const UINTN Size, IN const UINTN Align);


//
// Called after exiting UEFI boot services to initialize the OS runtime. It can never return, and panics on failure.
//
void SYSABI Sys_Init(IN SYS_BOOT_CONTEXT *Context);


#endif // SYS_BOOT_OS_INIT_H
