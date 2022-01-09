package main

import (
	"alt-os/api"
	api_os_container_bundle_v0 "alt-os/api/os/container/bundle/v0"
	"alt-os/os/container"
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/gogo/protobuf/types"
)

// newContainerBundleServiceServerImpl returns a new server-impl for ct-bundle.
func newContainerBundleServiceServerImpl(ctxt *CtBundleContext) *ContainerBundleServiceServerImpl {
	return &ContainerBundleServiceServerImpl{
		ctxt: ctxt,
	}
}

type ContainerBundleServiceServerImpl struct {
	api_os_container_bundle_v0.UnimplementedContainerBundleServiceServer
	ctxt *CtBundleContext
}

func (server *ContainerBundleServiceServerImpl) ApiServe(ctx context.Context,
	in *api_os_container_bundle_v0.ApiServeRequest) (*types.Empty, error) {

	server.ctxt.rootDir = filepath.Clean(in.RootDir)
	return &types.Empty{}, nil
}

func (server *ContainerBundleServiceServerImpl) ApiUnserve(ctx context.Context,
	in *api_os_container_bundle_v0.ApiUnserveRequest) (*types.Empty, error) {

	addr := fmt.Sprintf("%s:%d", in.ApiHostname, in.ApiPort)
	server.ctxt.AddrStopSignalMap[addr]()

	return &types.Empty{}, nil
}

func (server *ContainerBundleServiceServerImpl) Create(ctx context.Context,
	in *api_os_container_bundle_v0.CreateRequest) (*types.Empty, error) {

	// Load/verify at least one bundle definition to create.
	var bundles []*api_os_container_bundle_v0.Bundle
	if in.Bundles == nil && in.BundlesFile == "" {
		return &types.Empty{}, errors.New("missing bundle definitions")
	} else if in.Bundles != nil && in.BundlesFile != "" {
		return &types.Empty{}, errors.New("multiple bundle definition fields")
	}
	if in.Bundles != nil {
		bundles = in.Bundles
	} else if messages, err := api.UnmarshalApiProtoMessages(in.BundlesFile, ""); err != nil {
		return &types.Empty{}, err
	} else {
		for _, msg := range messages {
			kindVer := msg.Kind + "/" + msg.Version
			badTypeErr := errors.New("bad bundle definition message type")
			switch kindVer {
			default:
				return &types.Empty{}, errors.New("unrecognized bundle definition message " + kindVer)
			case "os.container.bundle.Bundle/v0":
				if msgDef, ok := msg.Def.(*api_os_container_bundle_v0.Bundle); !ok {
					return &types.Empty{}, badTypeErr
				} else {
					bundles = append(bundles, msgDef)
				}
			}
		}
	}
	if len(bundles) < 1 {
		return &types.Empty{}, errors.New("empty bundle definitions")
	}

	// Create the requested bundles and make sure there are no duplicate directories.
	sawDir := map[string]struct{}{}
	for _, bundle := range bundles {
		if _, ok := sawDir[bundle.BundleDir]; ok {
			return &types.Empty{}, errors.New("duplicate bundle dir: " + bundle.BundleDir)
		}
		if err := container.CreateBundle(bundle, server.ctxt.rootDir); err != nil {
			return &types.Empty{}, err
		}
		sawDir[bundle.BundleDir] = struct{}{}
	}

	return &types.Empty{}, nil
}
