#include <Uefi.h>
#include "Elf.h"


//
// ELF internal definitions.
//
#define ELF_MAX_DYNAMIC                     512

#define ELF_CUR_VERSION                     1
#define ELF_32BIT                           1
#define ELF_64BIT                           2
#define ELF_LE                              1
#define ELF_BE                              2
#define ELF_MACH_AMD64                      0x3e
#define ELF_OSABI_ALT                       88
#define ELF_DT_RELA                         7
#define ELF_PT_LOAD                         1
#define ELF_PT_DYNAMIC                      2
#define ELF_PT_INTERP                       3
#define ELF_PT_NOTE                         4
#define ELF_PT_PHDR                         6
#define ELF_PT_LOOS                         0x60000000
#define ELF_ET_DYN                          3
#define ELF_PF_X                            0x01
#define ELF_PF_W                            0x02
#define ELF_PF_R                            0x04
#define ELF_SHF_WRITE                       0x01
#define ELF_SHF_ALLOC                       0x02
#define ELF_SHF_EXEC                        0x04
#define ELF_STB_GLOBAL                      1
#define ELF_STB_WEAK                        2
#define ELF_STT_FILE                        4
#define ELF_X86_64_PLT_ENTRY_SIZE           16
#define ELF_X86_64_GOT_ENTRY_SIZE           8
#define ELF_X86_64_JUMP_ADDEND              6
#define ELF_X86_64_JUMP_OP_SIZE             2


//
// Macro definitions.
//

#define ZERO_MEM(From,To)                   for (UINTN Addr=From; Addr < To; Addr++) *((CHAR8*) Addr) = 0;



//
// Structure definitions.
//

typedef struct {
    UINT32 Type;
    UINT32 Flags;
    UINT64 Offset;
    UINT64 Vaddr;
    UINT64 _Reserved;
    UINT64 SizeFile;
    UINT64 SizeMem;
    UINT64 Align;
} ELF_PHDR_SEGMENT;

typedef struct {
    UINT64 Tag;
    UINT64 Val;
} ELF_PHDR_DYNAMIC;

typedef struct {
    UINT32 NameOffset;
    UINT32 Type;
    UINT64 Flags;
    UINT64 Vaddr;
    UINT64 Offset;
    UINT64 Size;
    UINT32 Link;
    UINT32 Info;
    UINT64 Align;
    UINT64 EntrySize;
} ELF_SHDR;

typedef struct {
    UINT32 NameOffset;
    UINT8 Info;
    UINT8 _Reserved;
    UINT16 SectionIndex;
    UINT64 Value;
    UINT64 Size;
} ELF_SYMTAB;

typedef struct {
    UINT64 Offset;
    UINT64 Info;
    UINT64 Addend;
} ELF_RELA;

typedef struct {
    SYS_ELF_LIB *Elf;
    SYS_HASH_TABLE *Htable;
    UINTN DebugPort;
    SYS_ELF_RELOCATION *Reloc;
    UINTN *SectionStartAddrs;
    UINTN *SectionEndAddrs;
    const CHAR8 **PltJumpNames;
} ELF_RELOC_PARAMS;



//
// Forward declarations.
//

static void Sys_Elf_InsertSymbols(IN CONST UINTN SectionIndex, IN const UINTN ShdrOffset,
    IN OUT SYS_ELF_LIB *Elf, IN OUT SYS_BOOT_CONTEXT *Context);
static void Sys_Elf_InsertSymRelocs(IN CONST UINTN SectionIndex, IN const UINTN ShdrOffset,
    IN OUT SYS_ELF_LIB *Elf, IN OUT SYS_BOOT_CONTEXT *Context);
static void Sys_Elf_InsertPltRelocs(IN CONST UINTN SectionIndex, IN const UINTN ShdrOffset,
    IN OUT SYS_ELF_LIB *Elf, IN OUT SYS_BOOT_CONTEXT *Context);

static BOOLEAN Sys_Elf_IsSectionSymTab(IN const SYS_ELF_SECTION *Section);
static BOOLEAN Sys_Elf_IsSectionRelaText(IN const SYS_ELF_SECTION *Section);
static BOOLEAN Sys_Elf_IsSectionRelaPlt(IN const SYS_ELF_SECTION *Section, IN const UINTN PltRelocVaddr);
static BOOLEAN Sys_Elf_IsSectionRelaEhFrame(IN const SYS_ELF_SECTION *Section);
static BOOLEAN Sys_Elf_IsSectionStrTab(IN const SYS_ELF_SECTION *Section, IN const UINTN StrTabVaddr);
static BOOLEAN Sys_Elf_IsSectionShStrTab(IN const SYS_ELF_SECTION *Section);
static BOOLEAN Sys_Elf_IsSectionDynStrTab(IN const SYS_ELF_SECTION *Section);
static BOOLEAN Sys_Elf_IsSectionDynSym(IN const SYS_ELF_SECTION *Section);
static BOOLEAN Sys_Elf_IsSectionDynamic(IN const SYS_ELF_SECTION *Section, IN const UINTN DynamicVaddr);
static BOOLEAN Sys_Elf_IsSectionEhFrame(IN const SYS_ELF_SECTION *Section);
static BOOLEAN Sys_Elf_IsSectionBss(IN const SYS_ELF_SECTION *Section);
static BOOLEAN Sys_Elf_IsSectionData(IN const SYS_ELF_SECTION *Section);
static BOOLEAN Sys_Elf_IsSectionRoData(IN const SYS_ELF_SECTION *Section);
static BOOLEAN Sys_Elf_IsSectionText(IN const SYS_ELF_SECTION *Section);
static BOOLEAN Sys_Elf_IsSectionPlt(IN const SYS_ELF_SECTION *Section);
static BOOLEAN Sys_Elf_IsSectionGot(IN const SYS_ELF_SECTION *Section, IN const UINTN GotVaddr);
static BOOLEAN Sys_Elf_Rel_PC32(IN OUT ELF_RELOC_PARAMS *Params);
static BOOLEAN Sys_Elf_Rel_PLT32(IN OUT ELF_RELOC_PARAMS *Params);
static BOOLEAN Sys_Elf_Rel_JMP_SLOT(IN OUT ELF_RELOC_PARAMS *Params);


//
// Exported functions. 
//

