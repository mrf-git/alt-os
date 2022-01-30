#include <Uefi.h>
#include <Library/BaseLib.h>
#include <Library/BaseMemoryLib.h>
#include <Library/UefiLib.h>
#include <Library/UefiApplicationEntryPoint.h>
#include <Library/BaseUcs2Utf8Lib.h>
#include <Library/MemoryAllocationLib.h>
#include <Protocol/LoadedImage.h>
#include <Protocol/GraphicsOutput.h>

#include "Common.h"
#include "Serial.h"

#define NUM_DESCRIPTORS 1000


//
// The UEFI Application entry point.
//
EFI_STATUS EFIAPI Sys_Boot_Entry(IN EFI_HANDLE ImageHandle, IN EFI_SYSTEM_TABLE* SystemTable) {

    EFI_STATUS Status;

    // Attempt to get the memory map from the boot services and immediately exit boot services. If this fails a few
    // times in a row, return failure status.
    EFI_MEMORY_DESCRIPTOR MemMap[NUM_DESCRIPTORS];
    UINTN MaxMemMapSize = sizeof(EFI_MEMORY_DESCRIPTOR) * NUM_DESCRIPTORS;
    UINTN MemMapSize = MaxMemMapSize;
    UINTN MemMapKey;
    UINTN MemMapDescSize;
    UINT32 MemMapDescVersion;

    for (UINTN Tries=0;;){
        Status = SystemTable->BootServices->GetMemoryMap(&MemMapSize, &MemMap[0], &MemMapKey, &MemMapDescSize, &MemMapDescVersion);
        if (EFI_ERROR(Status)){
            return Status;
        }
        if (MemMapSize > MaxMemMapSize){
            Status = EFI_BUFFER_TOO_SMALL;
            return Status;
        }

        Status = SystemTable->BootServices->ExitBootServices(ImageHandle, MemMapKey);
        if (Status == EFI_INVALID_PARAMETER && Tries < 3){
            Tries++;
            continue;
        } else if (EFI_ERROR(Status)){
            return Status;
        } else {
            // Successfully exited boot services.
            break;
        }
    }


    Sys_Serial_Reset();
    SYS_SERIAL_LOG("*ok* booting\r\n");

    while (TRUE) {}


    return EFI_SUCCESS;
}

