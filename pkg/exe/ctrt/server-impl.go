package main

import (
	api_ctrt_v0 "alt-os/api/ctrt/v0"
	"context"
	"fmt"

	"github.com/gogo/protobuf/types"
)

type ContainerRuntimeServerImpl struct {
	api_ctrt_v0.UnimplementedContainerRuntimeServer
	ctxt *CtrtContext
}

// NewContainerRuntimeServerImpl returns a new server-impl for ctrt.
func NewContainerRuntimeServerImpl(ctxt *CtrtContext) *ContainerRuntimeServerImpl {
	return &ContainerRuntimeServerImpl{
		ctxt: ctxt,
	}
}

func (server *ContainerRuntimeServerImpl) ApiServe(ctx context.Context,
	in *api_ctrt_v0.ApiServeRequest) (*types.Empty, error) {

	fmt.Println("serving")
	// TODO start host

	return &types.Empty{}, nil
}

func (server *ContainerRuntimeServerImpl) ApiUnserve(ctx context.Context,
	in *api_ctrt_v0.ApiUnserveRequest) (*types.Empty, error) {

	fmt.Println("unserving")
	// TODO stop and delete all containers
	addr := fmt.Sprintf("%s:%d", in.Hostname, in.Port)
	server.ctxt.AddrStopSignalMap[addr]()

	return &types.Empty{}, nil
}

func (server *ContainerRuntimeServerImpl) List(ctx context.Context,
	in *api_ctrt_v0.ListRequest) (*api_ctrt_v0.ListResponse, error) {
	fmt.Println("listing")
	return &api_ctrt_v0.ListResponse{}, nil
}

func (server *ContainerRuntimeServerImpl) QueryState(ctx context.Context,
	in *api_ctrt_v0.QueryStateRequest) (*types.Empty, error) {
	return nil, nil
}

func (server *ContainerRuntimeServerImpl) Create(ctx context.Context,
	in *api_ctrt_v0.CreateRequest) (*types.Empty, error) {
	return nil, nil
}

func (server *ContainerRuntimeServerImpl) Start(ctx context.Context,
	in *api_ctrt_v0.StartRequest) (*types.Empty, error) {
	return nil, nil
}

func (server *ContainerRuntimeServerImpl) Kill(ctx context.Context,
	in *api_ctrt_v0.KillRequest) (*types.Empty, error) {
	return nil, nil
}

func (server *ContainerRuntimeServerImpl) Delete(ctx context.Context,
	in *api_ctrt_v0.DeleteRequest) (*types.Empty, error) {
	return nil, nil
}
