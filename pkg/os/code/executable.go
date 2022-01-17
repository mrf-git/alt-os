package code

// ExecutableCode represents binary executable code that can be run by the OS.
type ExecutableCode interface {
	// GetSize returns the total number of memory bytes in the executable code.
	GetSize() uint
	// GetSharedSize returns the number of bytes that can be shared among instances.
	GetSharedSize() uint
	// GetUnsharedSize returns the number of bytes that must be duplicated for instances.
	GetUnsharedSize() uint
	// GetSectionInfo returns memory information about sections in the executable.
	GetSectionInfo() *SectionMemoryInfo
	// LoadShared relocates shared sections into the specified memory
	// addresses by modifying the current in-memory bytes.
	LoadShared(addresses *SharedAddresses) error
	// LoadInstance relocates instance sections into the specified memory
	// addresses by first copying instance bytes before modifying them.
	// Returns the bytes for all sections, including shared.
	LoadInstance(addresses *InstanceAddresses) (*LoadedInstance, error)
}

// SectionMemoryInfo holds memory information about sections needed to
// execute an ExecutableCode.
type SectionMemoryInfo struct {
	CodeSize        uint
	CodeAlignment   uint
	DataSize        uint
	DataAlignment   uint
	RoDataSize      uint
	RoDataAlignment uint
	BssSize         uint
	BssAlignment    uint
	PltSize         uint
	PltAlignment    uint
	GotSize         uint
	GotAlignment    uint
	EhSize          uint
	EhAlignment     uint
}

// SharedAddresses holds memory addresses for each shared
// memory section to be loaded into.
type SharedAddresses struct {
	CodeAddr   uint
	RoDataAddr uint
}

// InstanceAddresses holds memory addresses for each memory
// section of an executable instance to be loaded into.
type InstanceAddresses struct {
	DataAddr uint
	BssAddr  uint
	PltAddr  uint
	GotAddr  uint
	EhAddr   uint
}

// LoadedInstance holds contents for sections that make up an executable
// instance, after being fully loaded to memory addresses.
type LoadedInstance struct {
	Code   []byte
	Data   []byte
	RoData []byte
	Bss    []byte
	Plt    []byte
	Got    []byte
	Eh     []byte
}

// Implementation of ExecutableCode.
type _ExecutableCode struct {
	SectionMemoryInfo
	numPltEntries  uint
	numGotEntries  uint
	codeBytes      []byte
	dataBytes      []byte
	roDataBytes    []byte
	pltBytes       []byte
	gotBytes       []byte
	ehBytes        []byte
	codeAddrOrig   uint
	dataAddrOrig   uint
	roDataAddrOrig uint
	bssAddrOrig    uint
	pltAddrOrig    uint
	gotAddrOrig    uint
	ehAddrOrig     uint
}

func (exeCode *_ExecutableCode) GetSharedSize() uint {
	return exeCode.CodeSize + exeCode.RoDataSize
}

func (exeCode *_ExecutableCode) GetUnsharedSize() uint {
	return exeCode.DataSize + exeCode.BssSize + exeCode.PltSize +
		exeCode.GotSize + exeCode.EhSize
}

func (exeCode *_ExecutableCode) GetSize() uint {
	return exeCode.GetSharedSize() + exeCode.GetUnsharedSize()
}

func (exeCode *_ExecutableCode) GetSectionInfo() *SectionMemoryInfo {
	return &SectionMemoryInfo{
		CodeSize:        exeCode.CodeSize,
		CodeAlignment:   exeCode.CodeAlignment,
		DataSize:        exeCode.DataSize,
		DataAlignment:   exeCode.DataAlignment,
		RoDataSize:      exeCode.RoDataSize,
		RoDataAlignment: exeCode.RoDataAlignment,
		BssSize:         exeCode.BssSize,
		BssAlignment:    exeCode.BssAlignment,
		PltSize:         exeCode.PltSize,
		PltAlignment:    exeCode.PltAlignment,
		GotSize:         exeCode.GotSize,
		GotAlignment:    exeCode.GotAlignment,
		EhSize:          exeCode.EhSize,
		EhAlignment:     exeCode.EhAlignment,
	}
}

func (exeCode *_ExecutableCode) LoadShared(addresses *SharedAddresses) error {
	// TODO
	return nil
}

func (exeCode *_ExecutableCode) LoadInstance(addresses *InstanceAddresses) (*LoadedInstance, error) {

	instance := &LoadedInstance{}
	// TODO
	return instance, nil
}
