#include <Uefi.h>
#include <Library/BaseLib.h>
#include <Library/BaseMemoryLib.h>
#include <Library/UefiLib.h>
#include <Library/UefiApplicationEntryPoint.h>
#include <Library/BaseUcs2Utf8Lib.h>
#include <Library/MemoryAllocationLib.h>
#include <Protocol/LoadedImage.h>
#include <Protocol/GraphicsOutput.h>

#include "Conf.h"
#include "OsInit.h"
#include "FrameBuffer.h"


//
// Local static variables.
//

// Upward-growing memory reserve space used for booting, discarded after OS initialization.
static UINT8 StackMemoryReserve[SYS_BOOT_MAX_STACK_SIZE];


//
// Pre-runtime error macro.
//

#define ERROR_EXIT(Str,Status) \
    AsciiPrint("*error* "Str" (Status: %d)\r\n", Status); \
    return Status


//
// Forward declarations.
//

static const CHAR8 * SYSABI Sys_Boot_EntryUnicodeConvert(IN OUT SYS_BOOT_CONTEXT *Context, IN const CHAR16 *UnicodeString);
static EFI_STATUS SYSABI Sys_Boot_EntryInitializeFramebuffer(IN OUT SYS_BOOT_CONTEXT *Context);


//
// The UEFI Application entry point.
//
EFI_STATUS EFIAPI Sys_Boot_Entry(IN EFI_HANDLE ImageHandle, IN EFI_SYSTEM_TABLE* SystemTable) {

    EFI_STATUS Status;

    // Initialize a new boot context at the start of the reserve memory.
    SYS_BOOT_CONTEXT *Context = (SYS_BOOT_CONTEXT*) StackMemoryReserve;
    *((UINTN*)&Context->BootStackStartAddr) = (UINTN) &StackMemoryReserve[0];
    *((UINTN*)&Context->BootStackEndAddr) = (UINTN) &StackMemoryReserve[SYS_BOOT_MAX_STACK_SIZE-1];
    Context->BootStackTipAddr = Context->BootStackStartAddr + sizeof(SYS_BOOT_CONTEXT);
    Context->Conf = NULL;
    Context->Resources = NULL;
    Context->MemoryMap = NULL;
    Context->MemoryMapSize = 0;
    Context->MemoryMapDescriptorSize = 0;
    Context->FrameBuffer = NULL;
    Context->SmbiosVersion = 0;
    Context->SmbiosEntry = NULL;
    Context->AcpiRevision = 0;
    Context->Rsdp = NULL;
    Context->SysConfBootPath = NULL;
    Context->EfiBootPath = NULL;
    Context->FirmwareVendor = NULL;
    Context->FirmwareRevision = SystemTable->FirmwareRevision;
    Context->FirmwareVendorGuid = &SystemTable->ConfigurationTable->VendorGuid;
    Context->RS = SystemTable->RuntimeServices;
    Context->BS = SystemTable->BootServices;
    Context->Lip = NULL;

    Status = SystemTable->BootServices->HandleProtocol(ImageHandle, &gEfiLoadedImageProtocolGuid, &Context->Lip);
    if (EFI_ERROR(Status)){
        ERROR_EXIT("entry: failed to get Loaded Image Protocol", Status);
    }


    // Load the built-in boot resources.
    Context->Resources = (SYS_BOOT_RESOURCES*) Sys_Init_AllocBoot(Context, sizeof(SYS_BOOT_RESOURCES), CPU_STACK_ALIGNMENT);
    if (Context->Resources == NULL){
        ERROR_EXIT("entry: no memory to load boot resources", EFI_BUFFER_TOO_SMALL);
    }
    if (SYS_IS_ERROR(Sys_Boot_Resources(Context->Resources))){
        ERROR_EXIT("entry: failed to load boot resources", EFI_NOT_FOUND);
    }


    // Get the file paths of this image and the configuration file.
    CHAR16 *ImagePath = NULL;
    CHAR16 *SysConfPath = NULL;
    Status = Sys_Conf_InitializePaths(&ImagePath, &SysConfPath, Context);
    if (EFI_ERROR(Status) || ImagePath == NULL){
        ERROR_EXIT("entry: failed to initialize file paths", Status);
    }

    // Load the system configuration from the file at the initialized path if possible, otherwise set defaults and
    // clear the SysConfPath.
    Status = Sys_Conf_Initialize(&SysConfPath, Context);
    if (EFI_ERROR(Status)){
        ERROR_EXIT("entry: failed to initialize system configuration", Status);
    }
    
    // Copy Unicode strings to UTF8-encoded byte strings.
    Context->FirmwareVendor = Sys_Boot_EntryUnicodeConvert(Context, SystemTable->FirmwareVendor);
    if (Context->FirmwareVendor == NULL){
        ERROR_EXIT("entry: failed to read firmware vendor string", Status);
    }
    Context->EfiBootPath = Sys_Boot_EntryUnicodeConvert(Context, ImagePath);
    if (Context->FirmwareVendor == NULL){
        ERROR_EXIT("entry: failed to read image path string", Status);
    }
    if (SysConfPath) {
        Context->SysConfBootPath = Sys_Boot_EntryUnicodeConvert(Context, SysConfPath);
        if (Context->FirmwareVendor == NULL){
            ERROR_EXIT("entry: failed to read SYS.CONF path string", Status);
        }
    }


    // Locate important system tables from the UEFI configuration table.
    Status = Sys_Conf_ProcessUefiTable(SystemTable->ConfigurationTable, SystemTable->NumberOfTableEntries, Context);
    if (EFI_ERROR(Status)){
        ERROR_EXIT("entry: failed to process UEFI configuration", Status);
    }


    // Initialize framebuffer if graphics is not off.
    Status = Sys_Boot_EntryInitializeFramebuffer(Context);
    if (EFI_ERROR(Status)){
        ERROR_EXIT("entry: failed to initialize framebuffer", Status);
    }


    // Attempt to get the memory map from the boot services and immediately exit boot services. If this fails a few
    // times in a row, return failure status.
    UINTN NewTipAddr = ALIGN_VALUE(Context->BootStackTipAddr, CPU_STACK_ALIGNMENT);

    EFI_MEMORY_DESCRIPTOR *MemMap = (EFI_MEMORY_DESCRIPTOR*) NewTipAddr;
    UINTN MemMapSize = Context->BootStackEndAddr - NewTipAddr;
    UINTN MemMapKey;
    UINTN MemMapDescSize;
    UINT32 MemMapDescVersion;

    for (UINTN Tries=0;;){
        Status = SystemTable->BootServices->GetMemoryMap(&MemMapSize, &MemMap[0], &MemMapKey, &MemMapDescSize, &MemMapDescVersion);
        if (EFI_ERROR(Status)){
            ERROR_EXIT("entry: failed to get memory map", Status);
        }
        if (NewTipAddr + MemMapSize > Context->BootStackEndAddr){
            Status = EFI_BUFFER_TOO_SMALL;
            ERROR_EXIT("entry: memory map buffer too small", Status);
        }

        Status = SystemTable->BootServices->ExitBootServices(ImageHandle, MemMapKey);
        if (Status == EFI_INVALID_PARAMETER && Tries < 3){
            Tries++;
            continue;
        } else if (EFI_ERROR(Status)){
            ERROR_EXIT("entry: failed to exit boot services", Status);
        } else {
            // Successfully exited boot services.
            break;
        }
    }


    // Can no longer use Print without boot services.
    #undef ERROR_EXIT
    #define ERROR_EXIT PANIC_EXIT


    // Save the rest of the context variables needed to initialize the OS.
    Context->BS = NULL;
    Context->Lip = NULL;
    Context->BootStackTipAddr = NewTipAddr + MemMapSize;
    Context->MemoryMap = MemMap;
    Context->MemoryMapSize = MemMapSize;
    Context->MemoryMapDescriptorSize = MemMapDescSize;


    // Clear the framebuffer to black.
    if (!Context->Conf->IsGraphicsOff) {
        Sys_FrameBuffer_Clear(Context->FrameBuffer, 0);
    }


    // Send the initialization string to the debug serial port, indicating that the loader has taken control of I/O and
    // will begin initializing the OS.
    CHAR8 *OsInitString = Context->Conf->OsInitString;
    UINTN OsInitStringLen = Sys_Common_AsciiStrLen(OsInitString);
    if (!OsInitStringLen){
        PANIC_EXIT("entry: failed to get OsInitString", EFI_INVALID_PARAMETER, Context->Conf->DebugPort);
        return EFI_ABORTED;
    }
    Sys_Common_WriteSerial(Context->Conf->DebugPort, (UINT8*) OsInitString, OsInitStringLen);


    // Initialize the OS and never return.
    Sys_Init(Context);

    return EFI_ABORTED;
}


