/*
    Linker script for EFI binary objects. The output ELF object file is processed by
    EDK2 build tools and ultimately becomes a machine-bootable image, so its
    structure is different from OS runnables.
*/

SECTIONS
{
    . = PECOFF_HEADER_SIZE;

    .text : ALIGN(CONSTANT(MAXPAGESIZE))
    {
        *(.text .text.* .stub)
        *(.rodata .rodata.*)
        *(.got .got.*)
        *:AutoGen.obj(.data.g*Guid)
    }

    .data ALIGN(ALIGNOF(.text)) : ALIGN(CONSTANT(MAXPAGESIZE))
    {
        *(.data .data.*)
        *(.bss .bss.*)
    }

    .eh_frame ALIGN(CONSTANT(MAXPAGESIZE)) :
    {
        KEEP (*(.eh_frame))
    }

    .rela (INFO) :
    {
        *(.rela .rela.*)
    }

    .hii : ALIGN(CONSTANT(MAXPAGESIZE))
    {
        KEEP (*(.hii))
    }

    /DISCARD/ :
    {
        *(.note.GNU-stack)
        *(.gnu_debuglink)
        *(.gnu.*)
        *(.interp)
        *(.dynsym)
        *(.dynstr)
        *(.hash)
        *(.comment)
        *(COMMON)
    }
}
