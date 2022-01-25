#include "Common.h"
#include "FrameBuffer.h"


void SYSABI Sys_FrameBuffer_Clear(IN SYS_FRAMEBUFFER *FrameBuffer, IN const UINT32 Value) {

    void *CurRow = FrameBuffer->Base;
    for (UINTN y=0; y < FrameBuffer->Height; y++) {
        for (UINTN x=0; x < FrameBuffer->Width; x++) {
            UINT32* p = (UINT32*) (CurRow + (x << 2));
            *p = Value;
        }
        CurRow += FrameBuffer->RowSize;
    }
}

