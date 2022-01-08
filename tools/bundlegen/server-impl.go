package main

import (
	api_os_container_bundle_v0 "alt-os/api/os/container/bundle/v0"
	"alt-os/exe"
	"context"
	"fmt"
	"oci/specs"
	"os"
	"path/filepath"

	"github.com/gogo/protobuf/types"
)

// newContainerBundleServiceServerImpl returns a new server-impl for bundlegen.
func newContainerBundleServiceServerImpl(ctxt *BundlegenContext) *ContainerBundleServiceServerImpl {
	return &ContainerBundleServiceServerImpl{
		ctxt: ctxt,
	}
}

type ContainerBundleServiceServerImpl struct {
	api_os_container_bundle_v0.UnimplementedContainerBundleServiceServer
	ctxt *BundlegenContext
}

func (server *ContainerBundleServiceServerImpl) ApiServe(ctx context.Context,
	in *api_os_container_bundle_v0.ApiServeRequest) (*types.Empty, error) {

	server.ctxt.rootDir = filepath.Clean(in.RootDir)
	return &types.Empty{}, nil
}

func (server *ContainerBundleServiceServerImpl) ApiUnserve(ctx context.Context,
	in *api_os_container_bundle_v0.ApiUnserveRequest) (*types.Empty, error) {

	addr := fmt.Sprintf("%s:%d", in.Hostname, in.Port)
	server.ctxt.AddrStopSignalMap[addr]()

	return &types.Empty{}, nil
}

func (server *ContainerBundleServiceServerImpl) Create(ctx context.Context,
	in *api_os_container_bundle_v0.CreateRequest) (*types.Empty, error) {

	bundleDir := filepath.Clean(filepath.Join(server.ctxt.rootDir, in.BundleDir))
	if err := os.MkdirAll(bundleDir, 0755); err != nil {
		exe.Fatal("making dir "+bundleDir, err, server.ctxt.ExeContext)
	}

	fmt.Println(specs.Version)

	return &types.Empty{}, nil
}
