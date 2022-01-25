#ifndef SYS_BOOT_CONF_H
#define SYS_BOOT_CONF_H

#include <Uefi.h>
#include "Common.h"

#define SYS_BOOT_MAX_STACK_SIZE                 0x100000        // 1 MiB
#define SYS_BOOT_MAX_CONF_OPT_LEN               256


//
// System configuration structure with settings that affect how the machine boots.
//
typedef struct {
    BOOLEAN   IsGraphicsOff;  // If TRUE, graphical output is never enabled.
    UINT16    DebugPort;      // The serial port to use for debugging I/O, or 0 to disable.
    CHAR8     *OsInitString;  // The string sent to DebugPort to indicate that the loader has taken control of I/O and will now initialize the OS.
} SYS_BOOT_CONF;


//
// Default values.
//
#define SYS_BOOT_CONF_DEFAULT_IS_GRAPHICS_OFF    FALSE
#define SYS_BOOT_CONF_DEFAULT_DEBUG_PORT         0
#define SYS_BOOT_CONF_DEFAULT_OS_INIT_STRING     "\r\nOSINIT\r\n\r\n"


//
// Initializes copies of the device-dependent file paths for the booted image and system configuration file
// assuming that it's in the same directory.
//
EFI_STATUS SYSABI Sys_Conf_InitializePaths(OUT CHAR16 **ImagePath, OUT CHAR16 **SysConfPath, IN OUT void *Context);

//
// Attempts to initialize the system configuration for the boot context from the SYS.CONF file at the specified path.
// If the file cannot be opened, default configuration values are used and the SysConfPath is set to NULL. If the
// file can be opened but not parsed, an error status is returned.
//
EFI_STATUS SYSABI Sys_Conf_Initialize(IN OUT CHAR16 **SysConfPath, IN OUT void *Context);

//
// Processes the specified UEFI ConfigurationTable structure and sets the table locations and versions for the
// ACPI RSDP and SMBIOS entry point in the boot context, if found.
//
EFI_STATUS SYSABI Sys_Conf_ProcessUefiTable(IN void *ConfigurationTable,
    IN const UINTN NumberOfEntries, IN OUT void *Context);


#endif // SYS_BOOT_CONF_H