SYS_ELF_LIB * SYSABI Sys_Elf_ReadLib(IN const UINT8 *ElfBytes, IN const UINTN ElfBytesSize,
    IN OUT SYS_BOOT_CONTEXT *Context) {

    // Ensure normal, current-version ELF header.
    if (ElfBytesSize < 52 || ElfBytes[6] != ELF_CUR_VERSION || ElfBytes[20] != ELF_CUR_VERSION
        || ElfBytes[0] != 0x7f || ElfBytes[1] != 'E' || ElfBytes[2] != 'L' || ElfBytes[3] != 'F') {
        PANIC_EXIT("elf: bad header", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return NULL;
    }

    // Expecting dynamic library type for boot code.
    if (ElfBytes[16] != ELF_ET_DYN) {
        SYS_SERIAL_LOG("elf: type: ", Context->Conf->DebugPort);
        SYS_SERIAL_LOG_INT(ElfBytes[16], TRUE, Context->Conf->DebugPort);
        SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
        PANIC_EXIT("elf: bad type", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return NULL;
    }

    // Must be alt-os type.
    if (ElfBytes[7] != ELF_OSABI_ALT) {
        SYS_SERIAL_LOG("elf: abi: ", Context->Conf->DebugPort);
        SYS_SERIAL_LOG_INT(ElfBytes[7], TRUE, Context->Conf->DebugPort);
        SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
        PANIC_EXIT("elf: bad abi", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return NULL;
    }

#if defined(__amd64__)
    if (ElfBytesSize < 64 || ElfBytes[4] != ELF_64BIT || ElfBytes[5] != ELF_LE) {  // Ensure little-endian ELF64 header.
        PANIC_EXIT("elf: bad format for amd64", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return NULL;
    }
    if (ElfBytes[18] != ELF_MACH_AMD64){
        PANIC_EXIT("elf: bad machine for amd64", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return NULL;
    }

    // Read from file header.
    UINT64 PhdrOffset = *( (UINT64*) &ElfBytes[32]);
    UINT64 ShdrOffset = *( (UINT64*) &ElfBytes[40]);
    UINT16 PhdrEntSize = *( (UINT16*) &ElfBytes[54]);
    UINT16 PhdrEntCount = *( (UINT16*) &ElfBytes[56]);
    UINT16 ShdrEntSize = *( (UINT16*) &ElfBytes[58]);
    UINT16 ShdrEntCount = *( (UINT16*) &ElfBytes[60]);

    const UINT8 *CurBytes;

    // Alloc and init the output ELF structure.
    SYS_ELF_LIB *ElfLibrary = (SYS_ELF_LIB*) Sys_Init_AllocBoot(Context, sizeof(SYS_ELF_LIB), CPU_STACK_ALIGNMENT);
    if (ElfLibrary == NULL) {
        PANIC_EXIT("elf: not enough elf memory", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return NULL;
    }
    ElfLibrary->ElfBytes = (UINT8*) ElfBytes;
    ElfLibrary->ElfBytesSize = ElfBytesSize;
    ElfLibrary->InterpName = "";
    ElfLibrary->Symbols = NULL;
    ElfLibrary->NumSymbols = 0;
    ElfLibrary->SymRelocations = NULL;
    ElfLibrary->NumSymRelocations = 0;
    ElfLibrary->PltRelocations = NULL;
    ElfLibrary->NumPltRelocations = 0;
    ElfLibrary->EhRelocations = NULL;
    ElfLibrary->NumEhRelocations = 0;
    ElfLibrary->ExternalRefs = NULL;
    ElfLibrary->NumExternalRefs = 0;
    ElfLibrary->CodeMemorySize = 0;
    ElfLibrary->DataMemorySize = 0;
    ElfLibrary->RoDataMemorySize = 0;
    ElfLibrary->NumPltEntries = 0;
    ElfLibrary->NumGotEntries = 0;
    ElfLibrary->NumSectionPointers = ShdrEntCount;
    ElfLibrary->SectionPointerTable = (SYS_ELF_SECTION**) Sys_Init_AllocBoot(Context, sizeof(SYS_ELF_SECTION*) * ShdrEntCount, CPU_STACK_ALIGNMENT);
    ElfLibrary->StrTabSection = NULL;
    ElfLibrary->TextSection = NULL;
    ElfLibrary->DataSection = NULL;
    ElfLibrary->RoDataSection = NULL;
    ElfLibrary->BssSection = NULL;
    ElfLibrary->PltSection = NULL;
    ElfLibrary->GotSection = NULL;
    ElfLibrary->DynamicSection = NULL;
    ElfLibrary->EhFrameSection = NULL;
    for (UINTN i=0; i < ShdrEntCount; i++) {
        ElfLibrary->SectionPointerTable[i] = NULL;
    }

    // Get dynamic loading information from the ELF program header.
    const ELF_PHDR_SEGMENT *Segment;
    const UINT8 *DynamicBytes = NULL;
    UINTN DynamicVaddr = 0;
    UINTN DynamicOffset = 0;
    BOOLEAN IsLoadable = FALSE;
    CurBytes = &ElfBytes[PhdrOffset];
    for (UINTN i=0; i < PhdrEntCount; i++) {
        Segment = (const ELF_PHDR_SEGMENT*) CurBytes;
        CurBytes += PhdrEntSize;

        if(Segment->Type == ELF_PT_DYNAMIC) {
            DynamicBytes = (const UINT8*) &ElfBytes[Segment->Offset];
            DynamicVaddr = Segment->Vaddr;
            DynamicOffset = Segment->Offset;
        } else if(Segment->Type == ELF_PT_INTERP) {
            ElfLibrary->InterpName = (const CHAR8*) &ElfBytes[Segment->Offset];
        } else if(Segment->Type == ELF_PT_LOAD) {
            IsLoadable = TRUE;
        }
    }
    if (!IsLoadable || DynamicBytes == NULL) {
        PANIC_EXIT("elf: no loadable segments", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return NULL;
    }

    // Get dynamic section header info and locate the section names string table.
    const ELF_SHDR *Shdr;
    UINTN NumDynamic = 0;
    UINTN DynamicEntrySize = 0;
    const UINT8 *ShStrTabBytes = NULL;
    CurBytes = &ElfBytes[ShdrOffset];
    for (UINTN i=0; i < ShdrEntCount; i++) {
        Shdr = (const ELF_SHDR*) CurBytes;
        CurBytes += ShdrEntSize;

        if (Shdr->Type == ElfSectDynamic && Shdr->Vaddr == DynamicVaddr && Shdr->EntrySize != 0) {
            NumDynamic = Shdr->Size / Shdr->EntrySize;
            DynamicEntrySize = Shdr->EntrySize;
        } else if (Shdr->Type == ElfSectStrTab) {
            const CHAR8 *SecName = (const CHAR8*) &ElfBytes[Shdr->Offset+Shdr->NameOffset];
            if (Sys_Common_AsciiStrCmp(SecName, ".shstrtab") == 0) {
                ShStrTabBytes = &ElfBytes[Shdr->Offset];
                break;
            }
        }
    }
    if (ShStrTabBytes == NULL) {    
        PANIC_EXIT("elf: section names table not found", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return NULL;
    }

    // Parse the dynamic section bytes.
    const ELF_PHDR_DYNAMIC *Dyn;
    const UINT8 *StrTabBytes = NULL;
    UINTN StrTabVaddr = 0;
    UINTN StrTabSize = 0;
    UINTN SymTabEntrySize = 0;
    UINTN NumPltRelocations = 0;
    UINTN PltRelocTabSize = 0;
    UINTN PltRelocTabVaddr = 0;
    UINTN GotVaddr = 0;
    CurBytes = DynamicBytes;
    for (UINTN i=0; i < NumDynamic; i++) {
        Dyn = (const ELF_PHDR_DYNAMIC*) CurBytes;
        CurBytes += DynamicEntrySize;

        if (Dyn->Tag == ElfDynNull) {
            // End of dynamic array.
            break;
        }
        INTN DynOffset = (INTN) Dyn->Val - (INTN) DynamicVaddr + DynamicOffset;
        switch (Dyn->Tag) {
            default:
                if (Dyn->Tag < ElfDynMaxPosTags) {
                    SYS_SERIAL_LOG("elf: unknown dynamic tag: ", Context->Conf->DebugPort);
                    SYS_SERIAL_LOG_INT(Dyn->Tag, TRUE, Context->Conf->DebugPort);
                    SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
                    PANIC_EXIT("elf: unknown dynamic tag", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                    return NULL;
                }
                break;
            case ElfDynFlags:
                // Ignored.
                break;
            case ElfDynJumpRel:
                PltRelocTabVaddr = Dyn->Val;
                break;
            case ElfDynPltRel:
                if (Dyn->Val != ELF_DT_RELA) {
                    PANIC_EXIT("elf: unsupported relocation type", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                    return NULL;
                }
                break;
            case ElfDynPltRelSize:
                PltRelocTabSize = Dyn->Val;
                break;
            case ElfDynPltGot:
                GotVaddr = Dyn->Val;
                break;
            case ElfDynSymTab:
                // Handled below.
                break;
            case ElfDynSymTabEntrySize:
                SymTabEntrySize = Dyn->Val;
                break;
            case ElfDynStrTab:
                StrTabBytes = &ElfBytes[DynOffset];
                StrTabVaddr = Dyn->Val;
                break;
            case ElfDynStrTabSize:
                StrTabSize = Dyn->Val;
                break;
            case ElfDynPreInitArray:
                // Unsupported.
                break;
            case ElfDynPreInitArraySize:
                if (Dyn->Val) {
                    PANIC_EXIT("elf: ElfDynPreInitArray unsupported", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                    return NULL;
                }
                break;
            case ElfDynInitArray:
                // Unsupported.
                break;
            case ElfDynInitArraySize:
                if (Dyn->Val) {
                    PANIC_EXIT("elf: ElfDynInitArray unsupported", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                    return NULL;
                }
                break;
            case ElfDynFiniArray:
                // Unsupported.
                break;
            case ElfDynFiniArraySize:
                if (Dyn->Val) {
                    PANIC_EXIT("elf: ElfDynFiniArray unsupported", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                    return NULL;
                }
                break;
    
        }
    }
    if (SymTabEntrySize == 0) {
        PANIC_EXIT("elf: missing symbols", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return NULL;
    }
    if (StrTabBytes == NULL || StrTabSize == 0) {
        PANIC_EXIT("elf: missing strings", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return NULL;
    }
    NumPltRelocations = PltRelocTabSize / sizeof(ELF_RELA);
    if (NumPltRelocations > 0 && GotVaddr == 0) {
        PANIC_EXIT("elf: missing GOT", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return NULL;
    }

    // Reserve space for the sections and set their pointers, and calculate the required bytes needed to load Code,
    // Data, and RoData into memory.
    UINTN CodeMemorySize = 0;
    UINTN DataMemorySize = 0;
    UINTN RoDataMemorySize = 0;
    UINTN SymTabIndex = 0;
    UINTN SymRelocsIndex = 0;
    UINTN PltRelocsIndex = 0;

    CurBytes = &ElfBytes[ShdrOffset];
    for (UINTN i=0; i < ShdrEntCount; i++) {
        Shdr = (const ELF_SHDR*) CurBytes;
        CurBytes += ShdrEntSize;
        const CHAR8 *SecName = (const CHAR8*) ShStrTabBytes + Shdr->NameOffset;

        // Skip null section.
        if (Shdr->Type == 0 || Sys_Common_AsciiStrCmp(SecName, "") == 0) {
            continue;
        }
        // Skip already-read interp section.
        if (Sys_Common_AsciiStrCmp(SecName, ".interp") == 0) {
            continue;
        }

        SYS_ELF_SECTION *Section = (SYS_ELF_SECTION*) Sys_Init_AllocBoot(Context, sizeof(SYS_ELF_SECTION), CPU_STACK_ALIGNMENT);
        if (Section == NULL) {
            PANIC_EXIT("elf: not enough section memory", SYS_STATUS_FAIL, Context->Conf->DebugPort);
            return NULL;
        }
        UINTN SectionBaseAddr = (UINTN) Section;

        Section->Name = SecName;
        Section->Index = i;
        Section->Type = Shdr->Type;
        Section->IsWritable = (Shdr->Flags & ELF_SHF_WRITE) == ELF_SHF_WRITE;
        Section->IsExecutable = (Shdr->Flags & ELF_SHF_EXEC) == ELF_SHF_EXEC;
        Section->IsAllocated = (Shdr->Flags & ELF_SHF_ALLOC) == ELF_SHF_ALLOC;
        Section->Vaddr = Shdr->Vaddr;
        Section->Offset = Shdr->Offset;
        Section->SizeFile = Shdr->Size;
        Section->Align = Shdr->Align;
        Section->EntrySize = Shdr->EntrySize;

        if (Sys_Elf_IsSectionText(Section)){
            ElfLibrary->TextSection = Section;
            CodeMemorySize = ALIGN_VALUE(CodeMemorySize, Section->Align) + Section->SizeFile;
            CodeMemorySize = ALIGN_VALUE(CodeMemorySize, Section->Align);
        } else if (Sys_Elf_IsSectionPlt(Section)){
            ElfLibrary->PltSection = Section;
            ElfLibrary->NumPltEntries = ALIGN_VALUE(Section->SizeFile, Section->Align) / ELF_X86_64_PLT_ENTRY_SIZE;
            CodeMemorySize = ALIGN_VALUE(CodeMemorySize, Section->Align) + Section->SizeFile;
            CodeMemorySize = ALIGN_VALUE(CodeMemorySize, Section->Align);
        } else if (Sys_Elf_IsSectionGot(Section, GotVaddr)){
            ElfLibrary->GotSection = Section;
            ElfLibrary->NumGotEntries = ALIGN_VALUE(Section->SizeFile, Section->Align) / ELF_X86_64_GOT_ENTRY_SIZE;
            RoDataMemorySize = ALIGN_VALUE(RoDataMemorySize, Section->Align) + Section->SizeFile;
            RoDataMemorySize = ALIGN_VALUE(RoDataMemorySize, Section->Align);
        } else if (Sys_Elf_IsSectionData(Section)){
            ElfLibrary->DataSection = Section;
            DataMemorySize = ALIGN_VALUE(DataMemorySize, Section->Align) + Section->SizeFile;
            DataMemorySize = ALIGN_VALUE(DataMemorySize, Section->Align);
        } else if (Sys_Elf_IsSectionRoData(Section)){
            ElfLibrary->RoDataSection = Section;
            RoDataMemorySize = ALIGN_VALUE(RoDataMemorySize, Section->Align) + Section->SizeFile;
            RoDataMemorySize = ALIGN_VALUE(RoDataMemorySize, Section->Align);
        } else if (Sys_Elf_IsSectionBss(Section)){
            ElfLibrary->BssSection = Section;
            DataMemorySize = ALIGN_VALUE(DataMemorySize, Section->Align) + Section->SizeFile;
            DataMemorySize = ALIGN_VALUE(DataMemorySize, Section->Align);
        } else if (Sys_Elf_IsSectionStrTab(Section, StrTabVaddr)){
            ElfLibrary->StrTabSection = Section;
            RoDataMemorySize = ALIGN_VALUE(RoDataMemorySize, Section->Align) + Section->SizeFile;
            RoDataMemorySize = ALIGN_VALUE(RoDataMemorySize, Section->Align);
        } else if (Sys_Elf_IsSectionDynamic(Section, DynamicVaddr)) {
            ElfLibrary->DynamicSection = Section;
            RoDataMemorySize = ALIGN_VALUE(RoDataMemorySize, Section->Align) + Section->SizeFile;
            RoDataMemorySize = ALIGN_VALUE(RoDataMemorySize, Section->Align);
        }  else if (Sys_Elf_IsSectionEhFrame(Section)){
            ElfLibrary->EhFrameSection = Section;
            DataMemorySize = ALIGN_VALUE(DataMemorySize, Section->Align) + Section->SizeFile;
            DataMemorySize = ALIGN_VALUE(DataMemorySize, Section->Align);
        } else if (Sys_Elf_IsSectionSymTab(Section)){
            SymTabIndex = i;
            Context->BootStackTipAddr = SectionBaseAddr; // Free section.
        } else if (Sys_Elf_IsSectionRelaText(Section)){
            SymRelocsIndex = i;
            Context->BootStackTipAddr = SectionBaseAddr; // Free section.
        } else if (Sys_Elf_IsSectionRelaPlt(Section, PltRelocTabVaddr)){
            PltRelocsIndex = i;
            Context->BootStackTipAddr = SectionBaseAddr; // Free section.
        } else if (Sys_Elf_IsSectionRelaEhFrame(Section)){
            // EhRelocsIndex = i;
            // TODO
            Context->BootStackTipAddr = SectionBaseAddr; // Free section.
        } else if (Sys_Elf_IsSectionShStrTab(Section) || Sys_Elf_IsSectionDynStrTab(Section)
            || Sys_Elf_IsSectionDynSym(Section)){
            // Ignored, free the section.
            Context->BootStackTipAddr = SectionBaseAddr;
            Section = NULL;
        } else {
            SYS_SERIAL_LOG("elf: unhandled section: ", Context->Conf->DebugPort);
            SYS_SERIAL_LOG(Section->Name, Context->Conf->DebugPort);
            SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
            PANIC_EXIT("elf: unhandled section", SYS_STATUS_FAIL, Context->Conf->DebugPort);
            return NULL;
        }

        ElfLibrary->SectionPointerTable[i] = Section;
    }

    if (ElfLibrary->TextSection == NULL) {
        PANIC_EXIT("elf: no text section", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return NULL;
    }
    if (SymTabIndex == 0 || SymRelocsIndex == 0) {
        PANIC_EXIT("elf: no symbol relocs", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return NULL;
    }
    if (ElfLibrary->NumPltEntries != ElfLibrary->NumGotEntries - 2) {  // First 2 GOT entries are reserved.
        SYS_SERIAL_LOG("elf: bad PLT entries: ", Context->Conf->DebugPort);
        SYS_SERIAL_LOG_INT(ElfLibrary->NumGotEntries, TRUE, Context->Conf->DebugPort);
        SYS_SERIAL_LOG(" ", Context->Conf->DebugPort);
        SYS_SERIAL_LOG_INT(ElfLibrary->NumPltEntries, TRUE, Context->Conf->DebugPort);
        SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
        PANIC_EXIT("elf: bad PLT", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return NULL;
    }

    ElfLibrary->CodeMemorySize = CodeMemorySize;
    ElfLibrary->DataMemorySize = DataMemorySize;
    ElfLibrary->RoDataMemorySize = RoDataMemorySize;

    // Parse and insert the symbols and relocations.
    Sys_Elf_InsertSymbols(SymTabIndex, ShdrOffset, ElfLibrary, Context);
    Sys_Elf_InsertSymRelocs(SymRelocsIndex, ShdrOffset, ElfLibrary, Context);
    if (PltRelocsIndex > 0) {
        Sys_Elf_InsertPltRelocs(PltRelocsIndex, ShdrOffset, ElfLibrary, Context);
    }

    // Save external symbol references.
    for (UINTN i=0; i < ElfLibrary->NumSymbols; i++) {
        SYS_ELF_SYMBOL *Sym = &ElfLibrary->Symbols[i];
        if (Sym->IsGlobal && Sym->Type == ElfSymNone) {
            ElfLibrary->NumExternalRefs++;
        }
    }
    ElfLibrary->ExternalRefs = (const CHAR8**) Sys_Init_AllocBoot(Context, sizeof(const CHAR8*) * ElfLibrary->NumExternalRefs, CPU_STACK_ALIGNMENT);
    ElfLibrary->NumExternalRefs = 0;
    for (UINTN i=0; i < ElfLibrary->NumSymbols; i++) {
        SYS_ELF_SYMBOL *Sym = &ElfLibrary->Symbols[i];
        if (Sym->IsGlobal && Sym->Type == ElfSymNone) {
            ElfLibrary->ExternalRefs[ElfLibrary->NumExternalRefs++] = Sym->Name;
        }
    }

    return ElfLibrary;

#elif defined(__x86__)
    // Not implemented yet.
    return 0;
#else
    // #error "init: architecture not supported"
#endif
}


void SYSABI Sys_Elf_LoadLib(IN const UINTN CodeAddr, IN const UINTN DataAddr, IN const UINTN RoDataAddr,
    IN OUT SYS_ELF_LIB *Elf, IN OUT SYS_HASH_TABLE *Htable, IN OUT SYS_BOOT_CONTEXT *Context) {

    // Make sure all required external symbols already have addresses loaded into the hash table.
    for (UINTN i=0; i < Elf->NumExternalRefs; i++) {
        if (!Sys_Htable_Get(Htable, Elf->ExternalRefs[i], NULL)){
            SYS_SERIAL_LOG("elf: ref: ", Context->Conf->DebugPort);
            SYS_SERIAL_LOG(Elf->ExternalRefs[i], Context->Conf->DebugPort);
            SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
            PANIC_EXIT("elf: missing external reference", SYS_STATUS_FAIL, Context->Conf->DebugPort);
            return;
        }
    }

    // Save the boot stack tip to discard temp memory later.
    UINTN MemTipAddr = Context->BootStackTipAddr;
    
    // Initialize relocation parameters that will be partially overwritten for each relocation.
    ELF_RELOC_PARAMS *Params = Sys_Init_AllocBoot(Context, sizeof(ELF_RELOC_PARAMS), CPU_STACK_ALIGNMENT);
    if (Params == NULL) {
        PANIC_EXIT("elf: no memory to relocate", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return;
    }
    Params->Elf = Elf;
    Params->Htable = Htable;
    Params->DebugPort = Context->Conf->DebugPort;
    Params->Reloc = NULL;
    Params->SectionStartAddrs = (UINTN*) Sys_Init_AllocBoot(Context, sizeof(UINTN) * Elf->NumSectionPointers, CPU_STACK_ALIGNMENT);
    Params->SectionEndAddrs = (UINTN*) Sys_Init_AllocBoot(Context, sizeof(UINTN) * Elf->NumSectionPointers, CPU_STACK_ALIGNMENT);
    Params->PltJumpNames = (const CHAR8**) Sys_Init_AllocBoot(Context, sizeof(const CHAR8*) * Elf->NumPltEntries, CPU_STACK_ALIGNMENT);
    if (Params->SectionStartAddrs == NULL || Params->SectionEndAddrs == NULL || Params->PltJumpNames == NULL) {
        PANIC_EXIT("elf: no memory to relocate", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return;
    }
    for (UINTN i=0; i < Elf->NumSectionPointers; i++) {
        Params->SectionStartAddrs[i] = 0;
        Params->SectionEndAddrs[i] = 0;
    }
    for (UINTN i=0; i < Elf->NumPltEntries; i++) {
        Params->PltJumpNames[i] = NULL;
    }

    // Find relocated base addresses for ELF sections and allocate aligned memory there.
    SYS_ELF_SECTION *Section;
    UINTN CodeTipAddr = CodeAddr;
    UINTN DataTipAddr = DataAddr;
    UINTN RoDataTipAddr = RoDataAddr;

    Section = Elf->TextSection;
    Params->SectionStartAddrs[Section->Index] = ALIGN_VALUE(CodeTipAddr, Section->Align);
    Params->SectionEndAddrs[Section->Index] = ALIGN_VALUE(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Section->Align);
    CodeTipAddr = Params->SectionEndAddrs[Section->Index];
    if (Elf->PltSection != NULL){
        Section = Elf->PltSection;
        Params->SectionStartAddrs[Section->Index] = ALIGN_VALUE(CodeTipAddr, Section->Align);
        Params->SectionEndAddrs[Section->Index] = ALIGN_VALUE(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Section->Align);
        CodeTipAddr = Params->SectionEndAddrs[Section->Index];
    }
    if (Elf->GotSection != NULL){
        Section = Elf->GotSection;
        Params->SectionStartAddrs[Section->Index] = ALIGN_VALUE(RoDataTipAddr, Section->Align);
        Params->SectionEndAddrs[Section->Index] = ALIGN_VALUE(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Section->Align);
        RoDataTipAddr = Params->SectionEndAddrs[Section->Index];
    }
    if (Elf->DataSection != NULL){
        Section = Elf->DataSection;
        Params->SectionStartAddrs[Section->Index] = ALIGN_VALUE(DataTipAddr, Section->Align);
        Params->SectionEndAddrs[Section->Index] = ALIGN_VALUE(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Section->Align);
        DataTipAddr = Params->SectionEndAddrs[Section->Index];
    }
    if (Elf->RoDataSection != NULL){
        Section = Elf->RoDataSection;
        Params->SectionStartAddrs[Section->Index] = ALIGN_VALUE(RoDataTipAddr, Section->Align);
        Params->SectionEndAddrs[Section->Index] = ALIGN_VALUE(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Section->Align);
        RoDataTipAddr = Params->SectionEndAddrs[Section->Index];
    }
    if (Elf->BssSection != NULL){
        Section = Elf->BssSection;
        Params->SectionStartAddrs[Section->Index] = ALIGN_VALUE(DataTipAddr, Section->Align);
        Params->SectionEndAddrs[Section->Index] = ALIGN_VALUE(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Section->Align);
        DataTipAddr = Params->SectionEndAddrs[Section->Index];
    }
    if (Elf->StrTabSection != NULL){
        Section = Elf->StrTabSection;
        Params->SectionStartAddrs[Section->Index] = ALIGN_VALUE(RoDataTipAddr, Section->Align);
        Params->SectionEndAddrs[Section->Index] = ALIGN_VALUE(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Section->Align);
        RoDataTipAddr = Params->SectionEndAddrs[Section->Index];
    }
    if (Elf->DynamicSection != NULL){
        Section = Elf->DynamicSection;
        Params->SectionStartAddrs[Section->Index] = ALIGN_VALUE(RoDataTipAddr, Section->Align);
        Params->SectionEndAddrs[Section->Index] = ALIGN_VALUE(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Section->Align);
        RoDataTipAddr = Params->SectionEndAddrs[Section->Index];
    }
    if (Elf->EhFrameSection != NULL){
        Section = Elf->EhFrameSection;
        Params->SectionStartAddrs[Section->Index] = ALIGN_VALUE(DataTipAddr, Section->Align);
        Params->SectionEndAddrs[Section->Index] = ALIGN_VALUE(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Section->Align);
        DataTipAddr = Params->SectionEndAddrs[Section->Index];
    }

    // Execute the symbol relocations on each symbol in-memory.
    for (UINTN i=0; i < Elf->NumSymRelocations; i++) {
        Params->Reloc = &Elf->SymRelocations[i];

        switch (Params->Reloc->Type) {
        default:
            SYS_SERIAL_LOG("elf: relocation type: ", Context->Conf->DebugPort);
            SYS_SERIAL_LOG_INT(Params->Reloc->Type, TRUE, Context->Conf->DebugPort);
            SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
            PANIC_EXIT("elf: unhandled relocation", SYS_STATUS_FAIL, Context->Conf->DebugPort);
            return;

        case ElfRelocX86_64_PLT32:
            if (!Sys_Elf_Rel_PLT32(Params)) {
                SYS_SERIAL_LOG("elf: sym: ", Context->Conf->DebugPort);
                SYS_SERIAL_LOG(Params->Reloc->Symbol->Name, Context->Conf->DebugPort);
                SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
                PANIC_EXIT("elf: could not handle PLT32 relocation", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                return;
            }
            break;
        case ElfRelocX86_64_PC32:
            if (Params->Reloc->Symbol->Section == NULL || !Sys_Elf_Rel_PC32(Params)) {
                SYS_SERIAL_LOG("elf: sym: ", Context->Conf->DebugPort);
                SYS_SERIAL_LOG(Params->Reloc->Symbol->Name, Context->Conf->DebugPort);
                SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
                PANIC_EXIT("elf: could not handle PC32 relocation", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                return;
            }
            break;
        }
    }

    // Update hash table with addresses for global functions and make them available for local PLT relocations.
    for (UINTN i=0; i < Elf->NumSymbols; i++) {
        SYS_ELF_SYMBOL *Sym = &Elf->Symbols[i];
        if (Sym->IsGlobal && Sym->Type == ElfSymFunc && Sym->Vaddr != 0 && Sym->Section != NULL) {
            UINTN SectionAddr = Params->SectionStartAddrs[Sym->Section->Index];
            if (SectionAddr == 0) {
                PANIC_EXIT("elf: no relocation for symbol section", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                return;
            }
            UINTN SymVaddr = SectionAddr + Sym->Vaddr - Sym->Section->Vaddr;
            void *SymMem;
            if (Sys_Htable_Get(Htable, Sym->Name, &SymMem) && (UINTN) SymMem != SymVaddr) {
                SYS_SERIAL_LOG("elf: sym: ", Context->Conf->DebugPort);
                SYS_SERIAL_LOG(Sym->Name, Context->Conf->DebugPort);
                SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
                PANIC_EXIT("elf: duplicate symbol", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                return;
            }
            if (!Sys_Htable_Put(Htable, Sym->Name, (void*) SymVaddr)){
                PANIC_EXIT("elf: failed to put htable entry", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                return;
            }
        }
    }

    // Execute the PLT relocations on each symbol in-memory.
    for (UINTN i=0; i < Elf->NumPltRelocations; i++) {
        Params->Reloc = &Elf->PltRelocations[i];

        switch (Params->Reloc->Type) {
        default:
            SYS_SERIAL_LOG("elf: relocation type: ", Context->Conf->DebugPort);
            SYS_SERIAL_LOG_INT(Params->Reloc->Type, TRUE, Context->Conf->DebugPort);
            SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
            PANIC_EXIT("elf: unhandled relocation", SYS_STATUS_FAIL, Context->Conf->DebugPort);
            return;

        case ElfRelocX86_64_JMP_SLOT:
            if (!Sys_Elf_Rel_JMP_SLOT(Params)) {
                PANIC_EXIT("elf: could not handle JMP_SLOT relocation", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                return;
            }
            break;
        }
    }

    // TODO execute eh_frame relocations.

    // Copy relocated sections to Code, Data, and RoData memory and zero-initialize extra space.
    Section = Elf->TextSection;
    Sys_Common_MemCopy(&Elf->ElfBytes[Section->Offset], Section->SizeFile, (UINT8*) Params->SectionStartAddrs[Section->Index]);
    ZERO_MEM(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Params->SectionEndAddrs[Section->Index]);
    if (Elf->PltSection != NULL) {
        Section = Elf->PltSection;
        Sys_Common_MemCopy(&Elf->ElfBytes[Section->Offset], Section->SizeFile, (UINT8*) Params->SectionStartAddrs[Section->Index]);
        ZERO_MEM(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Params->SectionEndAddrs[Section->Index]);
    }
    if (Elf->GotSection != NULL) {
        Section = Elf->GotSection;
        Sys_Common_MemCopy(&Elf->ElfBytes[Section->Offset], Section->SizeFile, (UINT8*) Params->SectionStartAddrs[Section->Index]);
        ZERO_MEM(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Params->SectionEndAddrs[Section->Index]);
    }
    if (Elf->DataSection != NULL) {
        Section = Elf->DataSection;
        Sys_Common_MemCopy(&Elf->ElfBytes[Section->Offset], Section->SizeFile, (UINT8*) Params->SectionStartAddrs[Section->Index]);
        ZERO_MEM(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Params->SectionEndAddrs[Section->Index]);
    }
    if (Elf->RoDataSection != NULL) {
        Section = Elf->RoDataSection;
        Sys_Common_MemCopy(&Elf->ElfBytes[Section->Offset], Section->SizeFile, (UINT8*) Params->SectionStartAddrs[Section->Index]);
        ZERO_MEM(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Params->SectionEndAddrs[Section->Index]);
    }
    if (Elf->StrTabSection != NULL) {
        Section = Elf->StrTabSection;
        Sys_Common_MemCopy(&Elf->ElfBytes[Section->Offset], Section->SizeFile, (UINT8*) Params->SectionStartAddrs[Section->Index]);
        ZERO_MEM(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Params->SectionEndAddrs[Section->Index]);
    }
    if (Elf->DynamicSection != NULL) {
        Section = Elf->DynamicSection;
        Sys_Common_MemCopy(&Elf->ElfBytes[Section->Offset], Section->SizeFile, (UINT8*) Params->SectionStartAddrs[Section->Index]);
        ZERO_MEM(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Params->SectionEndAddrs[Section->Index]);
    }
    if (Elf->EhFrameSection != NULL) {
        Section = Elf->EhFrameSection;
        Sys_Common_MemCopy(&Elf->ElfBytes[Section->Offset], Section->SizeFile, (UINT8*) Params->SectionStartAddrs[Section->Index]);
        ZERO_MEM(Params->SectionStartAddrs[Section->Index] + Section->SizeFile, Params->SectionEndAddrs[Section->Index]);
    }
    if (Elf->BssSection != NULL) {
        Section = Elf->BssSection;
        ZERO_MEM(Params->SectionStartAddrs[Section->Index], Params->SectionEndAddrs[Section->Index]);
    }

    // Update symbol virtual addresses and names. Must be done prior to updating sections.
    for (UINTN i=0; i < Elf->NumSymbols; i++) {
        SYS_ELF_SYMBOL *Sym = &Elf->Symbols[i];
        if (Sym->Vaddr != 0 && Sym->Section != NULL) {
            UINTN SectionAddr = Params->SectionStartAddrs[Sym->Section->Index];
            if (SectionAddr == 0) {
                PANIC_EXIT("elf: no relocation for symbol section", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                return;
            }
            UINTN SectionOffset = Sym->Vaddr - Sym->Section->Vaddr;
            Sym->Vaddr = SectionAddr + SectionOffset;
            if (Sym->Vaddr > Params->SectionEndAddrs[Sym->Section->Index]) {
                PANIC_EXIT("elf: relocation outside section", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                return;
            }
            if (Sym->Type != ElfSymSection) {
                UINTN ElfOffset = (UINTN) Sym->Name - (UINTN) Elf->ElfBytes;
                UINTN StrTabOffset = ElfOffset - Elf->StrTabSection->Offset;
                Sym->Name = (const CHAR8*) Params->SectionStartAddrs[Elf->StrTabSection->Index] + StrTabOffset;
            }
        }
    }

    // Update section virtual addresses.
    for (UINTN i=0; i < Elf->NumSectionPointers; i++) {
        SYS_ELF_SECTION *Sec = Elf->SectionPointerTable[i];
        if (Sec == NULL) {
            continue;
        }
        UINTN SectionAddr = Params->SectionStartAddrs[Sec->Index];
        if (SectionAddr == 0) {
            PANIC_EXIT("elf: no relocation for section", SYS_STATUS_FAIL, Context->Conf->DebugPort);
            return;
        }
        Sec->Vaddr = SectionAddr;
    }

    // Update external reference names.
    for (UINTN i=0; i < Elf->NumExternalRefs; i++) {
        UINTN ElfOffset = (UINTN) Elf->ExternalRefs[i] - (UINTN) Elf->ElfBytes;
        UINTN StrTabOffset = ElfOffset - Elf->StrTabSection->Offset;
        Elf->ExternalRefs[i] = (const CHAR8*) Params->SectionStartAddrs[Elf->StrTabSection->Index] + StrTabOffset;
    }
    
    // Discard temp memory.
    Context->BootStackTipAddr = MemTipAddr;

    // Update hash table with provided addresses.
    for (UINTN i=0; i < Elf->NumSymbols; i++) {
        SYS_ELF_SYMBOL *Sym = &Elf->Symbols[i];
        if (Sym->IsGlobal && Sym->Type == ElfSymFunc) {
            void *SymMem;
            if (Sys_Htable_Get(Htable, Sym->Name, &SymMem) && (UINTN) SymMem != Sym->Vaddr) {
                SYS_SERIAL_LOG("elf: sym: ", Context->Conf->DebugPort);
                SYS_SERIAL_LOG(Sym->Name, Context->Conf->DebugPort);
                SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
                PANIC_EXIT("elf: duplicate symbol", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                return;
            }
            if (!Sys_Htable_Put(Htable, Sym->Name, (void*) Sym->Vaddr)){
                PANIC_EXIT("elf: failed to put htable entry", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                return;
            }
        }
    }
}

// Parses the symbol table and inserts symbol structures into the ELF library.
// Panics on failure.
static void Sys_Elf_InsertSymbols(IN CONST UINTN SectionIndex, IN const UINTN ShdrOffset,
    IN OUT SYS_ELF_LIB *Elf, IN OUT SYS_BOOT_CONTEXT *Context) {

    const ELF_SHDR *SymTabShdr = (const ELF_SHDR*) &Elf->ElfBytes[ShdrOffset + SectionIndex * sizeof(ELF_SHDR)];
    const ELF_SHDR *LinkShdr = (const ELF_SHDR*) &Elf->ElfBytes[ShdrOffset + SymTabShdr->Link * sizeof(ELF_SHDR)];
    const UINTN NumSymbols = SymTabShdr->Size / SymTabShdr->EntrySize;
    const UINT8 *SymTabBytes = &Elf->ElfBytes[SymTabShdr->Offset];
    const CHAR8 *StrTabBytes = (const CHAR8*) &Elf->ElfBytes[LinkShdr->Offset];

    Elf->Symbols = (SYS_ELF_SYMBOL*) Sys_Init_AllocBoot(Context, sizeof(SYS_ELF_SYMBOL) * NumSymbols, CPU_STACK_ALIGNMENT);
    Elf->NumSymbols = 0;
    if (Elf->Symbols == NULL) {
        PANIC_EXIT("elf: not enough symbol memory", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return;
    }

    const ELF_SYMTAB *Sym;
    for (UINTN i=0; i < NumSymbols; i++) {
        Sym = (const ELF_SYMTAB*) SymTabBytes;
        SymTabBytes += SymTabShdr->EntrySize;

        const CHAR8 *SymName = StrTabBytes + Sym->NameOffset;
        UINT8 Attr = (Sym->Info & 0xF0) >> 4;
        UINT8 Type = Sym->Info & 0X0F;
        SYS_ELF_SECTION *Section = Elf->SectionPointerTable[Sym->SectionIndex];
        if (Type == ElfSymSection) {
            if (Section != NULL) {
                SymName = Section->Name;
            }
        }

        if (Sys_Common_AsciiStrCmp(SymName, "") == 0 && i > 0) {
            SYS_SERIAL_LOG("elf: nameless type: ", Context->Conf->DebugPort);
            SYS_SERIAL_LOG_INT(Type, TRUE, Context->Conf->DebugPort);
            SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
            PANIC_EXIT("elf: nameless symbol", SYS_STATUS_FAIL, Context->Conf->DebugPort);
            return;
        }
        if (Section == NULL && i > 0) {
            if (Attr != ELF_STB_GLOBAL || Type != ElfSymNone){
                SYS_SERIAL_LOG("elf: sym name: ", Context->Conf->DebugPort);
                SYS_SERIAL_LOG(SymName, Context->Conf->DebugPort);
                SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
                PANIC_EXIT("elf: sectionless local symbol", SYS_STATUS_FAIL, Context->Conf->DebugPort);
                return;
            }
        }

        SYS_ELF_SYMBOL *S = &Elf->Symbols[Elf->NumSymbols++];
        S->Name = SymName;
        S->Index = i;
        S->IsGlobal = Attr == ELF_STB_GLOBAL || Attr == ELF_STB_WEAK;
        S->Type = Type;
        S->Section = Section;
        S->Vaddr = Sym->Value;
        S->Size = Sym->Size;
    }

}

// Parses the symbol relocation table and inserts relocation structures for symbols into the ELF library.
// Panics on failure.
static void Sys_Elf_InsertSymRelocs(IN CONST UINTN SectionIndex, IN const UINTN ShdrOffset,
    IN OUT SYS_ELF_LIB *Elf, IN OUT SYS_BOOT_CONTEXT *Context) {

    const ELF_SHDR *RelocShdr = (const ELF_SHDR*) &Elf->ElfBytes[ShdrOffset + SectionIndex * sizeof(ELF_SHDR)];
    const UINTN NumRelocs = RelocShdr->Size / RelocShdr->EntrySize;
    const UINT8 *SymRelocBytes = &Elf->ElfBytes[RelocShdr->Offset];

    Elf->SymRelocations = (SYS_ELF_RELOCATION*) Sys_Init_AllocBoot(Context, sizeof(SYS_ELF_RELOCATION) * NumRelocs, CPU_STACK_ALIGNMENT);
    Elf->NumSymRelocations = 0;
    if (Elf->SymRelocations == NULL) {
        PANIC_EXIT("elf: not enough symbol relocation memory", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return;
    }
    UINTN RelocationsBaseAddr = (UINTN) Elf->SymRelocations;

    const ELF_RELA *Reloc;
    for (UINTN i=0; i < NumRelocs; i++) {
        Reloc = (const ELF_RELA*) SymRelocBytes;
        SymRelocBytes += RelocShdr->EntrySize;
        UINT64 SymIndex = Reloc->Info >> 32;
        UINT64 Type = Reloc->Info & 0xFFFFFFFF;
        if (Type == ElfRelocX86_64_NONE) {
            continue;
        }
        SYS_ELF_SYMBOL *Symbol = &Elf->Symbols[SymIndex];
        if (Symbol == NULL) {
            // Skip adding relocation if referenced symbol isn't loaded.
            continue;
        }
        SYS_ELF_RELOCATION *R = &Elf->SymRelocations[Elf->NumSymRelocations++];
        R->Symbol = Symbol;
        R->Type = Type;
        R->Vaddr = Reloc->Offset;
        R->Addend = (INTN) Reloc->Addend;
    }

    // Discard excess relocation space.
    Context->BootStackTipAddr = RelocationsBaseAddr + sizeof(SYS_ELF_RELOCATION) * Elf->NumSymRelocations;
}

// Parses the PLT relocation table and inserts relocation structures for PLT entries into the ELF library.
// Panics on failure.
static void Sys_Elf_InsertPltRelocs(IN CONST UINTN SectionIndex, IN const UINTN ShdrOffset,
    IN OUT SYS_ELF_LIB *Elf, IN OUT SYS_BOOT_CONTEXT *Context) {

    const ELF_SHDR *RelocShdr = (const ELF_SHDR*) &Elf->ElfBytes[ShdrOffset + SectionIndex * sizeof(ELF_SHDR)];
    const UINTN NumRelocs = RelocShdr->Size / RelocShdr->EntrySize;
    const UINT8 *PltRelocBytes = &Elf->ElfBytes[RelocShdr->Offset];

    Elf->PltRelocations = (SYS_ELF_RELOCATION*) Sys_Init_AllocBoot(Context, sizeof(SYS_ELF_RELOCATION) * NumRelocs, CPU_STACK_ALIGNMENT);
    Elf->NumPltRelocations = 0;
    if (Elf->PltRelocations == NULL) {
        PANIC_EXIT("elf: not enough PLT relocation memory", SYS_STATUS_FAIL, Context->Conf->DebugPort);
        return;
    }
    UINTN RelocationsBaseAddr = (UINTN) Elf->PltRelocations;

    const ELF_RELA *Reloc;
    for (UINTN i=0; i < NumRelocs; i++) {
        Reloc = (const ELF_RELA*) PltRelocBytes;
        PltRelocBytes += RelocShdr->EntrySize;
        UINT64 SymIndex = Reloc->Info >> 32;
        UINT64 Type = Reloc->Info & 0xFFFFFFFF;
        if (Type != ElfRelocX86_64_JMP_SLOT) {
            SYS_SERIAL_LOG("elf: type: ", Context->Conf->DebugPort);
            SYS_SERIAL_LOG_INT(Type, TRUE, Context->Conf->DebugPort);
            SYS_SERIAL_LOG("\n", Context->Conf->DebugPort);
            PANIC_EXIT("elf: unknown plt reloc type", SYS_STATUS_FAIL, Context->Conf->DebugPort);
            return;
        }
        SYS_ELF_RELOCATION *R = &Elf->PltRelocations[Elf->NumPltRelocations++];
        R->Symbol = &Elf->Symbols[SymIndex];
        R->Type = Type;
        R->Vaddr = Reloc->Offset;
        R->Addend = (INTN) Reloc->Addend;
    }

    // Discard excess relocation space.
    Context->BootStackTipAddr = RelocationsBaseAddr + sizeof(SYS_ELF_RELOCATION) * Elf->NumPltRelocations;
}

static BOOLEAN Sys_Elf_IsSectionSymTab(IN const SYS_ELF_SECTION *Section) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".symtab") == 0
        && Section->Type == ElfSectSymTab) {
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionRelaText(IN const SYS_ELF_SECTION *Section) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".rela.text") == 0
        && Section->Type == ElfSectRela) {
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionRelaPlt(IN const SYS_ELF_SECTION *Section, IN const UINTN PltRelocVaddr) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".rela.plt") == 0
        && Section->Type == ElfSectRela && Section->Vaddr == PltRelocVaddr) {
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionRelaEhFrame(IN const SYS_ELF_SECTION *Section) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".rela.eh_frame") == 0
        && Section->Type == ElfSectRela) {
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionStrTab(IN const SYS_ELF_SECTION *Section, IN const UINTN StrTabVaddr) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".strtab") == 0
        && Section->Type == ElfSectStrTab && Section->Vaddr == StrTabVaddr) {
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionShStrTab(IN const SYS_ELF_SECTION *Section) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".shstrtab") == 0
        && Section->Type == ElfSectStrTab) {
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionDynStrTab(IN const SYS_ELF_SECTION *Section) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".dynstr") == 0
        && Section->Type == ElfSectStrTab) {
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionDynSym(IN const SYS_ELF_SECTION *Section) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".dynsym") == 0
        && Section->Type == ElfSectDynSym) {
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionDynamic(IN const SYS_ELF_SECTION *Section, IN const UINTN DynamicVaddr) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".dynamic") == 0
        && Section->Type == ElfSectDynamic && Section->Vaddr == DynamicVaddr) {
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionBss(IN const SYS_ELF_SECTION *Section) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".bss") == 0
        && Section->Type == ElfSectNoBits && Section->IsAllocated && Section->IsWritable){
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionEhFrame(IN const SYS_ELF_SECTION *Section) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".eh_frame") == 0
        && Section->Type == ElfSectProgBits){
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionData(IN const SYS_ELF_SECTION *Section) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".data") == 0
        && Section->Type == ElfSectProgBits && Section->IsAllocated && Section->IsWritable){
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionRoData(IN const SYS_ELF_SECTION *Section) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".rodata") == 0
        && Section->Type == ElfSectProgBits && Section->IsAllocated && !Section->IsWritable){
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionText(IN const SYS_ELF_SECTION *Section) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".text") == 0
        && Section->Type == ElfSectProgBits && Section->IsAllocated && Section->IsExecutable){
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionPlt(IN const SYS_ELF_SECTION *Section) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".plt") == 0
        && Section->Type == ElfSectProgBits && Section->IsAllocated && Section->IsExecutable){
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_IsSectionGot(IN const SYS_ELF_SECTION *Section, IN const UINTN GotVaddr) {
    if (Section != NULL && Sys_Common_AsciiStrCmp(Section->Name, ".got") == 0
        && Section->Type == ElfSectProgBits && Section->IsAllocated && Section->Vaddr == GotVaddr){
        return TRUE;
    }
    return FALSE;
}

static BOOLEAN Sys_Elf_Rel_PC32(IN OUT ELF_RELOC_PARAMS *Params) {

    // PC-relative references must be made by an instruction pointer within the text section.
    if (Params->Reloc->Vaddr < Params->Elf->TextSection->Vaddr
        || Params->Reloc->Vaddr > Params->Elf->TextSection->Vaddr + Params->Elf->TextSection->SizeFile){
        SYS_SERIAL_LOG("elf: PC-relative relocation outside text\n", Params->DebugPort);
        return FALSE;
    }

    // Calculate the offset within the relocated PC (text) section where this symbol is referenced, and
    // get the target within the file bytes.
    SYS_ELF_SECTION *PcSection = Params->Elf->TextSection;
    UINTN PcSectionIndex = PcSection->Index;
    UINTN PcOffset = Params->Reloc->Vaddr - PcSection->Vaddr;
    UINTN RelocPcRefAddr = Params->SectionStartAddrs[PcSectionIndex] + PcOffset;
    UINTN TargetBytesOffset = PcSection->Offset + PcOffset;
    INT32 *RelocTarget = (INT32*) &Params->Elf->ElfBytes[TargetBytesOffset];
    UINTN ExpectedPreRelocSymbolAddr = Params->Reloc->Vaddr + *RelocTarget - Params->Reloc->Addend;
    if (Params->Reloc->Symbol->Vaddr != ExpectedPreRelocSymbolAddr){
        SYS_SERIAL_LOG("elf: unexpected PC32 addr\n", Params->DebugPort);
        return FALSE;
    }

    // Use the symbol type to determine the address and relocated address.
    UINTN SectionOffset = 0;
    if (Params->Reloc->Symbol->Section != NULL) {
        SectionOffset = Params->Reloc->Symbol->Vaddr - Params->Reloc->Symbol->Section->Vaddr;
    }
    UINTN RelocSymbolAddr = 0;
    BOOLEAN RelocFound = FALSE;
    switch (Params->Reloc->Symbol->Type) {
    default:
        RelocFound = FALSE;
        break;

    case ElfSymSection:
    case ElfSymObject:
    case ElfSymFunc:
        if (Sys_Elf_IsSectionData(Params->Reloc->Symbol->Section)){
            RelocSymbolAddr = Params->SectionStartAddrs[Params->Elf->DataSection->Index] + SectionOffset;
            RelocFound = TRUE;
        } else if (Sys_Elf_IsSectionRoData(Params->Reloc->Symbol->Section)){
            RelocSymbolAddr = Params->SectionStartAddrs[Params->Elf->RoDataSection->Index] + SectionOffset;
            RelocFound = TRUE;
        } else if (Sys_Elf_IsSectionText(Params->Reloc->Symbol->Section)){
            RelocSymbolAddr = Params->SectionStartAddrs[Params->Elf->TextSection->Index] + SectionOffset;
            RelocFound = TRUE;
        } else if (Sys_Elf_IsSectionBss(Params->Reloc->Symbol->Section)){
            RelocSymbolAddr = Params->SectionStartAddrs[Params->Elf->BssSection->Index] + SectionOffset;
            RelocFound = TRUE;
        }
        break;
    }
    if (!RelocFound) {
        SYS_SERIAL_LOG("elf: unhandled PC32 relocation for ", Params->DebugPort);
        SYS_SERIAL_LOG(Params->Reloc->Symbol->Name, Params->DebugPort);
        SYS_SERIAL_LOG(", type ", Params->DebugPort);
        SYS_SERIAL_LOG_INT(Params->Reloc->Symbol->Type, TRUE, Params->DebugPort);
        SYS_SERIAL_LOG("\n", Params->DebugPort);
        return FALSE;
    }
    if (RelocSymbolAddr == SectionOffset) {
        SYS_SERIAL_LOG("elf: no reloc addr for ", Params->DebugPort);
        SYS_SERIAL_LOG(Params->Reloc->Symbol->Name, Params->DebugPort);
        SYS_SERIAL_LOG("\n", Params->DebugPort);
        return FALSE;
    }

    // Get the relocated symbol address relative to the relocated PC, save and verify that it saved correctly.
    INTN RelocRel = RelocSymbolAddr - RelocPcRefAddr + Params->Reloc->Addend;
    *RelocTarget = (INT32) RelocRel;
    if (*RelocTarget != RelocRel) {
        SYS_SERIAL_LOG("elf: bad PC32 relocation for ", Params->DebugPort);
        SYS_SERIAL_LOG(Params->Reloc->Symbol->Name, Params->DebugPort);
        SYS_SERIAL_LOG("\n", Params->DebugPort);
        return FALSE;
    }

    return TRUE;
}


static BOOLEAN Sys_Elf_Rel_PLT32(IN OUT ELF_RELOC_PARAMS *Params) {

    // PLT-relative references must be made by an instruction pointer within the text section.
    if (Params->Reloc->Vaddr < Params->Elf->TextSection->Vaddr
        || Params->Reloc->Vaddr > Params->Elf->TextSection->Vaddr + Params->Elf->TextSection->SizeFile){
        SYS_SERIAL_LOG("elf: PLT-relative relocation outside text\n", Params->DebugPort);
        return FALSE;
    }

    // Calculate the offset within the relocated PC (text) section where this symbol is referenced, and
    // get the target within the file bytes.
    SYS_ELF_SECTION *PcSection = Params->Elf->TextSection;
    UINTN PcSectionIndex = PcSection->Index;
    UINTN PcOffset = Params->Reloc->Vaddr - PcSection->Vaddr;
    UINTN RelocPcRefAddr = Params->SectionStartAddrs[PcSectionIndex] + PcOffset;
    UINTN TargetBytesOffset = PcSection->Offset + PcOffset;
    INT32 *RelocTarget = (INT32*) &Params->Elf->ElfBytes[TargetBytesOffset];

    // Get the relocated address of the position in the PLT being called.
    UINTN PltTargetVaddr = Params->Reloc->Vaddr + *RelocTarget - Params->Reloc->Addend;
    if (PltTargetVaddr < Params->Elf->PltSection->Vaddr
        || PltTargetVaddr > Params->Elf->PltSection->Vaddr + Params->Elf->PltSection->SizeFile){
        SYS_SERIAL_LOG("elf: PLT call outside section\n", Params->DebugPort);
        return FALSE;
    }
    UINTN PltTargetOffset = PltTargetVaddr - Params->Elf->PltSection->Vaddr;
    UINTN RelocPltTargetVaddr = Params->SectionStartAddrs[Params->Elf->PltSection->Index] + PltTargetOffset;

    // Get the relocated PLT target address relative to the relocated PC, save and verify that it saved correctly.
    INTN RelocRel = RelocPltTargetVaddr - RelocPcRefAddr + Params->Reloc->Addend;
    *RelocTarget = (INT32) RelocRel;
    if (*RelocTarget != RelocRel) {
        SYS_SERIAL_LOG("elf: bad PLT32 relocation for ", Params->DebugPort);
        SYS_SERIAL_LOG(Params->Reloc->Symbol->Name, Params->DebugPort);
        SYS_SERIAL_LOG("\n", Params->DebugPort);
        return FALSE;
    }

    // Set the symbol name in the PLT jump names.
    UINTN PltIndex = PltTargetOffset / ELF_X86_64_PLT_ENTRY_SIZE;
    Params->PltJumpNames[PltIndex] = Params->Reloc->Symbol->Name;

    return TRUE;
}


static BOOLEAN Sys_Elf_Rel_JMP_SLOT(IN OUT ELF_RELOC_PARAMS *Params) {

    // Jump slot references must be made from inside the GOT section.
    if (Params->Reloc->Vaddr < Params->Elf->GotSection->Vaddr
        || Params->Reloc->Vaddr > Params->Elf->GotSection->Vaddr + Params->Elf->GotSection->SizeFile){
        SYS_SERIAL_LOG("elf: jmp relocation outside got\n", Params->DebugPort);
        return FALSE;
    }

    // Calculate the offsets within the relocated GOT/PLT sections where this symbol is referenced,
    // get the target within the file bytes, and the GOT/PLT indices.
    UINTN GotOffset = Params->Reloc->Vaddr - Params->Elf->GotSection->Vaddr;
    UINTN GotTargetBytesOffset = Params->Elf->GotSection->Offset + GotOffset;
    UINT64 *GotRelocTarget = (UINT64*) &Params->Elf->ElfBytes[GotTargetBytesOffset];
    UINTN GotIndex = GotOffset / ELF_X86_64_GOT_ENTRY_SIZE;
    UINTN PltOffset = *GotRelocTarget - ELF_X86_64_JUMP_ADDEND - Params->Elf->PltSection->Vaddr;
    UINTN PltIndex = PltOffset / ELF_X86_64_PLT_ENTRY_SIZE;
    if (PltIndex != GotIndex - 2) {  // First 2 GOT symbols are reserved.
        SYS_SERIAL_LOG("elf: jmp relocation bad index\n", Params->DebugPort);
        return FALSE;
    }
    UINTN PltTargetBytesOffset = Params->Elf->PltSection->Offset + PltOffset + ELF_X86_64_JUMP_OP_SIZE;
    INT32 *PltRelocTarget = (INT32*) &Params->Elf->ElfBytes[PltTargetBytesOffset];
    UINTN PltJumpGotVaddr = Params->Elf->PltSection->Vaddr + PltOffset + *PltRelocTarget + ELF_X86_64_JUMP_ADDEND;
    if (PltJumpGotVaddr != Params->Reloc->Vaddr){
        SYS_SERIAL_LOG("elf: nonmatching GOT reloc address\n", Params->DebugPort);
        return FALSE;
    }

    // Get the external symbol address.
    const CHAR8 *SymName = Params->PltJumpNames[PltIndex];
    if (SymName == NULL) {
        SYS_SERIAL_LOG("elf: no jmp name\n", Params->DebugPort);
        return FALSE;
    }
    void *SymFn;
    if (!Sys_Htable_Get(Params->Htable, SymName, &SymFn)) {
        SYS_SERIAL_LOG("elf: no jmp symbol ", Params->DebugPort);
        SYS_SERIAL_LOG(SymName, Params->DebugPort);
        SYS_SERIAL_LOG("\n", Params->DebugPort);
        return FALSE;
    }
    UINTN SymAddr = (UINTN) SymFn;

    // Save the GOT value and verify that it saved correctly.
    *GotRelocTarget = (UINT64) SymAddr;
    if (*GotRelocTarget != SymAddr) {
        SYS_SERIAL_LOG("elf: bad jmp GOT relocation for ", Params->DebugPort);
        SYS_SERIAL_LOG(SymName, Params->DebugPort);
        SYS_SERIAL_LOG("\n", Params->DebugPort);
        return FALSE;
    }

    // Get the relocated address of the PLT entry target (the address of the GOT entry).
    UINTN RelocPltRefVaddr = Params->SectionStartAddrs[Params->Elf->PltSection->Index] + PltOffset;
    UINTN RelocPltJumpGotVaddr = Params->SectionStartAddrs[Params->Elf->GotSection->Index] + GotOffset;
    INTN PltRelocRel = RelocPltJumpGotVaddr - RelocPltRefVaddr - ELF_X86_64_JUMP_ADDEND;

    // Save the PLT value and verify that it saved correctly.
    *PltRelocTarget = (INT32) PltRelocRel;
    if (*PltRelocTarget != PltRelocRel) {
        SYS_SERIAL_LOG("elf: bad jmp PLT relocation for ", Params->DebugPort);
        SYS_SERIAL_LOG(SymName, Params->DebugPort);
        SYS_SERIAL_LOG("\n", Params->DebugPort);
        return FALSE;
    }
    
    return TRUE;
}

