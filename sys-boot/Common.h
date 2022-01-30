#ifndef SYS_BOOT_COMMON_H
#define SYS_BOOT_COMMON_H

#include <ProcessorBind.h>

#define SYSEXPORT                       __attribute__((visibility("default")))

#define SYS_STATUS                      __INTPTR_TYPE__
#define SYS_STATUS_SUCESS               ((SYS_STATUS) 0)
#define SYS_IS_ERROR(Status)            (((SYS_STATUS) Status) != SYS_STATUS_SUCESS)
#define SYS_STATUS_FAIL                 ((SYS_STATUS) -1)

#ifndef ALIGN_VALUE
    #define ALIGN_VALUE(Value,Align)    ((Value) + (((Align) - (Value)) & ((Align) - 1)))
#endif

#ifndef IN
    #define IN
#endif
#ifndef OUT
    #define OUT
#endif

#ifndef NULL
    #define NULL                        ((void*) 0)
#endif

#define SYS_GUID_SIZE                   16


#ifdef SYSABI
    #undef SYSABI
#endif

#define SYSABI                          __attribute__((ms_abi))

#define SYS_MAX_STR_LEN                 8192
#define SYS_MAX_INT_ASCII_LEN           32

#ifndef TRUE
    #define TRUE  ((BOOLEAN)(1==1))
#endif
#ifndef FALSE
    #define FALSE ((BOOLEAN)(0==1))
#endif


//
// Library functions.
//

//
// Computes the length of the null-terminated ASCII string.
//
SYSEXPORT UINTN SYSABI Sys_Common_AsciiStrLen(IN const CHAR8 *Str);

//
// Compares two null-terminated ASCII strings.
//
SYSEXPORT INTN SYSABI Sys_Common_AsciiStrCmp(IN const CHAR8 *Str1, IN const CHAR8 *Str2);

//
// Copies the source null-terminated ASCII string to the out buffer and returns the number of bytes copied.
//
SYSEXPORT UINTN SYSABI Sys_Common_AsciiStrCopy(IN const CHAR8 *Str, OUT CHAR8 *Out);

//
// Writes the integer Value to an ASCII string in Buffer.
//
SYSEXPORT void SYSABI Sys_Common_IntToAscii(IN const INTN Value, IN const BOOLEAN IsUnsigned,
    IN const UINTN Base, OUT CHAR8 *Buffer);

//
// Copies the specified memory to the out buffer byte-by-byte.
//
SYSEXPORT void SYSABI Sys_Common_MemCopy(IN const void *Mem, IN const UINTN Len, OUT const void *Buffer);


//
// Calculates a pseudorandom hash for the given string using the specified seed and modulus.
//
SYSEXPORT UINTN SYSABI Sys_Common_StringHash(IN const CHAR8 *String, IN const UINTN Seed, IN const UINTN Mod);


#endif // SYS_BOOT_COMMON_H
