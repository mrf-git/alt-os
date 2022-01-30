#if defined(__amd64__) || defined(__x86__)

#include "Serial.h"


// Intel requires the value in the A register and port in DX to enable the full range of port values.
#define SYS_SERIAL_OUTB(Val, Port) __asm__("outb %%al, %%dx": :"a"((UINT8) Val), "d"(Port))


void SYSABI Sys_Serial_Reset() {
    // Nothing needed.
}


void SYSABI Sys_Serial_Write(IN const UINT8 *Buffer, IN const UINTN BufferLen) {
    for (UINTN i=0; i < BufferLen; i++) {
        SYS_SERIAL_OUTB(Buffer[i], 0x2e8);  // Output to COM4.
    }
}



#endif // defined( x86 )
