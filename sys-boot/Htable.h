#ifndef SYS_BOOT_HTABLE_H
#define SYS_BOOT_HTABLE_H

#include "Common.h"
#include "Memory.h"


//
// ELF definitions.
//

#define SYS_ELF_MAX_SECTION_NAME                512
#define SYS_ELF_MAX_INTERP_LEN                  16


//
// Type enums.
//

typedef enum {
    ElfSymNone,
    ElfSymObject,
    ElfSymFunc,
    ElfSymSection,
} SYS_ELF_SYMBOL_TYPE;

typedef enum {
    ElfSectNull,
    ElfSectProgBits,
    ElfSectSymTab,
    ElfSectStrTab,
    ElfSectRela,
    ElfSectHash,
    ElfSectDynamic,
    ElfSectNote,
    ElfSectNoBits,
    ElfSectRel,
    ElfSectShLib,
    ElfSectDynSym,
} SYS_ELF_SECTION_TYPE;

typedef enum {
    ElfDynNull,
    ElfDynNeeded,
    ElfDynPltRelSize,
    ElfDynPltGot,
    ElfDynHash,
    ElfDynStrTab,
    ElfDynSymTab,
    ElfDynRela,
    ElfDynRelaSize,
    ElfDynRelaEntrySize,
    ElfDynStrTabSize,
    ElfDynSymTabEntrySize,
    ElfDynInit,
    ElfDynFini,
    ElfDynSoName,
    ElfDynRpath,
    ElfDynSymbolic,
    ElfDynRel,
    ElfDynRelSize,
    ElfDynRelEntrySize,
    ElfDynPltRel,
    ElfDynDebug,
    ElfDynTextRel,
    ElfDynJumpRel,
    ElfDynBindNow,
    ElfDynInitArray,
    ElfDynFiniArray,
    ElfDynInitArraySize,
    ElfDynFiniArraySize,
    ElfDynRunPath,
    ElfDynFlags,
    ElfDynEncoding,
    ElfDynPreInitArray,
    ElfDynPreInitArraySize,
    ElfDynMaxPosTags,

} SYS_ELF_DYNAMIC_TYPE;

typedef enum {
    ElfRelocX86_64_NONE,                 // No relocation
    ElfRelocX86_64_64,                   // Add 64 bit symbol value
    ElfRelocX86_64_PC32,                 // PC-relative 32 bit signed sym value
    ElfRelocX86_64_GOT32,                // PC-relative 32 bit GOT offset
    ElfRelocX86_64_PLT32,                // PC-relative 32 bit PLT offset
    ElfRelocX86_64_COPY,                 // Copy data from shared object
    ElfRelocX86_64_GLOB_DAT,             // Set GOT entry to data address
    ElfRelocX86_64_JMP_SLOT,             // Set GOT entry to code address
    ElfRelocX86_64_RELATIVE,             // Add load address of shared object
    ElfRelocX86_64_GOTPCREL,             // Add 32 bit signed pcrel offset to GOT
    ElfRelocX86_64_32,                   // Add 32 bit zero extended symbol value
    ElfRelocX86_64_32S,                  // Add 32 bit sign extended symbol value
    ElfRelocX86_64_16,                   // Add 16 bit zero extended symbol value
    ElfRelocX86_64_PC16,                 // Add 16 bit signed extended pc relative symbol value
    ElfRelocX86_64_8,                    // Add 8 bit zero extended symbol value
    ElfRelocX86_64_PC8,                  // Add 8 bit signed extended pc relative symbol value
    ElfRelocX86_64_DTPMOD64,             // ID of module containing symbol
    ElfRelocX86_64_DTPOFF64,             // Offset in TLS block
    ElfRelocX86_64_TPOFF64,              // Offset in static TLS block
    ElfRelocX86_64_TLSGD,                // PC relative offset to GD GOT entry
    ElfRelocX86_64_TLSLD,                // PC relative offset to LD GOT entry
    ElfRelocX86_64_DTPOFF32,             // Offset in TLS block
    ElfRelocX86_64_GOTTPOFF,             // PC relative offset to IE GOT entry
    ElfRelocX86_64_TPOFF32,              // Offset in static TLS block
    ElfRelocX86_64_PC64,                 // PC relative 64 bit
    ElfRelocX86_64_GOTOFF64,             // 64 bit offset to GOT
    ElfRelocX86_64_GOTPC3,               // 32 bit signed pc relative offset to GOT
    ElfRelocX86_64_GOT64,                // 64-bit GOT entry offset
    ElfRelocX86_64_GOTPCREL64,           // 64-bit PC relative offset to GOT entry
    ElfRelocX86_64_GOTPC64,              // 64-bit PC relative offset to GOT
    ElfRelocX86_64_GOTPLT64,             // like GOT64, says PLT entry needed
    ElfRelocX86_64_PLTOFF64,             // 64-bit GOT relative offset to PLT entry
    ElfRelocX86_64_SIZE32,               // Size of symbol plus 32-bit addend
    ElfRelocX86_64_SIZE64,               // Size of symbol plus 64-bit addend
    ElfRelocX86_64_GOTPC32_TLSDESC,      // GOT offset for TLS descriptor
    ElfRelocX86_64_TLSDESC_CALL,         // Marker for call through TLS descriptor
    ElfRelocX86_64_TLSDESC,              // TLS descriptor
    ElfRelocX86_64_IRELATIVE,            // Adjust indirectly by program base
    ElfRelocX86_64_RELATIVE64,           // 64-bit adjust by program base
    ElfRelocX86_64_GOTPCRELX,            // Load from 32 bit signed pc relative offset to GOT entry without REX prefix, relaxable
    ElfRelocX86_64_REX_GOTPCRELX,        // Load from 32 bit signed pc relative offset to GOT entry with REX prefix, relaxable
} SYS_ELF_RELOCATION_TYPE;


