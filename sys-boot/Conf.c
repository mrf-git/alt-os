#include <Uefi.h>
#include <Library/BaseLib.h>
#include <Library/BaseMemoryLib.h>
#include <Library/UefiLib.h>
#include <Library/PrintLib.h>
#include <Library/DevicePathLib.h>
#include <Library/MemoryAllocationLib.h>
#include <Guid/Acpi.h>
#include <Guid/SmBios.h>
#include <Protocol/LoadedImage.h>
#include <IndustryStandard/SmBios.h>
#include <Protocol/SimpleFileSystem.h>

#include "Conf.h"
#include "OsInit.h"


#define MAX_CONF_PATH_LEN               512  // Bigger than any FAT pathname can be.
#define MAX_CONF_OPT_INT_VALUE          0xFFFFFFFF

#define ACPI_RSDP_SIG                   "RSD PTR "
#define ACPI_RSDP_SIG_LEN               8

#define SMBIOS_ENTRY_ANCHOR             "_SM_"
#define SMBIOS_ENTRY_ANCHOR_LEN         4

#define SMBIOS3_ENTRY_ANCHOR            "_SM3_"
#define SMBIOS3_ENTRY_ANCHOR_LEN        5

#define SYS_CONF_FILENAME               L"SYS.CONF"
#define SYS_CONF_FILENAME_LEN           8


//
// Forward declarations.
//
static EFI_STATUS SYSABI Sys_Conf_SetBooleanFromAscii(IN const CHAR8 *ValueAscii, OUT BOOLEAN *OutPointer);
static EFI_STATUS SYSABI Sys_Conf_SetIntegerFromAscii(IN const CHAR8 *ValueAscii, IN const BOOLEAN IsUnsigned, OUT INTN *OutPointer);
static EFI_STATUS SYSABI Sys_Conf_SetStringFromAscii(IN const CHAR8 *ValueAscii, OUT CHAR8 *OutPointer);
static EFI_STATUS SYSABI Sys_Conf_SetConfOptFromAscii(IN const CHAR8 *OptName, IN const CHAR8 *OptValue, IN OUT SYS_BOOT_CONF *SysConf);


//
// Exported functions.
//

EFI_STATUS SYSABI Sys_Conf_InitializePaths(OUT CHAR16 **ImagePath, OUT CHAR16 **SysConfPath, IN OUT void *VContext) {

    EFI_STATUS Status;
    SYS_BOOT_CONTEXT *Context = (SYS_BOOT_CONTEXT*) VContext;
    EFI_LOADED_IMAGE_PROTOCOL *LoadedImage = (EFI_LOADED_IMAGE_PROTOCOL*) Context->Lip;
    CHAR16 *SourceImagePath = NULL;
    *ImagePath = NULL;
    *SysConfPath = NULL;
    CHAR16 *DestPath = NULL;

    SourceImagePath = ConvertDevicePathToText(LoadedImage->FilePath, TRUE, TRUE);  // NOTE: This must be freed.
    if (SourceImagePath == NULL){
        Status = EFI_LOAD_ERROR;
        goto cleanupReturn;
    }
    
    // Copy the source image path string.
    UINTN ImagePathLen = StrnLenS(SourceImagePath, MAX_CONF_PATH_LEN);
    UINTN ImagePathSize = (ImagePathLen + 1) << 1;  // +1 for terminating null; each char is 2 bytes.
    DestPath = (CHAR16*) Sys_Init_AllocBoot(Context, ImagePathSize, 2);
    if (DestPath == NULL){
        Status = EFI_BUFFER_TOO_SMALL;
        goto cleanupReturn;
    }
    DestPath[0] = 0;
    Status = StrCpyS(DestPath, MAX_CONF_PATH_LEN, SourceImagePath);
    if (EFI_ERROR(Status)){
        goto cleanupReturn;
    }
    *ImagePath = DestPath;

    // Copy the source base path string and append the filename.
    PathRemoveLastItem(SourceImagePath);
    UINTN BasePathLen = StrnLenS(SourceImagePath, MAX_CONF_PATH_LEN);
    UINTN SysConfPathLen = BasePathLen + SYS_CONF_FILENAME_LEN;
    UINTN SysConfPathSize = (SysConfPathLen + 1) << 1;
    DestPath = (CHAR16*) Sys_Init_AllocBoot(Context, SysConfPathSize, 2);
    if (DestPath == NULL){
        Status = EFI_BUFFER_TOO_SMALL;
        *ImagePath = NULL;
        goto cleanupReturn;
    }
    DestPath[0] = 0;
    Status = StrCpyS(DestPath, MAX_CONF_PATH_LEN, SourceImagePath);
    if (EFI_ERROR(Status)){
        *ImagePath = NULL;
        goto cleanupReturn;
    }
    Status = StrCpyS(&DestPath[BasePathLen], MAX_CONF_PATH_LEN, SYS_CONF_FILENAME);
    if (EFI_ERROR(Status)){
        *ImagePath = NULL;
        goto cleanupReturn;
    }
  
    *SysConfPath = DestPath;

    Status = EFI_SUCCESS;

cleanupReturn:
    if (SourceImagePath){
        FreePool(SourceImagePath);
    }
    return Status;
}


