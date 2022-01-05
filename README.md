README
======

Next-generation operating system - **codename: alt.os**.

### Lifecycle status

Version        | Status
-------------- | ------
v0.0.0         | PROTOTYPE


## Building

### Requirements
* [go-alt](https://github.com/mrf-git/go-alt/blob/feature/initialPort/README-alt.md) distribution of Go installed as the system Go
* Any dependency Go modules required by the `go.mod` files in this repo
* A [protobuf](https://developers.google.com/protocol-buffers) compiler

### Instructions
While it shouldn't be strictly required, it may be beneficial to download all module dependencies right away. To do this, from the root of the source directory run:
```
go mod download all
```

Before building, first ensure all auto-generated code is up to date by following the steps in the "Code generation" section below.

Next,
```
TODO
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

Alternatively from VS Code you can use the "Build Tool: codegen" and "Generate Api Protos" tasks.

Next all dependent code must be rebuilt by following the steps in the "Building" section above.

## Licensing

The OS is under a proprietary license. Third-party licenses are provided where sources appear.

### Disclaimer
Copyright © 2022. All rights reserved.

PROPRIETARY AND CONFIDENTIAL. UNAUTHORIZED ACCESS PROHIBITED.