//
// ELF structures.
//

// Represents a single ELF section read into memory.
typedef struct {
    const CHAR8 *Name;
    UINTN Index;
    SYS_ELF_SECTION_TYPE Type;
    BOOLEAN IsWritable;
    BOOLEAN IsExecutable;
    BOOLEAN IsAllocated;
    UINTN Vaddr;
    UINTN Offset;
    UINTN SizeFile;
    UINTN Align;
    UINTN EntrySize;
} SYS_ELF_SECTION;

// Represents a single ELF relocatable symbol.
typedef struct {
    const CHAR8 *Name;
    UINTN Index;
    BOOLEAN IsGlobal;
    SYS_ELF_SYMBOL_TYPE Type;
    SYS_ELF_SECTION *Section;
    UINTN Vaddr;
    UINTN Size;
} SYS_ELF_SYMBOL;

// Represents a single ELF symbol relocation with a virtual address and addend.
typedef struct {
    SYS_ELF_SYMBOL *Symbol;
    SYS_ELF_RELOCATION_TYPE Type;
    UINTN Vaddr;
    INTN Addend;
} SYS_ELF_RELOCATION;

// Represents an ELF program library.
typedef struct {
    UINT8 *ElfBytes;
    UINTN ElfBytesSize;
    const CHAR8 *InterpName;
    SYS_ELF_SYMBOL *Symbols;
    UINTN NumSymbols;
    SYS_ELF_RELOCATION *SymRelocations;
    UINTN NumSymRelocations;
    SYS_ELF_RELOCATION *PltRelocations;
    UINTN NumPltRelocations;
    SYS_ELF_RELOCATION *EhRelocations;
    UINTN NumEhRelocations;
    const CHAR8 **ExternalRefs;
    UINTN NumExternalRefs;
    UINTN CodeMemorySize;
    UINTN DataMemorySize;
    UINTN RoDataMemorySize;
    UINTN NumPltEntries;
    UINTN NumGotEntries;
    UINTN NumSectionPointers;
    SYS_ELF_SECTION **SectionPointerTable;
    SYS_ELF_SECTION *StrTabSection;
    SYS_ELF_SECTION *TextSection;
    SYS_ELF_SECTION *DataSection;
    SYS_ELF_SECTION *RoDataSection;
    SYS_ELF_SECTION *BssSection;
    SYS_ELF_SECTION *PltSection;
    SYS_ELF_SECTION *GotSection;
    SYS_ELF_SECTION *DynamicSection;
    SYS_ELF_SECTION *EhFrameSection;

} SYS_ELF_LIB;


//
// Forward declarations.
//

// Opaque structure representing a constant-size hash table with string keys and data pointer values.
typedef struct _SYS_HASH_TABLE SYS_HASH_TABLE;


//
// Function pointer type definitions.
//

// Sys_Link_ReadElf
typedef SYS_ELF_LIB * SYSABI (*SYS_FN_LINK_READ_ELF)(IN OUT SYS_MEMORY_TABLE *MemoryTable, IN const UINT8 *ElfBytes,
    IN const UINTN ElfBytesSize, IN const UINTN DebugPort);


//
// Library functions.
//

//
// Gets the constant number of bytes required by the hash table.
//
SYSEXPORT const UINTN SYSABI Sys_Htable_Size();

//
// Initializes the specified hash table to prepare for inserting and lookup.
//
SYSEXPORT void SYSABI Sys_Htable_Init(IN OUT SYS_HASH_TABLE *Htable);

//
// Inserts or updates the data pointer at the specified key and returns TRUE if successful.
//
SYSEXPORT BOOLEAN SYSABI Sys_Htable_Put(IN OUT SYS_HASH_TABLE *Htable, IN const CHAR8 *Key, IN void *Value);

//
// Gets the data pointer at the specified key and returns TRUE, or returns FALSE if the key doesn't exist.
// If the Value parameter is NULL, returns TRUE only if the key exists.
//
SYSEXPORT BOOLEAN SYSABI Sys_Htable_Get(IN OUT SYS_HASH_TABLE *Htable, IN const CHAR8 *Key, OUT void **Value);

//
// Removes the data pointer at the specified key if it exists.
//
SYSEXPORT void SYSABI Sys_Htable_Remove(IN OUT SYS_HASH_TABLE *Htable, IN const CHAR8 *Key);

//
// Gets the number of current keys in the table.
//
SYSEXPORT UINTN SYSABI Sys_Htable_NumKeys(IN OUT SYS_HASH_TABLE *Htable);

//
// Gets the key at the specified Index or NULL if it doesn't exist.
//
SYSEXPORT const CHAR8 * SYSABI Sys_Htable_Key(IN OUT SYS_HASH_TABLE *Htable, IN const UINTN Index);


#endif // SYS_BOOT_HTABLE_H
