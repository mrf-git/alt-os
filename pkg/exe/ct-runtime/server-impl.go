package main

import (
	api_os_container_runtime_v0 "alt-os/api/os/container/runtime/v0"
	"context"
	"fmt"

	"github.com/gogo/protobuf/types"
)

// newContainerRuntimeServiceServerImpl returns a new server-impl for ct-runtime.
func newContainerRuntimeServiceServerImpl(ctxt *CtRuntimeContext) *ContainerRuntimeServiceServerImpl {
	return &ContainerRuntimeServiceServerImpl{
		ctxt: ctxt,
	}
}

type ContainerRuntimeServiceServerImpl struct {
	api_os_container_runtime_v0.UnimplementedContainerRuntimeServiceServer
	ctxt *CtRuntimeContext
}

func (server *ContainerRuntimeServiceServerImpl) ApiServe(ctx context.Context,
	in *api_os_container_runtime_v0.ApiServeRequest) (*types.Empty, error) {

	fmt.Println("serving")
	// TODO start host

	return &types.Empty{}, nil
}

func (server *ContainerRuntimeServiceServerImpl) ApiUnserve(ctx context.Context,
	in *api_os_container_runtime_v0.ApiUnserveRequest) (*types.Empty, error) {

	fmt.Println("unserving")
	// TODO stop and delete all containers
	addr := fmt.Sprintf("%s:%d", in.ApiHostname, in.ApiPort)
	server.ctxt.AddrStopSignalMap[addr]()

	return &types.Empty{}, nil
}

func (server *ContainerRuntimeServiceServerImpl) List(ctx context.Context,
	in *api_os_container_runtime_v0.ListRequest) (*api_os_container_runtime_v0.ListResponse, error) {
	fmt.Println("listing")
	return &api_os_container_runtime_v0.ListResponse{}, nil
}

func (server *ContainerRuntimeServiceServerImpl) QueryState(ctx context.Context,
	in *api_os_container_runtime_v0.QueryStateRequest) (*types.Empty, error) {
	return nil, nil
}

func (server *ContainerRuntimeServiceServerImpl) Create(ctx context.Context,
	in *api_os_container_runtime_v0.CreateRequest) (*types.Empty, error) {
	return nil, nil
}

func (server *ContainerRuntimeServiceServerImpl) Start(ctx context.Context,
	in *api_os_container_runtime_v0.StartRequest) (*types.Empty, error) {
	return nil, nil
}

func (server *ContainerRuntimeServiceServerImpl) Kill(ctx context.Context,
	in *api_os_container_runtime_v0.KillRequest) (*types.Empty, error) {
	return nil, nil
}

func (server *ContainerRuntimeServiceServerImpl) Delete(ctx context.Context,
	in *api_os_container_runtime_v0.DeleteRequest) (*types.Empty, error) {
	return nil, nil
}
