package main

import (
	"alt-os/exe"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// edk2toolparams creates and returns edk2 tool parameter maps for the specified context.
func edk2toolparams(ctxt *OsbuildContext) (toolDefs map[string]string, buildParams map[string]string, toolParams map[string]string) {

	// Definitions used by tool params.
	toolDefs = map[string]string{
		"SYS_DLINK_FLAGS_COMMON":  "-nodefaultlibs -nostdlib -Wl,-q,--gc-sections -z max-page-size=0x40 -fuse-ld=lld",
		"SYS_DLINK2_FLAGS_COMMON": "-Wl,--script=" + path.Join(ctxt.SrcRootDir, "sys-boot", "link.lds"),
		"SYS_DLINK_FLAGS_APP":     "-Wl,--entry,$(IMAGE_ENTRY_POINT) -u $(IMAGE_ENTRY_POINT) -Wl,-Map,$(DEST_DIR_DEBUG)/$(BASE_NAME).map",
		"SYS_DLINK_FLAGS":         "DEF(SYS_DLINK_FLAGS_COMMON) DEF(SYS_DLINK_FLAGS_APP) -Wl,--whole-archive",
		"SYS_DLINK2_FLAGS":        "-Wl,--defsym=PECOFF_HEADER_SIZE=0x228 DEF(SYS_DLINK2_FLAGS_COMMON)",
		"SYS_ASLDLINK_FLAGS":      "-Wl,--defsym=PECOFF_HEADER_SIZE=0 -Wl,--entry,ReferenceAcpiTable -u ReferenceAcpiTable",
	}

	// Parameters affecting all tools.
	buildParams = map[string]string{
		"FAMILY":          "GCC",
		"BUILDRULEFAMILY": "CLANGGCC",
		"BUILDRULEORDER":  "nasm asm Asm ASM S s nasmb asm16",
	}

	// Parameters for individual tools.
	ccFlags := strings.Join(ctxt.FlagsCCBoot, " ")
	ccFlagsArch := strings.Join(ctxt.FlagsCCBootArch, " ")
	rcFlags := strings.Join(ctxt.FlagsRCBoot, " ")
	nasmFlags := strings.Join(ctxt.FlagsNasmBoot, " ")
	linkFlagsArch := strings.Join(ctxt.FlagsLinkBootArch, " ")
	toolParams = map[string]string{
		"ASL_PATH":               "iasl",
		"ASLCC_PATH":             path.Join(ctxt.LlvmDir, "clang"),
		"ASLDLINK_PATH":          path.Join(ctxt.LlvmDir, "clang"),
		"ASLPP_PATH":             path.Join(ctxt.LlvmDir, "clang"),
		"ASM_PATH":               path.Join(ctxt.LlvmDir, "clang"),
		"BROTLI_PATH":            "BrotliCompress",
		"CC_PATH":                path.Join(ctxt.LlvmDir, "clang"),
		"CRC32_PATH":             "GenCrc32",
		"DLINK_PATH":             path.Join(ctxt.LlvmDir, "clang"),
		"GENFW_PATH":             "GenFw",
		"LZMA_PATH":              "LzmaCompress",
		"LZMAF86_PATH":           "LzmaF86Compress",
		"MAKE_PATH":              "make",
		"NASM_PATH":              "nasm",
		"OBJCOPY_PATH":           "echo",
		"OPTROM_PATH":            "EfiRom",
		"PKCS7SIGN_PATH":         "Pkcs7Sign",
		"PP_PATH":                path.Join(ctxt.LlvmDir, "clang"),
		"RC_PATH":                "llvm-rc",
		"RSA2048SHA256SIGN_PATH": "Rsa2048Sha256Sign",
		"SLINK_PATH":             "llvm-ar",
		"TIANO_PATH":             "TianoCompress",
		"VFR_PATH":               "VfrCompile",
		"VFRPP_PATH":             path.Join(ctxt.LlvmDir, "clang"),
		"VPDTOOL_PATH":           "BPDG",
		"APP_FLAGS":              "",
		"ASL_FLAGS":              "",
		"ASL_OUTFLAGS":           "-p",
		"ASLCC_FLAGS":            "-fno-lto " + ccFlagsArch,
		"ASLDLINK_FLAGS":         "DEF(SYS_DLINK_FLAGS_COMMON) DEF(SYS_DLINK2_FLAGS_COMMON) DEF(SYS_ASLDLINK_FLAGS) " + linkFlagsArch,
		"ASLPP_FLAGS":            "-E -include AutoGen.h " + ccFlagsArch,
		"ASM_FLAGS":              "-c -x assembler -imacros AutoGen.h " + ccFlagsArch,
		"CC_FLAGS":               ccFlags + " -DSTRING_ARRAY_NAME=$(BASE_NAME)Strings",
		"DLINK_FLAGS":            "DEF(SYS_DLINK_FLAGS) -fno-lto -Wl,-O3 -Wl,-pie -Wl,--apply-dynamic-relocs " + linkFlagsArch,
		"DLINK2_FLAGS":           "DEF(SYS_DLINK2_FLAGS) -O3 -fuse-ld=lld",
		"GENFW_FLAGS":            "",
		"NASM_FLAGS":             nasmFlags,
		"NASMB_FLAGS":            "-f bin",
		"OBJCOPY_FLAGS":          "",
		"OPTROM_FLAGS":           "-e",
		"PP_FLAGS":               "-E -x assembler-with-cpp -include AutoGen.h -DOPENSBI_EXTERNAL_SBI_TYPES=OpensbiTypes.h " + ccFlagsArch,
		"RC_FLAGS":               "-I binary " + rcFlags + " --rename-section .data=.hii",
		"VFR_FLAGS":              "-l -n",
		"VFRPP_FLAGS":            "-E -P -DVFRCOMPILE --include $(MODULE_NAME)StrDefs.h " + ccFlagsArch,
	}

	// Expand references among toolDefs before returning.
	for expanding, replaced := true, false; expanding; expanding = replaced {
		replaced = false
		for key1, val1 := range toolDefs {
			for key2, val2 := range toolDefs {
				refStr := fmt.Sprintf("DEF(%s)", key2)
				refInd := strings.Index(val1, refStr)
				if refInd >= 0 {
					newVal1 := val1[:refInd] + val2 + val1[refInd+len(refStr):]
					toolDefs[key1] = newVal1
					replaced = true
				}
			}
		}
	}
	return
}

// edk2libparams creates and returns edk2 library parameter maps for the specified context.
func edk2libparams(ctxt *OsbuildContext) (libParams map[string]string) {

	libParams = map[string]string{
		"BaseLib":                     "MdePkg/Library/BaseLib/BaseLib.inf",
		"BaseMemoryLib":               "MdePkg/Library/BaseMemoryLib/BaseMemoryLib.inf",
		"CacheMaintenanceLib":         "MdePkg/Library/BaseCacheMaintenanceLib/BaseCacheMaintenanceLib.inf",
		"DebugLib":                    "MdePkg/Library/BaseDebugLibNull/BaseDebugLibNull.inf",
		"DebugPrintErrorLevelLib":     "MdePkg/Library/BaseDebugPrintErrorLevelLib/BaseDebugPrintErrorLevelLib.inf",
		"DevicePathLib":               "MdePkg/Library/UefiDevicePathLib/UefiDevicePathLib.inf",
		"DxeServicesTableLib":         "MdePkg/Library/DxeServicesTableLib/DxeServicesTableLib.inf",
		"DxeCoreEntryPoint":           "MdePkg/Library/DxeCoreEntryPoint/DxeCoreEntryPoint.inf",
		"DxeServicesLib":              "MdePkg/Library/DxeServicesLib/DxeServicesLib.inf",
		"FileHandleLib":               "MdePkg/Library/UefiFileHandleLib/UefiFileHandleLib.inf",
		"HobLib":                      "MdePkg/Library/DxeHobLib/DxeHobLib.inf",
		"IoLib":                       "MdePkg/Library/BaseIoLibIntrinsic/BaseIoLibIntrinsic.inf",
		"MemoryAllocationLib":         "MdePkg/Library/UefiMemoryAllocationLib/UefiMemoryAllocationLib.inf",
		"MmUnblockMemoryLib":          "MdePkg/Library/MmUnblockMemoryLib/MmUnblockMemoryLibNull.inf",
		"PcdLib":                      "MdePkg/Library/BasePcdLibNull/BasePcdLibNull.inf",
		"PciCf8Lib":                   "MdePkg/Library/BasePciCf8Lib/BasePciCf8Lib.inf",
		"PciLib":                      "MdePkg/Library/BasePciLibCf8/BasePciLibCf8.inf",
		"PciSegmentLib":               "MdePkg/Library/BasePciSegmentLibPci/BasePciSegmentLibPci.inf",
		"PeCoffLib":                   "MdePkg/Library/BasePeCoffLib/BasePeCoffLib.inf",
		"PeCoffExtraActionLib":        "MdePkg/Library/BasePeCoffExtraActionLibNull/BasePeCoffExtraActionLibNull.inf",
		"PeCoffGetEntryPointLib":      "MdePkg/Library/BasePeCoffGetEntryPointLib/BasePeCoffGetEntryPointLib.inf",
		"PeiCoreEntryPoint":           "MdePkg/Library/PeiCoreEntryPoint/PeiCoreEntryPoint.inf",
		"PeimEntryPoint":              "MdePkg/Library/PeimEntryPoint/PeimEntryPoint.inf",
		"PeiServicesLib":              "MdePkg/Library/PeiServicesLib/PeiServicesLib.inf",
		"PeiServicesTablePointerLib":  "MdePkg/Library/PeiServicesTablePointerLib/PeiServicesTablePointerLib.inf",
		"PerformanceLib":              "MdePkg/Library/BasePerformanceLibNull/BasePerformanceLibNull.inf",
		"PrintLib":                    "MdePkg/Library/BasePrintLib/BasePrintLib.inf",
		"RegisterFilterLib":           "MdePkg/Library/RegisterFilterLibNull/RegisterFilterLibNull.inf",
		"ReportStatusCodeLib":         "MdePkg/Library/BaseReportStatusCodeLibNull/BaseReportStatusCodeLibNull.inf",
		"SafeIntLib":                  "MdePkg/Library/BaseSafeIntLib/BaseSafeIntLib.inf",
		"SerialPortLib":               "MdePkg/Library/BaseSerialPortLibNull/BaseSerialPortLibNull.inf",
		"SmbusLib":                    "MdePkg/Library/DxeSmbusLib/DxeSmbusLib.inf",
		"SynchronizationLib":          "MdePkg/Library/BaseSynchronizationLib/BaseSynchronizationLib.inf",
		"TimerLib":                    "MdePkg/Library/BaseTimerLibNullTemplate/BaseTimerLibNullTemplate.inf",
		"UefiApplicationEntryPoint":   "MdePkg/Library/UefiApplicationEntryPoint/UefiApplicationEntryPoint.inf",
		"UefiBootServicesTableLib":    "MdePkg/Library/UefiBootServicesTableLib/UefiBootServicesTableLib.inf",
		"UefiDecompressLib":           "MdePkg/Library/BaseUefiDecompressLib/BaseUefiDecompressLib.inf",
		"UefiDriverEntryPoint":        "MdePkg/Library/UefiDriverEntryPoint/UefiDriverEntryPoint.inf",
		"UefiLib":                     "MdePkg/Library/UefiLib/UefiLib.inf",
		"UefiRuntimeLib":              "MdePkg/Library/UefiRuntimeLib/UefiRuntimeLib.inf",
		"UefiRuntimeServicesTableLib": "MdePkg/Library/UefiRuntimeServicesTableLib/UefiRuntimeServicesTableLib.inf",
		"UefiScsiLib":                 "MdePkg/Library/UefiScsiLib/UefiScsiLib.inf",
		"UefiUsbLib":                  "MdePkg/Library/UefiUsbLib/UefiUsbLib.inf",

		"AuthVariableLib":                      "MdeModulePkg/Library/AuthVariableLibNull/AuthVariableLibNull.inf",
		"BmpSupportLib":                        "MdeModulePkg/Library/BaseBmpSupportLib/BaseBmpSupportLib.inf",
		"CapsuleLib":                           "MdeModulePkg/Library/DxeCapsuleLibNull/DxeCapsuleLibNull.inf",
		"CpuExceptionHandlerLib":               "MdeModulePkg/Library/CpuExceptionHandlerLibNull/CpuExceptionHandlerLibNull.inf",
		"CustomizedDisplayLib":                 "MdeModulePkg/Library/CustomizedDisplayLib/CustomizedDisplayLib.inf",
		"DebugAgentLib":                        "MdeModulePkg/Library/DebugAgentLibNull/DebugAgentLibNull.inf",
		"DisplayUpdateProgressLib":             "MdeModulePkg/Library/DisplayUpdateProgressLibGraphics/DisplayUpdateProgressLibGraphics.inf",
		"FileExplorerLib":                      "MdeModulePkg/Library/FileExplorerLib/FileExplorerLib.inf",
		"FmpAuthenticationLib":                 "MdeModulePkg/Library/FmpAuthenticationLibNull/FmpAuthenticationLibNull.inf",
		"FrameBufferBltLib":                    "MdeModulePkg/Library/FrameBufferBltLib/FrameBufferBltLib.inf",
		"HiiLib":                               "MdeModulePkg/Library/UefiHiiLib/UefiHiiLib.inf",
		"NonDiscoverableDeviceRegistrationLib": "MdeModulePkg/Library/NonDiscoverableDeviceRegistrationLib/NonDiscoverableDeviceRegistrationLib.inf",
		"PciHostBridgeLib":                     "MdeModulePkg/Library/PciHostBridgeLibNull/PciHostBridgeLibNull.inf",
		"PlatformBootManagerLib":               "MdeModulePkg/Library/PlatformBootManagerLibNull/PlatformBootManagerLibNull.inf",
		"PlatformHookLib":                      "MdeModulePkg/Library/BasePlatformHookLibNull/BasePlatformHookLibNull.inf",
		"ResetSystemLib":                       "MdeModulePkg/Library/BaseResetSystemLibNull/BaseResetSystemLibNull.inf",
		"S3BootScriptLib":                      "MdeModulePkg/Library/PiDxeS3BootScriptLib/DxeS3BootScriptLib.inf",
		"SecurityManagementLib":                "MdeModulePkg/Library/DxeSecurityManagementLib/DxeSecurityManagementLib.inf",
		"SortLib":                              "MdeModulePkg/Library/BaseSortLib/BaseSortLib.inf",
		"TpmMeasurementLib":                    "MdeModulePkg/Library/TpmMeasurementLibNull/TpmMeasurementLibNull.inf",
		"UefiBootManagerLib":                   "MdeModulePkg/Library/UefiBootManagerLib/UefiBootManagerLib.inf",
		"UefiHiiServicesLib":                   "MdeModulePkg/Library/UefiHiiServicesLib/UefiHiiServicesLib.inf",
		"VarCheckLib":                          "MdeModulePkg/Library/VarCheckLib/VarCheckLib.inf",
		"VariablePolicyHelperLib":              "MdeModulePkg/Library/VariablePolicyHelperLib/VariablePolicyHelperLib.inf",
		"VariablePolicyLib":                    "MdeModulePkg/Library/VariablePolicyLib/VariablePolicyLib.inf",

		"Ucs2Utf8Lib": "RedfishPkg/Library/BaseUcs2Utf8Lib/BaseUcs2Utf8Lib.inf",
	}
	return
}

// edk2infparams creates and returns parameters for the edk2 build INF file.
func edk2infparams(guidFile *uuid.UUID, ctxt *OsbuildContext) (edk2DefinesInf map[string]string, edk2PackagesInf,
	edk2LibraryClassesInf, edk2GuidsInf, edk2ProtocolsInf, edk2SourcesInf []string) {

	edk2DefinesInf = map[string]string{
		"INF_VERSION":               "0x00010005",
		"BASE_NAME":                 "SysBoot",
		"MODULE_TYPE":               "UEFI_APPLICATION",
		"VERSION_STRING":            ctxt.BuildInfo.VersionStr,
		"UEFI_HII_RESOURCE_SECTION": "FALSE",
		"VALID_ARCHITECTURES":       ctxt.ArchEFI,
		"ENTRY_POINT":               "Sys_Boot_Entry",
		"FILE_GUID":                 guidFile.String(),
	}
	edk2PackagesInf = []string{
		"MdePkg/MdePkg.dec",
		"MdeModulePkg/MdeModulePkg.dec",
		"RedfishPkg/RedfishPkg.dec",
		"edk2Conf/SysModulePkg/SysModulePkg.dec",
	}
	edk2LibraryClassesInf = []string{
		"BaseLib",
		"BaseMemoryLib",
		"MemoryAllocationLib",
		"UefiLib",
		"UefiBootServicesTableLib",
		"UefiRuntimeServicesTableLib",
		"UefiApplicationEntryPoint",
		"Ucs2Utf8Lib",
	}
	edk2GuidsInf = []string{
		"gEfiAcpiTableGuid",
		"gEfiAcpi10TableGuid",
		"gEfiAcpi20TableGuid",
		"gEfiSmbiosTableGuid",
		"gEfiSmbios3TableGuid",
	}
	edk2ProtocolsInf = []string{
		"gEfiLoadedImageProtocolGuid",
		"gEfiSimpleFileSystemProtocolGuid",
		"gEfiGraphicsOutputProtocolGuid",
	}
	edk2SourcesInf = []string{}

	return
}

// edk2decparams creates and returns parameters for the edk2 build DEC file.
func edk2decparams(guidPackage *uuid.UUID, guidToken *uuid.UUID, ctxt *OsbuildContext) (edk2DefinesDec,
	edk2GuidsDec map[string]string, edk2IncludesDec []string) {

	edk2DefinesDec = map[string]string{
		"DEC_SPECIFICATION": "0x00010005",
		"PACKAGE_NAME":      "SysModulePkg",
		"PACKAGE_GUID":      guidPackage.String(),
		"PACKAGE_VERSION":   ctxt.BuildInfo.VersionStr,
	}

	edk2IncludesDec = []string{
		// Relative to edk2 package path sys-boot
	}

	tokenGuidByteStr := fmt.Sprintf("{ 0x%08x, 0x%04x, 0x%04x, { 0x%02x, 0x%02x, 0x%02x, 0x%02x, 0x%02x, 0x%02x, 0x%02x, 0x%02x } }",
		guidToken[0:4], guidToken[4:6], guidToken[6:8], guidToken[8], guidToken[9], guidToken[10], guidToken[11],
		guidToken[12], guidToken[13], guidToken[14], guidToken[15])

	edk2GuidsDec = map[string]string{
		"gEfiSysModulePkgTokenSpaceGuid": tokenGuidByteStr,
	}

	return
}

// edk2dscparams creates and returns parameters for the edk2 build DSC file.
func edk2dscparams(guidPlatform *uuid.UUID, ctxt *OsbuildContext) (edk2DefinesDsc map[string]string,
	edk2LibraryClassesDsc, edk2ComponentsDsc []string) {

	edk2DefinesDsc = map[string]string{
		"PLATFORM_NAME":           "SysModule",
		"PLATFORM_GUID":           guidPlatform.String(),
		"DSC_SPECIFICATION":       "0x00010005",
		"PLATFORM_VERSION":        ctxt.BuildInfo.VersionStr,
		"OUTPUT_DIRECTORY":        path.Join("Build", "SysModuleOutput"),
		"SUPPORTED_ARCHITECTURES": ctxt.ArchEFI,
		"BUILD_TARGETS":           "RELEASE",
		"SKUID_IDENTIFIER":        "DEFAULT",
	}
	edk2LibraryClassesDsc = []string{}
	for key, val := range edk2libparams(ctxt) {
		edk2LibraryClassesDsc = append(edk2LibraryClassesDsc, fmt.Sprintf("%s|%s", key, val))
	}
	edk2ComponentsDsc = []string{
		path.Join("edk2Conf", "SysModulePkg", "Application", "Sys.inf"),
	}

	return
}

// depEdk2 initializes and caches the edk2 dependency files.
func depEdk2(ctxt *OsbuildContext) (changed bool) {
	ctxt.Edk2WorkspacePath = path.Join(ctxt.CacheDir, "src", "sys-boot", "edk2Workspace")
	ctxt.Edk2ConfPath = path.Join(ctxt.CacheDir, "src", "sys-boot", "edk2Conf")
	ctxt.Edk2VarsPath = path.Join(ctxt.CacheDir, "dep", "edk2WorkspaceVars.sh")

	edk2Checkout := func() (hash string, changed bool) {
		edk2Path := path.Join(ctxt.CacheDir, "dep", "edk2")
		edk2Hash, edk2Hit := ctxt.CacheManifest["edk2Checkout"]
		if edk2Hit {
			// Hash is already defined. Check if the repo is on that hash and if it is a clean version.
			if hash, err := githash(edk2Path, ctxt); err != nil || hash != edk2Hash {
				edk2Hit = false
			} else if !isgitclean(edk2Path, ctxt) {
				edk2Hit = false
			}
		}
		if edk2Hit {
			ctxt.Logger.Info("Using cached dependency edk2")
		} else {
			os.RemoveAll(edk2Path)
			ctxt.Logger.Info("Checking out dependency edk2")
			args := []string{"clone", "--depth=1", "-b", ctxt.Conf.Deps.Edk2.Tag, "--single-branch",
				ctxt.Conf.Deps.Edk2.GitUrl, edk2Path}
			if stdOut, stdErr, err := exe.Doexec("", "git", args...); err != nil {
				exe.Fatal("cloning edk2", exe.ErrOutput(stdOut, stdErr, err), ctxt.ExeContext)
			}
			if hash, err := githash(edk2Path, ctxt); err != nil {
				exe.Fatal("getting edk2 hash", err, ctxt.ExeContext)
			} else {
				ctxt.Logger.Info("Initializing edk2")
				if stdOut, stdErr, err := exe.Doexec(edk2Path, "git", "submodule", "update", "--init", "--recursive"); err != nil {
					exe.Fatal("initializing edk2", exe.ErrOutput(stdOut, stdErr, err), ctxt.ExeContext)
				}
				ctxt.CacheManifest["edk2Checkout"] = hash
				return hash, true
			}
		}
		return edk2Hash, false
	}

	edk2Build := func(hash string, depChanged bool) (changed bool) {
		edk2Path := path.Join(ctxt.CacheDir, "dep", "edk2")
		edk2Hash, hit := ctxt.CacheManifest["edk2Build"]
		if !depChanged && hit && edk2Hash == hash {
			ctxt.Logger.Info("Using cached build of edk2")
			return false
		}
		ctxt.Logger.Info("Building edk2")
		switch runtime.GOOS + "/" + runtime.GOARCH {
		default:
			exe.Fatal("building edk2", errors.New("unknown host"), ctxt.ExeContext)
		case "linux/amd64":
			writeconfvars(ctxt, "x86_64-pc-linux")
		}
		writetoolsdef(ctxt)

		cmd := fmt.Sprintf("source %s ; source edksetup.sh ; make -C BaseTools", ctxt.Edk2VarsPath)
		if stdOut, stdErr, err := exe.Doexec(edk2Path, "bash", "-c", cmd); err != nil {
			exe.Fatal("building edk2", exe.ErrOutput(stdOut, stdErr, err), ctxt.ExeContext)
		}
		ctxt.CacheManifest["edk2Build"] = hash

		return true
	}

	edk2Hash, changed := edk2Checkout()
	if edk2Build(edk2Hash, changed) || changed {
		return true
	}
	return false
}

// writeconfvars writes the edk2 configuration vars for the specified target triple.
func writeconfvars(ctxt *OsbuildContext, targetTriple string) {
	os.MkdirAll(ctxt.Edk2WorkspacePath, 0755)
	os.MkdirAll(ctxt.Edk2ConfPath, 0755)

	if f, err := os.Create(ctxt.Edk2VarsPath); err != nil {
		exe.Fatal("creating "+ctxt.Edk2VarsPath, err, ctxt.ExeContext)
	} else {
		edk2Path := path.Join(ctxt.CacheDir, "dep", "edk2")
		packagePath := path.Join(ctxt.SrcRootDir, "sys-boot")
		cachePath := path.Join(ctxt.CacheDir, "src", "sys-boot")
		f.WriteString("#!/usr/bin/env bash\n")
		f.WriteString(fmt.Sprintf("export WORKSPACE=\"%s\"\n", ctxt.Edk2WorkspacePath))
		f.WriteString(fmt.Sprintf("export PACKAGES_PATH=\"%s:%s:%s\"\n", edk2Path, packagePath, cachePath))
		f.WriteString(fmt.Sprintf("export CONF_PATH=\"%s\"\n", ctxt.Edk2ConfPath))
		f.WriteString(fmt.Sprintf("export PATH=\"%s\"\n", os.Getenv("PATH")))
		f.WriteString(fmt.Sprintf("export BUILD_CC=\"%s -Wno-unknown-warning-option --target=%s\"\n", os.Getenv("CC"), targetTriple))
		f.WriteString(fmt.Sprintf("export BUILD_CXX=\"%s --target=%s\"\n", os.Getenv("CXX"), targetTriple))
		f.WriteString("export TARGET=RELEASE\n")
		f.WriteString(fmt.Sprintf("export TARGET_ARCH=%s\n", ctxt.ArchEFI))
		f.Close()
	}
}

// bootbuild generates the edk2 workspace files and creates the target bootable EFI image from the built OS and resources.
func bootbuild(ctxt *OsbuildContext) {

	// Scan the source directory to locate boot source files.
	ctxt.Logger.Info("Scanning boot sources")
	srcCachePath := path.Join(ctxt.CacheDir, "src")
	chSrcPaths := make(chan string, _MAX_NUM_OBJECTS)
	wgScan := &sync.WaitGroup{}
	wgScan.Add(1)
	go srcscan(path.Join(ctxt.SrcRootDir, "sys-boot"), ".c", "", chSrcPaths, wgScan, ctxt)
	wgScan.Wait()
	close(chSrcPaths)
	srcfiles := []string{}
	for {
		srcname, ok := <-chSrcPaths
		if srcname == "" || !ok {
			break
		}
		srcfiles = append(srcfiles, srcname)
	}

	// Scan the boot source files to list their dependencies.
	chDepPaths := make(chan string, _MAX_NUM_OBJECTS)
	wgDeps := &sync.WaitGroup{}
	for _, srcname := range srcfiles {
		wgDeps.Add(1)
		go func(srcname string) {
			args := []string{"-xc"}
			args = append(args, ctxt.FlagsCCRes...)
			args = append(args, "-frewrite-includes", "-M", "-E")
			args = append(args, srcname)
			if stdOut, stdErr, err := exe.Doexec("", path.Join(ctxt.LlvmDir, "clang"), args...); err != nil {
				exe.Fatal("scanning boot source dependencies", exe.ErrOutput(stdOut, stdErr, err), ctxt.ExeContext)
			} else {
				start := strings.Index(stdOut, ".o: ")
				if start >= 0 {
					stdOut = string(stdOut[start+4:])
				}
				trim := func(strs []string) []string {
					out := make([]string, 0)
					for _, in := range strs {
						s := strings.TrimSpace(in)
						if strings.HasPrefix(s, ctxt.SrcRootDir) || strings.HasPrefix(s, srcCachePath) {
							out = append(out, s)
						}
					}
					return out
				}
				for _, path := range trim(strings.Split(stdOut, "\\\n")) {
					chDepPaths <- path
				}
			}
			wgDeps.Done()
		}(srcname)
	}
	wgDeps.Wait()
	close(chDepPaths)

	// Read, deduplicate, and sort dependencies. Then calculate a hash of dependency names and get the mtime for each.
	changed := false
	updates := map[string]string{}
	depmap := make(map[string]interface{})
	for _, srcname := range srcfiles {
		depmap[srcname] = nil
	}
	for {
		depname, ok := <-chDepPaths
		if depname == "" || !ok {
			break
		}
		depmap[depname] = nil
	}
	depfiles := []string{}
	for depname := range depmap {
		depfiles = append(depfiles, depname)
	}
	sort.Strings(depfiles)
	sha := sha256.New()
	for _, depname := range depfiles {
		sha.Write([]byte(depname))
		if stat, err := os.Stat(depname); err != nil {
			exe.Fatal("stat depfile "+depname, err, ctxt.ExeContext)
		} else {
			curMtime := stat.ModTime().String()
			if prevMtime, hit := ctxt.CacheManifest["boot-mtime-"+depname]; !hit || curMtime != prevMtime {
				changed = true
			}
			updates["boot-mtime-"+depname] = curMtime
		}
	}
	hash := hex.EncodeToString(sha.Sum(nil)[:])
	updates["bootHash"] = hash
	if prevHash, hit := ctxt.CacheManifest["bootHash"]; !hit || prevHash != hash {
		changed = true
	}
	if !changed {
		ctxt.Logger.Info("Using cached boot image")
		return
	}

	// If we reach this point it means we can't use the cached image.
	dobootbuild(ctxt, srcfiles)
	for key, value := range updates {
		ctxt.CacheManifest[key] = value
	}
}

// writetoolsdef writes the edk2 tools definition file.
func writetoolsdef(ctxt *OsbuildContext) {
	if f, err := os.Create(path.Join(ctxt.Edk2ConfPath, "tools_def.txt")); err != nil {
		exe.Fatal("creating tools_def.txt", err, ctxt.ExeContext)
	} else {
		toolDefs, buildParams, toolParams := edk2toolparams(ctxt)
		for k, v := range toolDefs {
			line := fmt.Sprintf("DEFINE %-45s = %s\n", k, v)
			f.WriteString(line)
		}
		for k, v := range buildParams {
			line := fmt.Sprintf("*_SYS_*_*_%-42s = %s\n", k, v)
			f.WriteString(line)
		}
		for k, v := range toolParams {
			line := fmt.Sprintf("*_SYS_*_%-44s = %s\n", k, v)
			f.WriteString(line)
		}
		f.Close()
	}

	if f, err := os.Create(path.Join(ctxt.Edk2ConfPath, "target.txt")); err != nil {
		exe.Fatal("creating target.txt", err, ctxt.ExeContext)
	} else {
		f.WriteString("ACTIVE_PLATFORM = EmulatorPkg/EmulatorPkg.dsc\n")
		f.WriteString("TARGET = RELEASE\n")
		f.WriteString(fmt.Sprintf("TARGET_ARCH = %s\n", ctxt.ArchEFI))
		f.WriteString(fmt.Sprintf("TOOL_CHAIN_CONF = %s\n", "tools_def.txt"))
		f.WriteString("TOOL_CHAIN_TAG = SYS\n")
		f.WriteString(fmt.Sprintf("BUILD_RULE_CONF = %s\n", "build_rule.txt"))
		f.Close()
	}
}

// dobootbuild does the actual boot build steps without any cache interaction.
func dobootbuild(ctxt *OsbuildContext, srcfiles []string) {
	var guidPlatform uuid.UUID
	var guidFile uuid.UUID
	var guidPackage uuid.UUID
	var guidToken uuid.UUID
	setGuid := func(str string, guid *uuid.UUID) {
		if str != "" {
			if g, err := uuid.Parse(str); err != nil {
				exe.Fatal("guid "+str, err, ctxt.ExeContext)
			} else {
				*guid = g
			}
		} else {
			if g, err := uuid.NewRandom(); err != nil {
				exe.Fatal("random guid", err, ctxt.ExeContext)
			} else {
				*guid = g
			}
		}
	}
	setGuid(ctxt.Profile.EfiGuidPlatform, &guidPlatform)
	setGuid(ctxt.Profile.EfiGuidFile, &guidFile)
	setGuid(ctxt.Profile.EfiGuidPackage, &guidPackage)
	setGuid(ctxt.Profile.EfiGuidToken, &guidToken)

	modulePath := path.Join(ctxt.Edk2ConfPath, "SysModulePkg")
	applicationPath := path.Join(modulePath, "Application")

	if err := os.MkdirAll(applicationPath, 0755); err != nil {
		exe.Fatal("creating "+applicationPath, err, ctxt.ExeContext)
	}

	writeconfvars(ctxt, ctxt.TargetTriple)
	writetoolsdef(ctxt)

	writeEdk2Section := func(f *os.File, name string, vals []string) {
		f.WriteString(fmt.Sprintf("[%s]\n", name))
		for _, val := range vals {
			f.WriteString(fmt.Sprintf("  %s\n", val))
		}
		f.WriteString("\n")
	}
	writeEdk2SectionMap := func(f *os.File, name string, vals map[string]string) {
		lines := []string{}
		for key, val := range vals {
			lines = append(lines, fmt.Sprintf("%-42s = %s", key, val))
		}
		writeEdk2Section(f, name, lines)
	}

	// Write DSC file.
	if f, err := os.Create(path.Join(modulePath, "SysModulePkg.dsc")); err != nil {
		exe.Fatal("creating SysModulePkg.dsc", err, ctxt.ExeContext)
	} else {
		definesDsc, libraryClassesDsc, componentsDsc := edk2dscparams(&guidPlatform, ctxt)
		writeEdk2SectionMap(f, "Defines", definesDsc)
		writeEdk2Section(f, "Components", componentsDsc)
		writeEdk2Section(f, "LibraryClasses", libraryClassesDsc)
		f.Close()
	}

	// Write DEC file.
	if f, err := os.Create(path.Join(modulePath, "SysModulePkg.dec")); err != nil {
		exe.Fatal("creating SysModulePkg.dec", err, ctxt.ExeContext)
	} else {
		definesDec, guidsDec, includesDec := edk2decparams(&guidPackage, &guidToken, ctxt)
		writeEdk2SectionMap(f, "Defines", definesDec)
		writeEdk2Section(f, "Includes", includesDec)
		writeEdk2SectionMap(f, "Guids", guidsDec)
		f.Close()
	}

	// Write INF file.
	if f, err := os.Create(path.Join(applicationPath, "Sys.inf")); err != nil {
		exe.Fatal("creating Sys.inf", err, ctxt.ExeContext)
	} else {
		definesInf, packagesInf, libraryClassesInf, guidsInf, protocolsInf, sourcesInf := edk2infparams(&guidFile, ctxt)
		prefix := path.Join(ctxt.SrcRootDir, "sys-boot")
		for _, srcname := range srcfiles {
			relname := strings.TrimLeft(strings.TrimPrefix(srcname, prefix), "/\\")
			sourcesInf = append(sourcesInf, relname)
		}
		writeEdk2SectionMap(f, "Defines", definesInf)
		writeEdk2Section(f, "Packages", packagesInf)
		writeEdk2Section(f, "LibraryClasses", libraryClassesInf)
		writeEdk2Section(f, "Guids", guidsInf)
		writeEdk2Section(f, "Protocols", protocolsInf)
		writeEdk2Section(f, "Sources", sourcesInf)
		f.Close()
	}

	// Run the edk2 toolchain.
	ctxt.Logger.Info("Building boot image")
	for tries := 0; ; tries++ {
		os.RemoveAll(path.Join(ctxt.Edk2WorkspacePath, "Build"))
		edk2OutPath := path.Join(ctxt.Edk2WorkspacePath, "Build", "SysModuleOutput")
		if err := os.MkdirAll(edk2OutPath, 0755); err != nil {
			exe.Fatal("creating "+edk2OutPath, err, ctxt.ExeContext)
		}

		edk2SetupPath := path.Join(ctxt.CacheDir, "dep", "edk2", "edksetup.sh")
		buildCmd := fmt.Sprintf("build -n 0 -b RELEASE -t SYS -a %s -p %s -m %s", ctxt.ArchEFI,
			path.Join("edk2Conf", "SysModulePkg", "SysModulePkg.dsc"), path.Join("edk2Conf", "SysModulePkg", "Application", "Sys.inf"))
		cmd := fmt.Sprintf("source %s ; source %s ; %s", ctxt.Edk2VarsPath, edk2SetupPath, buildCmd)
		if stdOut, stdErr, err := exe.Doexec(ctxt.Edk2ConfPath, "bash", "-c", cmd); err != nil && tries < 2 {
			continue
		} else if err == nil {
			break
		} else {
			exe.Fatal("running edk2 build", exe.ErrOutput(stdOut, stdErr, err), ctxt.ExeContext)
		}
	}

}