EFI_STATUS SYSABI Sys_Conf_Initialize(IN OUT CHAR16 **SysConfPath, IN OUT void *VContext) {
  
    EFI_STATUS Status;
    SYS_BOOT_CONTEXT *Context = (SYS_BOOT_CONTEXT*) VContext;
    EFI_LOADED_IMAGE_PROTOCOL *LoadedImage = (EFI_LOADED_IMAGE_PROTOCOL*) Context->Lip;
    EFI_FILE_HANDLE FsRoot = NULL;
    EFI_FILE_PROTOCOL *FileHandle = NULL;

    // Open simple filesystem root on the device where the image was loaded.
    EFI_SIMPLE_FILE_SYSTEM_PROTOCOL *Filesystem = NULL;
    Status = LoadedImage->SystemTable->BootServices->HandleProtocol(LoadedImage->DeviceHandle,
                                                                    &gEfiSimpleFileSystemProtocolGuid,
                                                                    (void**) &Filesystem);
    if (EFI_ERROR(Status)){
        goto cleanupReturn;
    }

    Status = Filesystem->OpenVolume(Filesystem, &FsRoot);
    if (EFI_ERROR(Status)){
        goto cleanupReturn;
    }

    // Initialize configuration with default values.
    SYS_BOOT_CONF *DestSysConf = (SYS_BOOT_CONF*) Sys_Init_AllocBoot(Context, sizeof(SYS_BOOT_CONF), CPU_STACK_ALIGNMENT);
    if (DestSysConf == NULL){
        Status = EFI_BUFFER_TOO_SMALL;
        goto cleanupReturn;
    }
    DestSysConf->OsInitString = (CHAR8*) Sys_Init_AllocBoot(Context, SYS_BOOT_MAX_CONF_OPT_LEN, CPU_STACK_ALIGNMENT);
    if (DestSysConf->OsInitString == NULL){
        Status = EFI_BUFFER_TOO_SMALL;
        goto cleanupReturn;
    }
    DestSysConf->OsInitString[0] = 0;
    AsciiStrCpyS(DestSysConf->OsInitString, SYS_BOOT_MAX_CONF_OPT_LEN, SYS_BOOT_CONF_DEFAULT_OS_INIT_STRING);
    DestSysConf->IsGraphicsOff = SYS_BOOT_CONF_DEFAULT_IS_GRAPHICS_OFF;
    DestSysConf->DebugPort = SYS_BOOT_CONF_DEFAULT_DEBUG_PORT;

    // Attempt to read the contents of the configuration file as a null-terminated ASCII string.
    Status = FsRoot->Open(FsRoot, &FileHandle, *SysConfPath, EFI_FILE_MODE_READ, 0);
    if (EFI_ERROR(Status)){
        // If the configuration file can't be opened successfully set the default configuration and clear the return path.
        Context->Conf = DestSysConf;
        *SysConfPath = NULL;
        Status = EFI_SUCCESS;
        goto cleanupReturn;
    }

    UINTN TempTipAddr = ALIGN_VALUE(Context->BootStackTipAddr, CPU_STACK_ALIGNMENT);
    UINTN SysConfAsciiLen = Context->BootStackEndAddr - TempTipAddr - 1;  // Ensure one extra padding byte for the loop below.
    CHAR8 *SysConfAscii = (CHAR8*) TempTipAddr;
    Status = FileHandle->Read(FileHandle, &SysConfAsciiLen, (void*) SysConfAscii);
    if (EFI_ERROR(Status)){
        goto cleanupReturn;
    }
    SysConfAscii[SysConfAsciiLen] = 0;
    SysConfAscii[SysConfAsciiLen+1] = 0;

    // Parse the contents of the configuration file and set each specified configuration option.
    CHAR8 *CurAscii = SysConfAscii;
    UINTN CurLen = 0;
    CHAR8 *OptName = NULL;
    BOOLEAN IsUpperCasing = TRUE;

    for (UINTN i=0; i < SysConfAsciiLen+1; i++) {
        CHAR8 c = SysConfAscii[i];
        switch (c){
            case 0:
            case '\n':
            case '\r':
            case ';':
                SysConfAscii[i] = 0;
                if (CurLen && OptName){
                    Status = Sys_Conf_SetConfOptFromAscii(OptName, CurAscii, DestSysConf);
                    if (EFI_ERROR(Status)){
                        goto cleanupReturn;
                    }
                }
                CurAscii = &SysConfAscii[i+1];
                OptName = NULL;
                CurLen = 0;
                IsUpperCasing = TRUE;
                break;

            case '=':
                if (!CurLen || OptName){
                    Status = EFI_INVALID_PARAMETER;
                    goto cleanupReturn;
                }
                SysConfAscii[i] = 0;
                OptName = CurAscii;
                CurAscii = &SysConfAscii[i+1];
                CurLen = 0;
                IsUpperCasing = AsciiStrnCmp(OptName, "OSINITSTRING", SYS_BOOT_MAX_CONF_OPT_LEN) != 0;
                break;

            case '\\':
            case '-':
            case '_':
            case ' ':
                CurLen++;
                break;  // Exception chars allowed.

            default:
                if (c < '0' || (c > '9' && c < 'A') || (c > 'Z' && c < 'a') || c > 'z'){
                    Status = EFI_INVALID_PARAMETER;
                    goto cleanupReturn;
                }
                if (IsUpperCasing) {
                    SysConfAscii[i] = AsciiCharToUpper(c);
                } else {
                    SysConfAscii[i] = c;
                }
                CurLen++;
        }

        if (c == 0){
            break;
        }
    }

    // Successfully set the configuration, cleanup and return.
    Context->Conf = DestSysConf;
    Status = EFI_SUCCESS;

cleanupReturn:
    if (FileHandle){
        FileHandle->Close(FileHandle);
    }
    if (FsRoot){
        FsRoot->Close(FsRoot);
    }

    return Status;
}


