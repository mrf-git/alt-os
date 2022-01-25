#ifndef SYS_BOOT_ELF_H
#define SYS_BOOT_ELF_H

#include "Common.h"
#include "OsInit.h"
#include "Htable.h"


//
// Reads ELF shared library bytes into a new structure allocated on the boot stack and
// returns a pointer to it, or panics on failure. The returned structure must be
// kept in memory together with ElfBytes to remain valid.
//
SYS_ELF_LIB * SYSABI Sys_Elf_ReadLib(IN const UINT8 *ElfBytes, IN const UINTN ElfBytesSize,
    IN OUT SYS_BOOT_CONTEXT *Context);

//
// Relocates the specified ELF library to the specified page-aligned memory addresses for
// Code, Data, and RoData. This is done directly on the in-memory ELF bytes. Next copies those
// segments to the specified memory and initializes them for execution. Finally updates the ELF library
// structure to reflect relocation, and inserts addresses to all global functions into the hash table.
// The ELF structure and the library file bytes can then be freed after return if relocation/symbols
// are no longer needed. Panics on failure.
//
void SYSABI Sys_Elf_LoadLib(IN const UINTN CodeAddr, IN const UINTN DataAddr, IN const UINTN RoDataAddr,
    IN OUT SYS_ELF_LIB *Elf, IN OUT SYS_HASH_TABLE *Htable, IN OUT SYS_BOOT_CONTEXT *Context);


#endif // SYS_BOOT_ELF_H
