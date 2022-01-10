package main

import (
	api_os_machine_runtime_v0 "alt-os/api/os/machine/runtime/v0"
	"context"
	"fmt"

	"github.com/gogo/protobuf/types"
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

	fmt.Println("serving")
	// TODO start vmm

	return &types.Empty{}, nil
}

func (server *VmmRuntimeServiceServerImpl) ApiUnserve(ctx context.Context,
	in *api_os_machine_runtime_v0.ApiUnserveRequest) (*types.Empty, error) {

	fmt.Println("unserving")
	// TODO stop and delete all virtual machines
	addr := fmt.Sprintf("%s:%d", in.ApiHostname, in.ApiPort)
	server.ctxt.AddrStopSignalMap[addr]()

	return &types.Empty{}, nil
}

func (server *VmmRuntimeServiceServerImpl) List(ctx context.Context,
	in *api_os_machine_runtime_v0.ListRequest) (*api_os_machine_runtime_v0.ListResponse, error) {
	fmt.Println("listing")
	return &api_os_machine_runtime_v0.ListResponse{}, nil
}

func (server *VmmRuntimeServiceServerImpl) QueryState(ctx context.Context,
	in *api_os_machine_runtime_v0.QueryStateRequest) (*types.Empty, error) {
	return nil, nil
}

func (server *VmmRuntimeServiceServerImpl) Create(ctx context.Context,
	in *api_os_machine_runtime_v0.CreateRequest) (*types.Empty, error) {
	return nil, nil
}

func (server *VmmRuntimeServiceServerImpl) Start(ctx context.Context,
	in *api_os_machine_runtime_v0.StartRequest) (*types.Empty, error) {
	return nil, nil
}

func (server *VmmRuntimeServiceServerImpl) Kill(ctx context.Context,
	in *api_os_machine_runtime_v0.KillRequest) (*types.Empty, error) {
	return nil, nil
}

func (server *VmmRuntimeServiceServerImpl) Delete(ctx context.Context,
	in *api_os_machine_runtime_v0.DeleteRequest) (*types.Empty, error) {
	return nil, nil
}