EFI_STATUS SYSABI Sys_Conf_ProcessUefiTable(IN void *VConfigurationTable,
        IN const UINTN NumberOfEntries, IN OUT void *VContext) {

    EFI_CONFIGURATION_TABLE *ConfigurationTable = (EFI_CONFIGURATION_TABLE*) VConfigurationTable;
    if (!ConfigurationTable || !ConfigurationTable->VendorTable){
        return EFI_INVALID_PARAMETER;
    }

    SYS_BOOT_CONTEXT *Context = (SYS_BOOT_CONTEXT*) VContext;
    EFI_ACPI_1_0_ROOT_SYSTEM_DESCRIPTION_POINTER *AcpiRoot = NULL;
    SMBIOS_TABLE_ENTRY_POINT *SmbiosEntry = NULL;
    SMBIOS_TABLE_3_0_ENTRY_POINT *Smbios3Entry = NULL;

    // Scan the configuration table and find the pointers.
    for (UINTN i = 0; i < NumberOfEntries; i++) {
        EFI_GUID *Guid = &ConfigurationTable[i].VendorGuid;
        void *Table = ConfigurationTable[i].VendorTable;

        if ((CompareGuid(Guid, &gEfiAcpiTableGuid) ||
            CompareGuid(Guid, &gEfiAcpi10TableGuid) ||
            CompareGuid(Guid, &gEfiAcpi20TableGuid)) &&
            !AsciiStrnCmp(Table, ACPI_RSDP_SIG, ACPI_RSDP_SIG_LEN)){
            EFI_ACPI_1_0_ROOT_SYSTEM_DESCRIPTION_POINTER *NewAcpiRoot = Table;
            if (AcpiRoot == NULL || NewAcpiRoot->Reserved > AcpiRoot->Reserved){
                AcpiRoot = NewAcpiRoot;
            }
        }

        if (CompareGuid(Guid, &gEfiSmbiosTableGuid) && !AsciiStrnCmp(Table, SMBIOS_ENTRY_ANCHOR, SMBIOS_ENTRY_ANCHOR_LEN)){
            SmbiosEntry = Table;
        }

        if (CompareGuid(Guid, &gEfiSmbios3TableGuid) && !AsciiStrnCmp(Table, SMBIOS3_ENTRY_ANCHOR, SMBIOS3_ENTRY_ANCHOR_LEN)){
            Smbios3Entry = Table;
        }
    }

    // Set global pointers and version/revision.
    if (Smbios3Entry != NULL) {
        Context->SmbiosEntry = Smbios3Entry;
        Context->SmbiosVersion = 3;
    } else if (SmbiosEntry != NULL) {
        Context->SmbiosEntry = SmbiosEntry;
        Context->SmbiosVersion = 1;
    }

    if (AcpiRoot != NULL) {
        Context->Rsdp = AcpiRoot;
        Context->AcpiRevision = (AcpiRoot->Reserved > 0) ? 2 : 1;
    }

    return EFI_SUCCESS;
}


