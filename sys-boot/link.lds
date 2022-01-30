/*
    Linker script for EFI binary objects. The output ELF object file is processed by
    EDK2 build tools and ultimately becomes a machine-bootable image, so its
    structure is different from OS runnables.
*/

SECTIONS
{
    . = PECOFF_HEADER_SIZE;

    .text ALIGN(0x1000) :
    {
        *(.text .text.* .stub)
        *(.rodata .rodata.*)
        *(.got .got.*)
        *:AutoGen.obj(.data.g*Guid)
    }

    .data ALIGN(0x1000) :
    {
        *(.data .data.*)
        *(.bss .bss.*)
    }

    .eh_frame ALIGN(0x1000) :
    {
        KEEP (*(.eh_frame))
    }

    .rela (INFO) :
    {
        *(.rela .rela.*)
    }

    .hii ALIGN(0x1000) :
    {
        KEEP (*(.hii))
    }

    .build-id (INFO) : {
        *(.note.gnu.build-id)
    }

    /DISCARD/ :
    {
        *(.note.GNU-stack)
        *(.gnu_debuglink)
        *(.interp)
        *(.dynsym)
        *(.dynstr)
        *(.hash .gnu.hash)
        *(.comment)
        *(COMMON)
    }
}
