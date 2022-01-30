#if defined(__aarch64__) || defined(__arm64__)

#include "Serial.h"


// UART register offsets from the base address.
// See Chapter 3 of the PrimeCell UART (PL011) Technical Reference Manual.
#define UARTDR_RW               0x000
#define UARTRSR_RW              0x004
#define UARTECR_RW              0x004
#define UARTFR_RO               0x018
#define UARTILPR_RW             0x020
#define UARTIBRD_RW             0x024
#define UARTFBRD_RW             0x028
#define UARTLCR_H_RW            0x02C
#define UARTCR_RW               0x030
#define UARTIFLS_RW             0x034
#define UARTIMSC_RW             0x038
#define UARTRIS_RO              0x03C
#define UARTMIS_RO              0x040
#define UARTICR_WO              0x044
#define UARTDMACR_RW            0x048
#define UARTPeriphID0_RO        0xFE0
#define UARTPeriphID1_RO        0xFE4
#define UARTPeriphID2_RO        0xFE8
#define UARTPeriphID3_RO        0xFEC
#define UARTPCellID0_RO         0xFF0
#define UARTPCellID1_RO         0xFF4
#define UARTPCellID2_RO         0xFF8
#define UARTPCellID3_RO         0xFFC

#define UARTFR_BUSY             ((UINT16) (1 << 3))
#define UARTLCR_H_8WORD_1STOP   ((UINT8) 0x60)  // 8 bit word length and 1 stop bit.

#define UARTCR_UARTEN           ((UINT16) (1 << 0))
#define UARTCR_TXE              ((UINT16) (1 << 8))


static inline volatile void* Sys_Serial_Reg(UINTN Register) {
    UINTN BaseAddress = 0x9000000;
    UINTN Addr = BaseAddress + Register;
    return (volatile void *)Addr;
}


static inline void Sys_Serial_Wait() {
    while ((*((UINT16*)Sys_Serial_Reg(UARTFR_RO)) & UARTFR_BUSY) == UARTFR_BUSY) {}
}


void SYSABI Sys_Serial_Reset() {
    UINTN ClockRate = 0x16e3600;
    UINTN BaudRate = 115200;

    // Disable UART and wait for it to be ready.
    UINT16 CrVal = *(UINT16*)Sys_Serial_Reg(UARTCR_RW);
    *(UINT16*)Sys_Serial_Reg(UARTCR_RW) = (CrVal & 0xFFFE);
    Sys_Serial_Wait();

    // Configure line control and baud rate.
    *(UINT8*)Sys_Serial_Reg(UARTLCR_H_RW) = UARTLCR_H_8WORD_1STOP;

    // Rounded fractional and integer calculation
    // from https://krinkinmu.github.io/2020/11/29/PL011.html
    UINTN Div = 4 * ClockRate / BaudRate;
    UINT8 Fractional = Div & 0x3f;
    UINT16 Integer = (Div >> 6) & 0xffff;
    *(UINT8*)Sys_Serial_Reg(UARTFBRD_RW) = Fractional;
    *(UINT16*)Sys_Serial_Reg(UARTIBRD_RW) = Integer;

    // Enable UART transmission.
    CrVal = UARTCR_UARTEN | UARTCR_TXE;
    *(UINT16*)Sys_Serial_Reg(UARTCR_RW) = CrVal;
}


void SYSABI Sys_Serial_Write(IN const UINT8 *Buffer, IN const UINTN BufferLen) {
    Sys_Serial_Wait();
    for (UINTN i = 0; i < BufferLen; i++) {
        *(UINT8*)Sys_Serial_Reg(UARTDR_RW) = Buffer[i];
        Sys_Serial_Wait();
    }
}


#endif // defined( arm )