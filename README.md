README
======

Next-generation operating system - **codename: alt.os**.

### Lifecycle status

Version        | Status
-------------- | ------
v0.0.0         | PROTOTYPE


## Running

### Requirements
* VMM support requires nestable hardware virtualization features (VT-x/AMD-V)


## Building

### Requirements
* [go-alt](https://github.com/mrf-git/go-alt/blob/feature/initialPort/README-alt.md) distribution of Go installed as the system Go
* [llvm-alt](https://github.com/mrf-git/llvm-alt/blob/feature/initialPort/README-alt.md) distribution of the LLVM compiler toolchain installed as the system clang
* Any dependency Go modules required by the `go.mod` files in this repo
* A [protobuf](https://developers.google.com/protocol-buffers) compiler

### Instructions
While it shouldn't be strictly required, it may be beneficial to download all module dependencies right away. To do this, from the root of the source directory run:
```
go mod download all
```

Before building, first ensure all auto-generated code is up to date by following the steps in the "Code generation" section below.


#### Host platform instructions: Windows
Building on Windows requires [mingw-w64](https://www.mingw-w64.org) with the [mingw-w64-x86_64-clang](https://packages.msys2.org/package/mingw-w64-x86_64-clang) package installed.

Next, the following environment variables should be set:
```
export PATH=/mingw64/bin:$PATH
export CGO_ENABLED=1
export CC=clang
export CGO_CFLAGS="-fuse-ld=/mingw64/x86_64-w64-mingw32/bin/ld.exe"
```


## Code generation
All code generation is done through the `codegen` tool. First, from the root of the source tree, build the tool:
```
go build -o ./workspace/tools/codegen ./tools/codegen
```

Then, to generate the protocol buffer source code, run:
```
./workspace/tools/codegen -p
```

You can use the defined tasks from VS Code to do these steps.

Next all dependent code must be rebuilt by following the steps in the "Building" section above.

## Licensing

The OS is under a proprietary license. Third-party licenses are provided where sources appear.

### Disclaimer
Copyright © 2022. All rights reserved.

PROPRIETARY AND CONFIDENTIAL. UNAUTHORIZED ACCESS PROHIBITED.
