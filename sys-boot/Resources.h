#ifndef SYS_BOOT_RESOURCES_H
#define SYS_BOOT_RESOURCES_H

#include "Common.h"


//
// Type declarations.
//

// SYS_BOOT_RESOURCES structure stores information about resources needed to boot the OS.
typedef struct {

    // TODO

} SYS_BOOT_RESOURCES;


//
// Loads all the packed resource data into the given Resources structure.
//
SYS_STATUS SYSABI Sys_Boot_Resources(IN OUT SYS_BOOT_RESOURCES *Resources);


#endif // SYS_BOOT_RESOURCES_H