// Converts the specified UTF-16 Unicode string to a new UTF-8 encoded string allocated on the boot stack.
static const CHAR8 * SYSABI Sys_Boot_EntryUnicodeConvert(IN OUT SYS_BOOT_CONTEXT *Context, IN const CHAR16 *UnicodeString) {
    UINTN AlignedTipAddr = ALIGN_VALUE(Context->BootStackTipAddr, CPU_STACK_ALIGNMENT);
    CHAR8 *Utf8String;
    EFI_STATUS Status = UCS2StrToUTF8((CHAR16*) UnicodeString, &Utf8String);
    if (EFI_ERROR(Status)){
        return NULL;
    }
    UINTN Size = Sys_Common_AsciiStrCopy(Utf8String, (CHAR8*) AlignedTipAddr);
    FreePool(Utf8String);
    if (AlignedTipAddr + Size > Context->BootStackEndAddr) {
        return NULL;
    }
    Context->BootStackTipAddr = AlignedTipAddr + Size;
    return (CHAR8*) AlignedTipAddr;
}


//
// Attempts to initialize the framebuffer for the specified context. If an error occurs, force graphics off and
// save the status. If graphics off, just set to an uninitialized framebuffer and return success.
//
static EFI_STATUS SYSABI Sys_Boot_EntryInitializeFramebuffer(IN OUT SYS_BOOT_CONTEXT *Context) {

    SYS_FRAMEBUFFER *FrameBuffer = (SYS_FRAMEBUFFER*) Sys_Init_AllocBoot(Context, sizeof(SYS_FRAMEBUFFER), CPU_STACK_ALIGNMENT);
    if (FrameBuffer == NULL){
        return EFI_BUFFER_TOO_SMALL;
    }
    FrameBuffer->Status = EFI_NOT_STARTED;
    FrameBuffer->Base = NULL;
    FrameBuffer->Size = 0;
    FrameBuffer->PixelFormat = 0;
    FrameBuffer->RowSize = 0;
    FrameBuffer->Width = 0;
    FrameBuffer->Height = 0;

    if (Context->Conf->IsGraphicsOff) {
        Context->FrameBuffer = FrameBuffer;
        return EFI_SUCCESS;
    }

    EFI_GRAPHICS_OUTPUT_PROTOCOL *Gop = NULL;
    EFI_BOOT_SERVICES *BootServices = Context->BS;
    FrameBuffer->Status = BootServices->LocateProtocol(&gEfiGraphicsOutputProtocolGuid, NULL, (void**) &Gop);
        
    if (Gop && !EFI_ERROR(FrameBuffer->Status)) {
        if (Gop->Mode->Info->PixelFormat == PixelRedGreenBlueReserved8BitPerColor) {
            FrameBuffer->PixelFormat = PixelFormatRGB;
        } else if (Gop->Mode->Info->PixelFormat == PixelBlueGreenRedReserved8BitPerColor) {
            FrameBuffer->PixelFormat = PixelFormatBGR;
        } else {
            FrameBuffer->Status = EFI_UNSUPPORTED;
        }
    }

    if (!EFI_ERROR(FrameBuffer->Status)) {
        FrameBuffer->Base = (void*) Gop->Mode->FrameBufferBase;
        FrameBuffer->Size = Gop->Mode->FrameBufferSize;
        FrameBuffer->RowSize = Gop->Mode->Info->PixelsPerScanLine << 2;
        FrameBuffer->Width = Gop->Mode->Info->HorizontalResolution;
        FrameBuffer->Height = Gop->Mode->Info->VerticalResolution;
        UINTN NeededSize = FrameBuffer->Height * FrameBuffer->RowSize;
        if (FrameBuffer->Size < NeededSize) {
            FrameBuffer->Status = EFI_BAD_BUFFER_SIZE;
        }
    }
    
    if (EFI_ERROR(FrameBuffer->Status)){
        Context->Conf->IsGraphicsOff = TRUE;
    } else {
        FrameBuffer->Status = EFI_SUCCESS;
    }

    Context->FrameBuffer = FrameBuffer;

    return EFI_SUCCESS;
}

