package main

import (
	"alt-os/api"
	api_os_machine_image_v0 "alt-os/api/os/machine/image/v0"
	"alt-os/os/machine"
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/gogo/protobuf/types"
)

// newVmmImageServiceServerImpl returns a new server-impl for vmm-image.
func newVmmImageServiceServerImpl(ctxt *VmmImageContext) *VmmImageServiceServerImpl {
	return &VmmImageServiceServerImpl{
		ctxt: ctxt,
	}
}

type VmmImageServiceServerImpl struct {
	api_os_machine_image_v0.UnimplementedVmmImageServiceServer
	ctxt *VmmImageContext
}

func (server *VmmImageServiceServerImpl) ApiServe(ctx context.Context,
	in *api_os_machine_image_v0.ApiServeRequest) (*types.Empty, error) {

	server.ctxt.rootDir = filepath.Clean(in.RootDir)
	return &types.Empty{}, nil
}

func (server *VmmImageServiceServerImpl) ApiUnserve(ctx context.Context,
	in *api_os_machine_image_v0.ApiUnserveRequest) (*types.Empty, error) {

	addr := fmt.Sprintf("%s:%d", in.ApiHostname, in.ApiPort)
	server.ctxt.AddrStopSignalMap[addr]()

	return &types.Empty{}, nil
}

func (server *VmmImageServiceServerImpl) Create(ctx context.Context,
	in *api_os_machine_image_v0.CreateRequest) (*types.Empty, error) {

	// Load/verify at least one virtual machine definition to create.
	var machines []*api_os_machine_image_v0.VirtualMachine
	if in.VirtualMachines == nil && in.VirtualMachinesFile == "" {
		return &types.Empty{}, errors.New("missing virtual machine definitions")
	} else if in.VirtualMachines != nil && in.VirtualMachinesFile != "" {
		return &types.Empty{}, errors.New("multiple virtual machine definition fields")
	}
	if in.VirtualMachines != nil {
		machines = in.VirtualMachines
	} else if messages, err := api.UnmarshalApiProtoMessages(in.VirtualMachinesFile, ""); err != nil {
		return &types.Empty{}, err
	} else {
		for _, msg := range messages {
			kindVer := msg.Kind + "/" + msg.Version
			badTypeErr := errors.New("bad virtual machine definition message type")
			switch kindVer {
			default:
				return &types.Empty{}, errors.New("unrecognized virtual machine definition message " + kindVer)
			case "os.machine.image.VirtualMachine/v0":
				if msgDef, ok := msg.Def.(*api_os_machine_image_v0.VirtualMachine); !ok {
					return &types.Empty{}, badTypeErr
				} else {
					machines = append(machines, msgDef)
				}
			}
		}
	}
	if len(machines) < 1 {
		return &types.Empty{}, errors.New("empty virtual machine definitions")
	}

	// Create the requested virtual machines and make sure there are no duplicate directories.
	sawDir := map[string]struct{}{}
	for _, vm := range machines {
		if _, ok := sawDir[vm.ImageDir]; ok {
			return &types.Empty{}, errors.New("duplicate virtual machine image dir: " + vm.ImageDir)
		}
		if err := machine.CreateImage(vm, server.ctxt.rootDir); err != nil {
			return &types.Empty{}, err
		}
		sawDir[vm.ImageDir] = struct{}{}
	}

	return &types.Empty{}, nil
}
