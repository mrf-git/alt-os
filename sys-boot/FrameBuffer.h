#ifndef SYS_BOOT_FRAME_BUFFER_H
#define SYS_BOOT_FRAME_BUFFER_H

#include "Common.h"


//
// Enum for pixel format.
//
typedef enum {
    PixelFormatRGB,     // P[0] is red, P[1] is green, P[2] is blue.
    PixelFormatBGR,     // P[0] is blue, P[1] is green, P[2] is red.
} SYS_PIXEL_FORMAT;


//
// Represents a pixel-addressable framebuffer.
//
typedef struct {
    UINTN Status;                   // The EFI_STATUS of the graphical framebuffer setup.
    void *Base;                     // Pointer to the framebuffer base.
    UINTN Size;                     // Size in bytes of the framebuffer.
    SYS_PIXEL_FORMAT PixelFormat;   // Sys-defined pixel format of the framebuffer.
    UINTN RowSize;                  // Size in bytes of a single row in the framebuffer.
    UINTN Width;                    // Number of pixels in the width of the framebuffer.
    UINTN Height;                   // Number of pixels in the height of the framebuffer.
} SYS_FRAMEBUFFER;


//
// Clears the framebuffer to the specified value, which must be given in the supported pixel format.
//
SYSEXPORT void SYSABI Sys_FrameBuffer_Clear(IN SYS_FRAMEBUFFER *FrameBuffer, IN const UINT32 Value);


#endif // SYS_BOOT_FRAME_BUFFER_H
