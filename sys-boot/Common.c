#include "Common.h"


UINTN SYSABI Sys_Common_AsciiStrLen(IN const CHAR8 *Str) {
    UINTN Len;
    for (Len=0; *Str && Len < SYS_MAX_STR_LEN; Str++, Len++) {}
    return Len;
}


INTN SYSABI Sys_Common_AsciiStrCmp(IN const CHAR8 *Str1, IN const CHAR8 *Str2) {
    UINTN i = 0;
    while (*Str1 != 0 && *Str2 != 0 && *Str1 == *Str2 && i < SYS_MAX_STR_LEN-1) {
        Str1++;
        Str2++;
        i++;
    }
    return *Str1 - *Str2;
}

UINTN SYSABI Sys_Common_AsciiStrCopy(IN const CHAR8 *Str, OUT CHAR8 *Out) {
    CHAR8 *Src = (CHAR8*) Str;
    CHAR8 *Dest = (CHAR8*) Out;
    UINTN Size = 0;
    for (; *Src && Size < SYS_MAX_STR_LEN; Size++) {
        *Dest++ = *Src++;
    }
    Out[Size++] = 0;
    return Size;
}

void SYSABI Sys_Common_IntToAscii(IN const INTN Value, IN const BOOLEAN IsUnsigned, IN const UINTN Base, OUT CHAR8 *Buffer) {
    UINTN Ind = 0;
    UINTN CurValue;
    if (!IsUnsigned && Value < 0) {
        Buffer[Ind++] = '-';
        CurValue = -1 * Value;
    } else {
        CurValue = (UINTN) Value;
    }
    if (Base == 16) {
        Buffer[Ind++] = '0';
        Buffer[Ind++] = 'x';
    }
    if (!Value) {
        Buffer[Ind++] = '0';
    } else {
        // Fill the buffer with each digit.
        UINTN FirstDigitInd = Ind;
        UINTN NumDigits = 0;
        for (; CurValue && Ind < SYS_MAX_INT_ASCII_LEN - 1;) {
            UINTN Digit;
            if (Base == 16) {
                for (UINTN i=0; i < 2; i++) {
                    Digit = CurValue & 0x0F;
                    CurValue >>= 4;
                    if (Digit > 9) {
                        Digit = 'a' + Digit - 10;
                    } else {
                        Digit += '0';
                    }
                    Buffer[Ind++] = (CHAR8) Digit;
                    NumDigits++;
                }
            } else if (Base == 10) {
                Digit = CurValue % Base;
                CurValue /= Base;
                Digit += '0';
                Buffer[Ind++] = (CHAR8) Digit;
                NumDigits++;
            } else {
                // Base not supported.
                CurValue /= Base;
                Buffer[Ind++] = '-';
                NumDigits++;
            }
        }

        // Reverse the order to make it most significant digit first.
        for (UINTN i=0; i < (NumDigits << 1); i++) {
            UINTN StartInd = FirstDigitInd + i;
            UINTN EndInd = Ind - i - 1;
            if (StartInd >= EndInd) {
                break;
            }
            CHAR8 c = Buffer[EndInd];
            Buffer[EndInd] = Buffer[StartInd];
            Buffer[StartInd] = c;
        }
    }

    Buffer[Ind] = 0;
}


void SYSABI Sys_Common_MemCopy(IN const void *Mem, IN const UINTN Len, OUT const void *Buffer) {
    UINT8 *Src = (UINT8*) Mem;
    UINT8 *Dest = (UINT8*) Buffer;
    for (UINTN i=0; i < Len; i++) {
        *Dest++ = *Src++;
    }
}


UINTN SYSABI Sys_Common_StringHash(IN const CHAR8 *String, IN const UINTN Seed, IN const UINTN Mod) {
    UINTN Hash = Seed;
    for (UINTN i=0; i < SYS_MAX_STR_LEN; i++) {
        CHAR8 Char = String[i];
        if (!Char) {
            break;
        }
        Hash ^= Char;
        Hash *= Mod;
    }
    return Hash % Mod;
}

