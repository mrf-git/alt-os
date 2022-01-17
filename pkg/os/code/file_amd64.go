package code

import (
	"alt-os/os/limits"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"unsafe"
)

var _ELF_PLT_ENTRY_SIZE = 16
var _ELF_GOT_ENTRY_SIZE = 8

var _ELF_DYN_BUFFER_SIZE = 1024

// readAmd64CodeFromElf reads amd64 code from the specified elf file.
func readAmd64CodeFromElf(r io.ReaderAt, elfFile *elf.File, exeCode *_ExecutableCode) error {
	if elfFile.ByteOrder != binary.LittleEndian {
		return fmt.Errorf("bad ELF byte order: %s", elfFile.ByteOrder.String())
	}
	if elfFile.Machine != elf.EM_X86_64 {
		return fmt.Errorf("bad ELF machine: %s", elfFile.Machine.String())
	}
	if elfFile.Class != elf.ELFCLASS64 {
		return fmt.Errorf("bad ELF class: %s", elfFile.Class.String())
	}

	curFileOffset := 0
	curFileBytes := make([]byte, _ELF_DYN_BUFFER_SIZE)

	// Read program headers and make sure the ELF is dynamically loadable
	// with the expected dynamic loader.
	isLoadable := false
	dynLoader := ""
	dynOffset := 0
	dynVaddr := 0
	dynSize := 0
	for _, progEntry := range elfFile.Progs {
		if progEntry.Type == elf.PT_LOAD {
			isLoadable = true
		} else if progEntry.Type == elf.PT_DYNAMIC {
			if dynOffset != 0 && dynSize != 0 && dynVaddr != 0 {
				return fmt.Errorf("multiple ELF dynamic sections")
			}
			dynOffset = int(progEntry.Off)
			dynSize = int(progEntry.Filesz)
			dynVaddr = int(progEntry.Vaddr)
		} else if progEntry.Type == elf.PT_INTERP && progEntry.Filesz > 1 {
			data := make([]byte, progEntry.Filesz-1) // -1 to exclude terminating null.
			if n, err := r.ReadAt(data, int64(progEntry.Off)); err != nil {
				return err
			} else if n != int(progEntry.Filesz-1) {
				return fmt.Errorf("failed to read ELF interp")
			} else {
				dynLoader = string(data)
			}
		}
	}
	if !isLoadable {
		return fmt.Errorf("ELF not loadable")
	}
	if dynLoader != _ELF_INTERP {
		return fmt.Errorf("bad ELF interp: expected '%s' but got '%s'", _ELF_INTERP, dynLoader)
	}
	if dynOffset == 0 || dynSize == 0 || dynVaddr == 0 {
		return fmt.Errorf("missing ELF dynamic section")
	}

	// Parse the dynamic section to get relocation information.
	symTabEntrySize := 0
	numPltRelocations := 0
	pltRelocTabSize := 0
	pltRelocTabVaddr := 0
	gotVaddr := 0
	dynEntrySize := 0
	dynNumEntries := 0
	for _, section := range elfFile.Sections {
		if section.Addr == uint64(dynVaddr) && section.Entsize != 0 {
			dynNumEntries = int(section.Size / section.Entsize)
			dynEntrySize = int(section.Entsize)
			break
		}
	}
	errUnsupported := func(dyn *elf.Dyn64) error {
		return fmt.Errorf("the ELF dynamic tag %d is unsupported", int(dyn.Tag))
	}
	curFileOffset = dynOffset
	for i := 0; i < dynNumEntries; i++ {
		if n, err := r.ReadAt(curFileBytes[:dynEntrySize], int64(curFileOffset)); err != nil {
			return err
		} else if n != dynEntrySize {
			return fmt.Errorf("failed to read ELF dynamic entry")
		}
		dyn := (*elf.Dyn64)(unsafe.Pointer(&curFileBytes[0]))
		curFileOffset += dynEntrySize
		if dyn.Tag == int64(elf.DT_NULL) {
			break // Terminating entry.
		}
		switch elf.DynTag(dyn.Tag) {
		default:
			if int(dyn.Tag) < int(elf.DT_LOOS) {
				return fmt.Errorf("unrecognized ELF dynamic tag: %d", int(dyn.Tag))
			}
		case elf.DT_FLAGS, elf.DT_HASH: // Ignored.
		case elf.DT_JMPREL:
			pltRelocTabVaddr = int(dyn.Val)
		case elf.DT_PLTRELSZ:
			pltRelocTabSize = int(dyn.Val)
			numPltRelocations = pltRelocTabSize / _ELF_RELA_SIZE
		case elf.DT_PLTREL:
			if elf.DynTag(dyn.Val) != elf.DT_RELA {
				return fmt.Errorf("ELF relocations missing addend")
			}
		case elf.DT_PLTGOT:
			gotVaddr = int(dyn.Val)
		case elf.DT_SYMTAB: // Handled later.
		case elf.DT_SYMENT:
			symTabEntrySize = int(dyn.Val)
		case elf.DT_STRTAB, elf.DT_STRSZ: // Already parsed.
		case elf.DT_PREINIT_ARRAY,
			elf.DT_PREINIT_ARRAYSZ,
			elf.DT_INIT_ARRAY,
			elf.DT_INIT_ARRAYSZ,
			elf.DT_FINI_ARRAY,
			elf.DT_FINI_ARRAYSZ:
			return errUnsupported(dyn)
		}
	}
	if symTabEntrySize == 0 {
		return fmt.Errorf("no ELF symbols")
	}
	if numPltRelocations != 0 && gotVaddr == 0 {
		return fmt.Errorf("missing GOT")
	}
	if pltRelocTabSize != 0 && pltRelocTabVaddr == 0 {
		return fmt.Errorf("missing PLT relocations")
	}

	// Calculate required memory sizes and get references to needed
	// sections from the elfFile sections slice.
	symRelocsIndex := 0
	pltRelocsIndex := 0
	ehRelocsIndex := 0
	var textSection *elf.Section
	var pltSection *elf.Section
	var gotSection *elf.Section
	var dataSection *elf.Section
	var roDataSection *elf.Section
	var bssSection *elf.Section
	var ehSection *elf.Section
	for sectionIndex, section := range elfFile.Sections {
		memorySize := uint(alignVal(int(section.FileSize), int(section.Addralign)))
		switch {
		default:
			return fmt.Errorf("unsupported ELF section '%s'", section.Name)
		case section.Name == "", section.Type == 0, section.Name == ".comment",
			section.Name == ".hash", strings.HasPrefix(section.Name, ".gnu"),
			section.Name == ".eh_frame_hdr":
			// Skip unused sections.
		case section.Name == ".text" && section.Type == elf.SHT_PROGBITS:
			textSection = section
			exeCode.CodeSize = memorySize
		case section.Name == ".plt" && section.Type == elf.SHT_PROGBITS:
			pltSection = section
			exeCode.PltSize = memorySize
			exeCode.numPltEntries = exeCode.PltSize / uint(_ELF_PLT_ENTRY_SIZE)
		case section.Name == ".got" && section.Type == elf.SHT_PROGBITS:
			gotSection = section
			exeCode.GotSize = memorySize
			exeCode.numGotEntries = exeCode.GotSize / uint(_ELF_GOT_ENTRY_SIZE)
		case section.Name == ".data" && section.Type == elf.SHT_PROGBITS:
			dataSection = section
			exeCode.DataSize = memorySize
		case section.Name == ".rodata" && section.Type == elf.SHT_PROGBITS:
			roDataSection = section
			exeCode.RoDataSize = memorySize
		case section.Name == ".eh_frame" && section.Type == elf.SHT_PROGBITS:
			ehSection = section
			exeCode.EhSize = memorySize
		case section.Name == ".bss" && section.Type == elf.SHT_NOBITS:
			bssSection = section
			exeCode.BssSize = memorySize
		case section.Name == ".rela.text" && section.Type == elf.SHT_RELA:
			symRelocsIndex = sectionIndex
		case section.Name == ".rela.plt" && section.Type == elf.SHT_RELA:
			pltRelocsIndex = sectionIndex
		case section.Name == ".rela.eh_frame" && section.Type == elf.SHT_RELA:
			ehRelocsIndex = sectionIndex
		case section.Name == ".interp", // Already parsed.
			section.Name == ".dynamic",
			section.Name == ".dynsym",
			section.Name == ".dynstr",
			section.Name == ".strtab",
			section.Name == ".symtab",
			section.Name == ".shstrtab":
		}
	}
	if textSection == nil {
		return fmt.Errorf("no .text section")
	}
	if exeCode.GetSize() > limits.MAX_EXECUTABLE_SIZE {
		return &limits.ErrLimitExceeded{
			LimitName:   "MAX_EXECUTABLE_SIZE",
			LimitValue:  limits.MAX_EXECUTABLE_SIZE,
			ActualValue: int(exeCode.GetSize()),
		}
	}
	if exeCode.numPltEntries != 0 &&
		exeCode.numPltEntries != exeCode.numGotEntries-2 {
		// First 2 GOT entries are reserved.
		return fmt.Errorf("bad GOT entries")
	}

	// Read section bytes into memory.
	errSection := func(sectionName string) error {
		return fmt.Errorf("failed to read ELF %s section", sectionName)
	}
	exeCode.codeBytes = make([]byte, exeCode.CodeSize)
	if n, err := r.ReadAt(exeCode.codeBytes[:textSection.FileSize],
		int64(textSection.Offset)); err != nil {
		return err
	} else if n != int(textSection.FileSize) {
		return errSection("text")
	} else {
		exeCode.codeAddrOrig = uint(textSection.Addr)
	}
	if exeCode.DataSize > 0 {
		exeCode.dataBytes = make([]byte, exeCode.DataSize)
		if n, err := r.ReadAt(exeCode.dataBytes[:dataSection.FileSize],
			int64(dataSection.Offset)); err != nil {
			return err
		} else if n != int(dataSection.FileSize) {
			return errSection("data")
		} else {
			exeCode.dataAddrOrig = uint(dataSection.Addr)
		}
	}
	if exeCode.RoDataSize > 0 {
		exeCode.roDataBytes = make([]byte, exeCode.RoDataSize)
		if n, err := r.ReadAt(exeCode.roDataBytes[:roDataSection.FileSize],
			int64(roDataSection.Offset)); err != nil {
			return err
		} else if n != int(roDataSection.FileSize) {
			return errSection("rodata")
		} else {
			exeCode.roDataAddrOrig = uint(roDataSection.Addr)
		}
	}
	if exeCode.PltSize > 0 {
		exeCode.pltBytes = make([]byte, exeCode.PltSize)
		if n, err := r.ReadAt(exeCode.pltBytes[:pltSection.FileSize],
			int64(pltSection.Offset)); err != nil {
			return err
		} else if n != int(pltSection.FileSize) {
			return errSection("plt")
		} else {
			exeCode.pltAddrOrig = uint(pltSection.Addr)
		}
	}
	if exeCode.GotSize > 0 {
		exeCode.gotBytes = make([]byte, exeCode.GotSize)
		if n, err := r.ReadAt(exeCode.gotBytes[:gotSection.FileSize],
			int64(gotSection.Offset)); err != nil {
			return err
		} else if n != int(gotSection.FileSize) {
			return errSection("got")
		} else {
			exeCode.gotAddrOrig = uint(gotSection.Addr)
		}
	}
	if exeCode.EhSize > 0 {
		exeCode.ehBytes = make([]byte, exeCode.EhSize)
		if n, err := r.ReadAt(exeCode.ehBytes[:ehSection.FileSize],
			int64(ehSection.Offset)); err != nil {
			return err
		} else if n != int(ehSection.FileSize) {
			return errSection("eh_frame")
		} else {
			exeCode.ehAddrOrig = uint(ehSection.Addr)
		}
	}
	if bssSection != nil {
		exeCode.bssAddrOrig = uint(bssSection.Addr)
	}

	// TODO parse the relocations and save external references
	_ = symRelocsIndex
	_ = pltRelocsIndex
	_ = ehRelocsIndex

	return nil
}
