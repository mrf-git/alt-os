#ifndef SYS_BOOT_SERIAL_H
#define SYS_BOOT_SERIAL_H

#include "Common.h"

//
// Resets the serial device and prepares it for I/O.
//
void SYSABI Sys_Serial_Reset();

//
// Writes the contents of the specified buffer to the boot serial device.
//
SYSEXPORT void SYSABI Sys_Serial_Write(IN const UINT8 *Buffer, IN const UINTN BufferLen);

//
// Formats the specified integer as an ASCII string and writes the contents to the boot serial device.
//
SYSEXPORT void SYSABI Sys_Serial_WriteInt(IN const INTN Int, IN const BOOLEAN IsUnsigned, IN const UINTN Base);


//
// Macro definitions.
//

#define SYS_SERIAL_LOG(Str) \
    Sys_Serial_Write((UINT8*) Str, Sys_Common_AsciiStrLen((CHAR8*) Str))

#define SYS_SERIAL_LOG_INT(Int,IsUnsigned) \
    Sys_Serial_WriteInt((INTN) Int, (BOOLEAN) IsUnsigned, 10)

#define SYS_SERIAL_LOG_INT_HEX(Int,IsUnsigned) \
    Sys_Serial_WriteInt((INTN) Int, (BOOLEAN) IsUnsigned, 16)


#endif // SYS_BOOT_SERIAL_H
