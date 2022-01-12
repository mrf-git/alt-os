package main

import (
	api_os_machine_runtime_v0 "alt-os/api/os/machine/runtime/v0"
	"alt-os/os/code"
	"context"
	"fmt"
	"path/filepath"

	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// newVmmRuntimeServiceServerImpl returns a new server-impl for vmm-runtime.
func newVmmRuntimeServiceServerImpl(ctxt *VmmRuntimeContext) *VmmRuntimeServiceServerImpl {
	return &VmmRuntimeServiceServerImpl{
		ctxt: ctxt,
	}
}

type VmmRuntimeServiceServerImpl struct {
	api_os_machine_runtime_v0.UnimplementedVmmRuntimeServiceServer
	ctxt *VmmRuntimeContext
}

func (server *VmmRuntimeServiceServerImpl) ApiServe(ctx context.Context,
	in *api_os_machine_runtime_v0.ApiServeRequest) (*types.Empty, error) {

	if !isVmxSupported() {
		return &types.Empty{}, status.Errorf(codes.FailedPrecondition, "vmx features not enabled on host cpu")
	}
	server.ctxt.imageDir = filepath.Clean(in.ImageDir)
	if server.ctxt.imageDir == "" {
		return &types.Empty{}, status.Errorf(codes.InvalidArgument, "missing imageDir")
	}
	server.ctxt.maxMachines = int(in.MaxMachines)
	if server.ctxt.maxMachines == 0 {
		return &types.Empty{}, status.Errorf(codes.InvalidArgument, "missing maxMachines")
	}
	return &types.Empty{}, nil
}

func (server *VmmRuntimeServiceServerImpl) ApiUnserve(ctx context.Context,
	in *api_os_machine_runtime_v0.ApiUnserveRequest) (*types.Empty, error) {

	// TODO stop and delete all virtual machines
	addr := fmt.Sprintf("%s:%d", in.ApiHostname, in.ApiPort)
	server.ctxt.AddrStopSignalMap[addr]()

	return &types.Empty{}, nil
}

func (server *VmmRuntimeServiceServerImpl) List(ctx context.Context,
	in *api_os_machine_runtime_v0.ListRequest) (*api_os_machine_runtime_v0.ListResponse, error) {
	// TODO
	return &api_os_machine_runtime_v0.ListResponse{}, nil
}

func (server *VmmRuntimeServiceServerImpl) QueryState(ctx context.Context,
	in *api_os_machine_runtime_v0.QueryStateRequest) (*api_os_machine_runtime_v0.QueryStateResponse, error) {
	// TODO
	return &api_os_machine_runtime_v0.QueryStateResponse{}, nil
}

func (server *VmmRuntimeServiceServerImpl) Create(ctx context.Context,
	in *api_os_machine_runtime_v0.CreateRequest) (*types.Empty, error) {

	if len(server.ctxt.vmEnvs) >= server.ctxt.maxMachines {
		return &types.Empty{}, status.Errorf(codes.ResourceExhausted, "at maxMachines")
	}
	initPath := filepath.Clean(filepath.Join(server.ctxt.imageDir, in.Image, "init.code"))
	if exeCode, err := code.FromFile(initPath); err != nil {
		return &types.Empty{}, status.Errorf(codes.InvalidArgument,
			"could not read %s: %s", initPath, err.Error())
	} else {

		fmt.Println("size", exeCode.GetSize())
	}

	// TODO send code to a new vm
	// server.ctxt.vmEnvs[in.Id]

	return &types.Empty{}, nil
}

func (server *VmmRuntimeServiceServerImpl) Start(ctx context.Context,
	in *api_os_machine_runtime_v0.StartRequest) (*types.Empty, error) {

	// chReturnCode chan int
	// TODO
	return &types.Empty{}, nil
}

func (server *VmmRuntimeServiceServerImpl) Kill(ctx context.Context,
	in *api_os_machine_runtime_v0.KillRequest) (*types.Empty, error) {

	// TODO
	return &types.Empty{}, nil
}

func (server *VmmRuntimeServiceServerImpl) Delete(ctx context.Context,
	in *api_os_machine_runtime_v0.DeleteRequest) (*types.Empty, error) {

	// TODO
	return &types.Empty{}, nil
}
