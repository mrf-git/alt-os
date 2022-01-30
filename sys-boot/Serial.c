#include "Serial.h"


static CHAR8 AsciiIntBuffer[SYS_MAX_INT_ASCII_LEN];


void SYSABI Sys_Serial_WriteInt(IN const INTN Int, IN const BOOLEAN IsUnsigned, IN const UINTN Base) {
    Sys_Common_IntToAscii((INTN) Int, IsUnsigned, Base, AsciiIntBuffer);
    Sys_Serial_Write((UINT8*)AsciiIntBuffer, Sys_Common_AsciiStrLen((CHAR8*) AsciiIntBuffer));
}