//
// Parses the specified ASCII-valued configuration option name and value, and sets the corresponding value in the
// specified configuration structure. Returns error status if the option name or value is invalid.
//
static EFI_STATUS SYSABI Sys_Conf_SetConfOptFromAscii(IN const CHAR8 *OptName, IN const CHAR8 *OptValue, IN OUT SYS_BOOT_CONF *SysConf) {

    EFI_STATUS Status;

    if (!AsciiStrnCmp(OptName, "ISGRAPHICSOFF", SYS_BOOT_MAX_CONF_OPT_LEN)) {
        Status = Sys_Conf_SetBooleanFromAscii(OptValue, &SysConf->IsGraphicsOff);
        if (EFI_ERROR(Status)){
            return Status;
        }
        return EFI_SUCCESS;
    }

    if (!AsciiStrnCmp(OptName, "DEBUGPORT", SYS_BOOT_MAX_CONF_OPT_LEN)) {
        INTN DebugPortIntn;  // DebugPort is defined as a 16-bit int, so we convert it to full machine width here.
        Status = Sys_Conf_SetIntegerFromAscii(OptValue, TRUE, &DebugPortIntn);
        if (EFI_ERROR(Status)){
            return Status;
        }
        SysConf->DebugPort = DebugPortIntn;
        return EFI_SUCCESS;
    }

    if (!AsciiStrnCmp(OptName, "OSINITSTRING", SYS_BOOT_MAX_CONF_OPT_LEN)) {
        Status = Sys_Conf_SetStringFromAscii(OptValue, SysConf->OsInitString);
        if (EFI_ERROR(Status)){
            return Status;
        }
        return EFI_SUCCESS;
    }

    return EFI_INVALID_PARAMETER;
}


//
// Parses the specified ASCII value as a BOOLEAN and sets the value to OutPointer.
//
static EFI_STATUS SYSABI Sys_Conf_SetBooleanFromAscii(IN const CHAR8 *ValueAscii, OUT BOOLEAN *OutPointer) {
    if (!AsciiStrnCmp(ValueAscii, "TRUE", SYS_BOOT_MAX_CONF_OPT_LEN) ||
        !AsciiStrnCmp(ValueAscii, "YES", SYS_BOOT_MAX_CONF_OPT_LEN) ||
        !AsciiStrnCmp(ValueAscii, "1", SYS_BOOT_MAX_CONF_OPT_LEN)) {
        *OutPointer = TRUE;
        return EFI_SUCCESS;

    } else if (!AsciiStrnCmp(ValueAscii, "FALSE", SYS_BOOT_MAX_CONF_OPT_LEN) ||
        !AsciiStrnCmp(ValueAscii, "NO", SYS_BOOT_MAX_CONF_OPT_LEN) ||
        !AsciiStrnCmp(ValueAscii, "0", SYS_BOOT_MAX_CONF_OPT_LEN)) {
        *OutPointer = FALSE;
        return EFI_SUCCESS;
    }

    return EFI_INVALID_PARAMETER;
}


