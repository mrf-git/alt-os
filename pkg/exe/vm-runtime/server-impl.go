package main

import (
	api_os_machine_runtime_v0 "alt-os/api/os/machine/runtime/v0"
	"alt-os/os/code"
	"alt-os/os/limits"
	"context"
	"fmt"
	"path/filepath"

	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// newVmRuntimeServiceServerImpl returns a new server-impl for vm-runtime.
func newVmRuntimeServiceServerImpl(ctxt *VmRuntimeContext) *VmRuntimeServiceServerImpl {
	return &VmRuntimeServiceServerImpl{
		ctxt: ctxt,
	}
}

type VmRuntimeServiceServerImpl struct {
	api_os_machine_runtime_v0.UnimplementedVmRuntimeServiceServer
	ctxt *VmRuntimeContext
}

func (server *VmRuntimeServiceServerImpl) ApiServe(ctx context.Context,
	in *api_os_machine_runtime_v0.ApiServeRequest) (*types.Empty, error) {

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

func (server *VmRuntimeServiceServerImpl) ApiUnserve(ctx context.Context,
	in *api_os_machine_runtime_v0.ApiUnserveRequest) (*types.Empty, error) {

	// TODO stop and delete all virtual machines
	addr := fmt.Sprintf("%s:%d", in.ApiHostname, in.ApiPort)
	server.ctxt.AddrStopSignalMap[addr]()

	return &types.Empty{}, nil
}

func (server *VmRuntimeServiceServerImpl) List(ctx context.Context,
	in *api_os_machine_runtime_v0.ListRequest) (*api_os_machine_runtime_v0.ListResponse, error) {
	// TODO
	return &api_os_machine_runtime_v0.ListResponse{}, nil
}

func (server *VmRuntimeServiceServerImpl) QueryState(ctx context.Context,
	in *api_os_machine_runtime_v0.QueryStateRequest) (*api_os_machine_runtime_v0.QueryStateResponse, error) {
	// TODO
	return &api_os_machine_runtime_v0.QueryStateResponse{}, nil
}

func (server *VmRuntimeServiceServerImpl) Create(ctx context.Context,
	in *api_os_machine_runtime_v0.CreateRequest) (*types.Empty, error) {

	if len(server.ctxt.vmEnvs) >= server.ctxt.maxMachines {
		return &types.Empty{}, status.Errorf(codes.ResourceExhausted, "at maxMachines")
	}
	initPath := filepath.Clean(filepath.Join(server.ctxt.imageDir, in.Image, "init.code"))
	if exeCode, err := code.FromFile(initPath); err != nil {
		return &types.Empty{}, status.Errorf(codes.InvalidArgument,
			"could not read %s: %s", initPath, err.Error())
	} else {
		vmEnv := newVmEnvironment(exeCode, server.ctxt)
		server.ctxt.vmEnvs[in.Id] = vmEnv
	}

	return &types.Empty{}, nil
}

func (server *VmRuntimeServiceServerImpl) Start(ctx context.Context,
	in *api_os_machine_runtime_v0.StartRequest) (*types.Empty, error) {

	vmEnv, ok := server.ctxt.vmEnvs[in.Id]
	if !ok {
		return &types.Empty{}, status.Errorf(codes.NotFound, in.Id)
	}
	if _, ok = server.ctxt.vmRetChs[in.Id]; ok {
		return &types.Empty{}, status.Errorf(codes.AlreadyExists, in.Id)
	}
	signalCh := make(chan int, limits.MAX_PROCESS_SIGNALS)
	returnCodeCh := make(chan int, 1)
	if err := vmEnv.Run(signalCh, returnCodeCh); err != nil {
		return &types.Empty{}, status.Errorf(codes.Internal, err.Error())
	}
	server.ctxt.vmSigChs[in.Id] = signalCh
	server.ctxt.vmRetChs[in.Id] = returnCodeCh

	return &types.Empty{}, nil
}

func (server *VmRuntimeServiceServerImpl) Kill(ctx context.Context,
	in *api_os_machine_runtime_v0.KillRequest) (*types.Empty, error) {

	_, ok := server.ctxt.vmEnvs[in.Id]
	if !ok {
		return &types.Empty{}, status.Errorf(codes.NotFound, in.Id)
	}
	vmSigCh, ok := server.ctxt.vmSigChs[in.Id]
	if !ok {
		return &types.Empty{}, status.Errorf(codes.FailedPrecondition,
			"%s not started", in.Id)
	}
	select {
	default:
		return &types.Empty{}, status.Errorf(codes.ResourceExhausted, in.Id)
	case vmSigCh <- int(in.Signal):
	}

	return &types.Empty{}, nil
}

func (server *VmRuntimeServiceServerImpl) Delete(ctx context.Context,
	in *api_os_machine_runtime_v0.DeleteRequest) (*types.Empty, error) {

	// TODO
	return &types.Empty{}, nil
}
