package main

import (
	api_os_build_v0 "alt-os/api/os/build/v0"
	"alt-os/exe"
	"os/user"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
)

// toolinit initializes all compiler/linker flags and toolchain parameters
// according to the selected profile.
func toolinit(ctxt *OsbuildContext) {
	flagsBase := []string{
		"-g", "-O2", "-fPIC", "-nodefaultlibs", "-nostdlib", "-mno-red-zone",
		"-fno-strict-aliasing", "-fno-ms-extensions", "-fno-common", "-fvisibility=hidden",
		"-mms-bitfields", "-Wall", "-Werror", "-Wno-empty-body", "-Wno-unused-const-variable", "-Wno-varargs",
		"-Wno-unknown-warning-option", "-Wno-address", "-Wno-shift-negative-value", "-Wno-unknown-pragmas",
		"-Wno-incompatible-library-redeclaration", "-Wno-null-dereference",
	}
	profileDef := strings.ReplaceAll(strings.ToUpper(ctxt.Profile.Name), "-", "_")
	flagsBase = append(flagsBase, "-DOS_BUILD_PROFILE="+ctxt.Profile.Name)
	flagsBase = append(flagsBase, "-DOS_BUILD_PROFILE_"+profileDef+"=1")
	flagsBaseBoot := []string{
		"-ffreestanding", "-static", "-fno-stack-protector", "-fshort-wchar", "-fno-asynchronous-unwind-tables",
		"-mno-implicit-float", "-mcmodel=small", "-fno-builtin",
		"-funsigned-char", "-DUSE_MS_ABI=1",
	}
	flagsBaseBoot = append(flagsBaseBoot, flagsBase...)
	flagsBaseRuntime := []string{}
	flagsBaseRuntime = append(flagsBaseRuntime, flagsBase...)

	ctxt.FlagsCCRuntime = append(ctxt.FlagsCCRuntime, flagsBaseRuntime...)
	ctxt.FlagsCCRuntime = append(ctxt.FlagsCCRuntime, "-I"+ctxt.CacheDir)
	ctxt.FlagsCCRuntime = append(ctxt.FlagsCCRuntime, "-I"+path.Join(ctxt.SrcRootDir, "sys", "include"))
	ctxt.FlagsCCRes = append(ctxt.FlagsCCRes, flagsBaseBoot...)
	ctxt.FlagsCCRes = append(ctxt.FlagsCCRes, "-I"+path.Join(ctxt.SrcRootDir, "sys", "include"))
	ctxt.FlagsCCRes = append(ctxt.FlagsCCRes, "-I"+path.Join(ctxt.CacheDir, "dep", "edk2", "MdeModulePkg", "Include"))
	ctxt.FlagsCCRes = append(ctxt.FlagsCCRes, "-I"+path.Join(ctxt.CacheDir, "dep", "edk2", "MdePkg", "Include"))
	ctxt.FlagsCCRes = append(ctxt.FlagsCCRes, "-I"+path.Join(ctxt.CacheDir, "dep", "edk2", "RedfishPkg", "Include"))
	ctxt.FlagsCCBoot = append(ctxt.FlagsCCBoot, flagsBaseBoot...)
	ctxt.FlagsCCBoot = append(ctxt.FlagsCCBoot, "-include", "AutoGen.h")

	switch ctxt.Profile.Arch {
	case "amd64":
		ctxt.FlagsCCRes = append(ctxt.FlagsCCRes, "-I"+path.Join(ctxt.CacheDir, "dep", "edk2", "MdePkg", "Include", "X64"))
		ctxt.FlagsCCRuntime = append(ctxt.FlagsCCRuntime, "-I"+path.Join(ctxt.CacheDir, "dep", "edk2", "MdePkg", "Include", "X64"))
		ctxt.FlagsCCBoot = append(ctxt.FlagsCCBoot, "-I"+path.Join(ctxt.CacheDir, "dep", "edk2", "MdePkg", "Include", "X64"))
		ccBoot := []string{"--target=x86_64-pc-linux", "-m64", "-mno-sse", "-mno-mmx", "-msoft-float",
			"-DEFIAPI=\"__attribute__((ms_abi))\""}
		linkBoot := []string{"-fuse-ld=lld", "--target=x86_64-pc-linux", "-Wl,-melf_x86_64", "-Wl,--oformat,elf64-x86-64"}
		ccRuntime := []string{"--target=x86_64-pc-alt"}
		linkRuntime := []string{"-nodefaultlibs", "-nostdlib", "-fuse-ld=lld", "--target=x86_64-pc-alt",
			"-Wl,--dynamic-linker=boot", "-Wl,--strip-debug", "-Wl,--emit-relocs", "-Wl,-shared", "-Wl,--no-pie"}
		ctxt.FlagsCCRuntime = append(ctxt.FlagsCCRuntime, ccRuntime...)
		ctxt.FlagsLinkRuntime = append(ctxt.FlagsLinkRuntime, linkRuntime...)
		ctxt.FlagsCCRes = append(ctxt.FlagsCCRes, ccBoot...)
		ctxt.FlagsCCBoot = append(ctxt.FlagsCCBoot, ccBoot...)
		ctxt.FlagsCCBootArch = ccBoot
		ctxt.FlagsLinkBoot = append(ctxt.FlagsLinkBoot, linkBoot...)
		ctxt.FlagsLinkBootArch = linkBoot
		ctxt.FlagsRCBoot = append(ctxt.FlagsRCBoot, "-O", "elf64-x86-64", "-B", "i386")
		ctxt.FlagsNasmBoot = append(ctxt.FlagsNasmBoot, "-f", "elf64")
		ctxt.TargetTriple = "x86_64-pc-alt"

	case "aarch64":
		ctxt.FlagsCCRes = append(ctxt.FlagsCCRes, "-I"+path.Join(ctxt.CacheDir, "dep", "edk2", "MdePkg", "Include", "AArch64"))
		ctxt.FlagsCCRuntime = append(ctxt.FlagsCCRuntime, "-I"+path.Join(ctxt.CacheDir, "dep", "edk2", "MdePkg", "Include", "AArch64"))
		ctxt.FlagsCCBoot = append(ctxt.FlagsCCBoot, "-I"+path.Join(ctxt.CacheDir, "dep", "edk2", "MdePkg", "Include", "AArch64"))
		ccBoot := []string{"--target=aarch64-pc-linux", "-m64", "-DEFIAPI="}
		linkBoot := []string{"-fuse-ld=lld", "--target=aarch64-pc-linux", "-Wl,-melf_aarch64", "-Wl,--oformat,elf64-littleaarch64"}
		ccRuntime := []string{"--target=aarch64-pc-alt"}
		linkRuntime := []string{"-nodefaultlibs", "-nostdlib", "-fuse-ld=lld", "--target=aarch64-pc-alt",
			"-Wl,--dynamic-linker=boot", "-Wl,--strip-debug", "-Wl,--emit-relocs", "-Wl,-shared", "-Wl,--no-pie"}
		ctxt.FlagsCCRuntime = append(ctxt.FlagsCCRuntime, ccRuntime...)
		ctxt.FlagsLinkRuntime = append(ctxt.FlagsLinkRuntime, linkRuntime...)
		ctxt.FlagsCCRes = append(ctxt.FlagsCCRes, ccBoot...)
		ctxt.FlagsCCBoot = append(ctxt.FlagsCCBoot, ccBoot...)
		ctxt.FlagsCCBootArch = ccBoot
		ctxt.FlagsLinkBoot = append(ctxt.FlagsLinkBoot, linkBoot...)
		ctxt.FlagsLinkBootArch = linkBoot
		ctxt.FlagsRCBoot = append(ctxt.FlagsRCBoot, "-O", "elf64-littleaarch64", "-B", "aarch64")
		ctxt.FlagsNasmBoot = append(ctxt.FlagsNasmBoot, "-f", "elf64")
		ctxt.TargetTriple = "aarch64-pc-alt"
	}
}

// infoinit initializes the build information that persists with the built image artifact.
// Also verifies we can run the git command.
func infoinit(ctxt *OsbuildContext) {
	var username string
	if user, err := user.Current(); err != nil {
		exe.Fatal("getting user", err, ctxt.ExeContext)
	} else {
		username = user.Username
	}
	ctxt.BuildInfo = &api_os_build_v0.BuildInfo{
		VersionStr:      "v0.0.0-prototype",
		VersionMajorNum: 0,
		VersionMinorNum: 0,
		RevisionNum:     0,
		Uuid:            uuid.Must(uuid.NewRandom()).String(),
		Scm:             gitinfo(ctxt),
		Timestamp:       uint64(time.Now().UTC().Unix()),
		Username:        username,
		GoVersionStr:    runtime.Version(),
		HostOs:          runtime.GOOS,
		HostArch:        runtime.GOARCH,
	}
}
