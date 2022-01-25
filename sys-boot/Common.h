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
// Writes the contents of the specified buffer to the specified serial port.
//
SYSEXPORT void SYSABI Sys_Common_WriteSerial(IN const UINTN Port, IN const UINT8 *Buffer, IN const UINTN BufferLen);


//
// Formats the specified integer as an ASCII string and writes the contents to the specified serial port.
//
SYSEXPORT void SYSABI Sys_Common_WriteSerialInt(IN const UINTN Port, IN const INTN Int,
    IN const BOOLEAN IsUnsigned, IN const UINTN Base);


//
// Causes the system to dump the contents of the specified buffer to the specified serial port and then hang forever.
//
SYSEXPORT void SYSABI Sys_Common_Panic(IN const UINTN Port, IN const UINT8 *Buffer, IN const UINTN BufferLen);


//
// Panics with the specified ASCII message and error status code.
//
SYSEXPORT void SYSABI Sys_Common_PanicError(IN const UINTN Port, IN const CHAR8 *Message, IN const INTN ErrorStatus);


//
// Calculates a pseudorandom hash for the given string using the specified seed and modulus.
//
SYSEXPORT UINTN SYSABI Sys_Common_StringHash(IN const CHAR8 *String, IN const UINTN Seed, IN const UINTN Mod);


//
// Macro definitions.
//

#define SYS_SERIAL_LOG(Str,Port) \
    Sys_Common_WriteSerial((UINTN) Port, (UINT8*) Str, Sys_Common_AsciiStrLen((CHAR8*) Str))

#define SYS_SERIAL_LOG_INT(Int,IsUnsigned,Port) \
    Sys_Common_WriteSerialInt((UINTN) Port, (INTN) Int, (BOOLEAN) IsUnsigned, 10)

#define SYS_SERIAL_LOG_INT_HEX(Int,IsUnsigned,Port) \
    Sys_Common_WriteSerialInt((UINTN) Port, (INTN) Int, (BOOLEAN) IsUnsigned, 16)

#define PANIC_EXIT(FailMessage,Status,Port) \
    Sys_Common_PanicError((UINTN) Port, (CHAR8*) FailMessage, (INTN) Status)


#endif // SYS_BOOT_COMMON_H