//
// Parses the specified ASCII value as a INTN and sets the value to OutPointer.
//
static EFI_STATUS SYSABI Sys_Conf_SetIntegerFromAscii(IN const CHAR8 *ValueAscii, IN const BOOLEAN IsUnsigned, OUT INTN *OutPointer) {
    UINTN Len = AsciiStrnLenS(ValueAscii, SYS_BOOT_MAX_CONF_OPT_LEN);

    if (Len < 1){
        return EFI_INVALID_PARAMETER;
    }

    BOOLEAN IsNegative = ValueAscii[0] == '-';
    if (IsNegative) {
        if (IsUnsigned){
            return EFI_INVALID_PARAMETER;
        }
        ValueAscii++;
    }

    INTN RetVal = 0;
    BOOLEAN IsError = TRUE;

    if (Len > 1 && ValueAscii[0] == '0' && ValueAscii[1] == 'X') {
        // Value string starts with "0x" and is interpreted as hex.
        ValueAscii += 2;
        CHAR8 c = *(ValueAscii++);
        while (c) {
            if ((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
                RetVal = (RetVal << 4) | (c - (c >= 'A' ? 'A' - 10 : '0'));
                IsError = FALSE;
                if (RetVal > MAX_CONF_OPT_INT_VALUE) {
                    IsError = TRUE;
                    break;
                }
            } else {
                IsError = TRUE;
                break;
            }
            c = *(ValueAscii++);
        }

    } else {
        // Value is interpreted as decimal.
        CHAR8 c = *(ValueAscii++);
        while (c) {
            if (c >= '0' && c <= '9') {
                RetVal = (RetVal * 10) + c - '0';
                IsError = FALSE;
                if (RetVal > MAX_CONF_OPT_INT_VALUE) {
                    IsError = TRUE;
                    break;
                }
            } else {
                IsError = TRUE;
                break;
            }
            c = *(ValueAscii++);
        }
    }

    if (IsError || RetVal < 0) {
        return EFI_INVALID_PARAMETER;
    }
    
    if (IsNegative) {
        RetVal = -RetVal;
    }
    
    *OutPointer = RetVal;
    
    return EFI_SUCCESS;
}


//
// Parses the specified ASCII value as an escapable string and sets the value to OutPointer.
//
static EFI_STATUS SYSABI Sys_Conf_SetStringFromAscii(IN const CHAR8 *ValueAscii, OUT CHAR8 *OutPointer) {
    
    BOOLEAN IsEscaped = FALSE;
    UINTN OutInd = 0;
    CHAR8 c = *(ValueAscii++);
    while (c) {
        if (OutInd >= SYS_BOOT_MAX_CONF_OPT_LEN) {
            return EFI_INVALID_PARAMETER;
        }
        if (IsEscaped) {
            if (c == 'n' || c == 'N') {
                OutPointer[OutInd++] = '\n';
            } else if (c == 'r' || c == 'R') {
                OutPointer[OutInd++] = '\r';
            } else {
                // Not really escaped.
                OutPointer[OutInd++] = '\\';
                OutPointer[OutInd++] = c;
            }
            IsEscaped = FALSE;
        } else if (c == '\\') {
            IsEscaped = TRUE;
        } else {
            OutPointer[OutInd++] = c;
            IsEscaped = FALSE;
        }
        c = *(ValueAscii++);
    }

    if (IsEscaped) {
        // Not really escaped.
        OutPointer[OutInd++] = '\\';
    }
    OutPointer[OutInd] = 0;

    return EFI_SUCCESS;
}
