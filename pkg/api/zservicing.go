// Code generated by codegen. DO NOT EDIT.
package api

import (
	api_ctrt_v0 "alt-os/api/ctrt/v0"
	"errors"
	"fmt"
)

// serviceMessageKinds processes each specific message kind.
func serviceMessageKinds(ctxt *ApiServiceContext) error {
	for _, apiMsg := range ctxt.MessageQueue {
		switch msg := apiMsg.Def.(type) {
		default:
			return errors.New("invalid message kind: " + apiMsg.Kind)
		case *api_ctrt_v0.ApiServeRequest:
			if err := newServer("api.ctrt.ContainerRuntime", "v0", msg, ctxt); err != nil {
				return err
			} else if err := req_ContainerRuntime_v0_ApiServe(msg, ctxt); err != nil {
				return err
			}
		case *api_ctrt_v0.ApiUnserveRequest:
			if err := req_ContainerRuntime_v0_ApiUnserve(msg, ctxt); err != nil {
				return err
			} else if err := stopServer(fmt.Sprintf("%s:%d", msg.GetHostname(), msg.GetPort()), ctxt); err != nil {
				return err
			}
		case *api_ctrt_v0.ListRequest:
			if err := req_ContainerRuntime_v0_List(msg, ctxt); err != nil {
				return err
			}
		case *api_ctrt_v0.QueryStateRequest:
			if err := req_ContainerRuntime_v0_QueryState(msg, ctxt); err != nil {
				return err
			}
		case *api_ctrt_v0.CreateRequest:
			if err := req_ContainerRuntime_v0_Create(msg, ctxt); err != nil {
				return err
			}
		case *api_ctrt_v0.StartRequest:
			if err := req_ContainerRuntime_v0_Start(msg, ctxt); err != nil {
				return err
			}
		case *api_ctrt_v0.KillRequest:
			if err := req_ContainerRuntime_v0_Kill(msg, ctxt); err != nil {
				return err
			}
		case *api_ctrt_v0.DeleteRequest:
			if err := req_ContainerRuntime_v0_Delete(msg, ctxt); err != nil {
				return err
			}
		}
	}
	for addr := range ctxt.AddrServerMap {
		if err := ctxt.AddrStopWaitMap[addr](); err != nil {
			return nil
		}
	}
	ctxt.ServerWg.Wait()
	return nil
}

// makeClientKind makes a new specific grpc client kind in the context.
func makeClientKind(addr string, ctxt *ApiServiceContext) error {
	switch ctxt.AddrKindVerMap[addr] {
	default:
		return errors.New("invalid client kindVer: " + ctxt.AddrKindVerMap[addr])
	case "api.ctrt.ContainerRuntime/v0":
		ctxt.AddrClientMap[addr] = api_ctrt_v0.NewContainerRuntimeClient(ctxt.AddrGrpcConnMap[addr])
	}
	if ctxt.AddrClientMap[addr] == nil {
		return errors.New("failed to make client " + ctxt.AddrKindVerMap[addr] + " for " + addr)
	}
	return nil
}

// makeServerKind makes a new specific grpc server kind in the context.
func makeServerKind(addr string, ctxt *ApiServiceContext) error {
	switch ctxt.AddrKindVerMap[addr] {
	default:
		return errors.New("invalid impl kindVer: " + ctxt.AddrKindVerMap[addr])
	case "api.ctrt.ContainerRuntime/v0":
		srv := ctxt.KindImplMap[ctxt.AddrKindVerMap[addr]].(api_ctrt_v0.ContainerRuntimeServer)
		api_ctrt_v0.RegisterContainerRuntimeServer(ctxt.AddrGrpcServerMap[addr], srv)
		ctxt.AddrServerMap[addr] = srv
	}
	return nil
}

// Specific kinds follow grpc request calls follow. Functions called by serviceMessageKinds.

func req_ContainerRuntime_v0_ApiServe(req *api_ctrt_v0.ApiServeRequest, ctxt *ApiServiceContext) error {
	if addr, grpcContext, grpcCancel, err := makeClientGrpcContextForMsg("api.ctrt.ContainerRuntime", "v0", req, ctxt); err != nil {
		return err
	} else {
		defer grpcCancel()
		client, ok := ctxt.AddrClientMap[addr].(api_ctrt_v0.ContainerRuntimeClient)
		if !ok {
			return errors.New("no client for " + addr)
		}
		if resp, err := client.ApiServe(grpcContext, req); err != nil {
			return err
		} else if handler := ctxt.RespHandlerMap["api.ctrt.ContainerRuntime/v0.ApiServe"]; handler == nil {
			return nil
		} else if err := handler(resp); err != nil {
			return err
		}
	}
	return nil
}

func req_ContainerRuntime_v0_ApiUnserve(req *api_ctrt_v0.ApiUnserveRequest, ctxt *ApiServiceContext) error {
	if addr, grpcContext, grpcCancel, err := makeClientGrpcContextForMsg("api.ctrt.ContainerRuntime", "v0", req, ctxt); err != nil {
		return err
	} else {
		defer grpcCancel()
		client, ok := ctxt.AddrClientMap[addr].(api_ctrt_v0.ContainerRuntimeClient)
		if !ok {
			return errors.New("no client for " + addr)
		}
		if resp, err := client.ApiUnserve(grpcContext, req); err != nil {
			return err
		} else if handler := ctxt.RespHandlerMap["api.ctrt.ContainerRuntime/v0.ApiUnserve"]; handler == nil {
			return nil
		} else if err := handler(resp); err != nil {
			return err
		}
	}
	return nil
}

func req_ContainerRuntime_v0_List(req *api_ctrt_v0.ListRequest, ctxt *ApiServiceContext) error {
	if addr, grpcContext, grpcCancel, err := makeClientGrpcContextForMsg("api.ctrt.ContainerRuntime", "v0", req, ctxt); err != nil {
		return err
	} else {
		defer grpcCancel()
		client, ok := ctxt.AddrClientMap[addr].(api_ctrt_v0.ContainerRuntimeClient)
		if !ok {
			return errors.New("no client for " + addr)
		}
		if resp, err := client.List(grpcContext, req); err != nil {
			return err
		} else if handler := ctxt.RespHandlerMap["api.ctrt.ContainerRuntime/v0.List"]; handler == nil {
			return nil
		} else if err := handler(resp); err != nil {
			return err
		}
	}
	return nil
}

func req_ContainerRuntime_v0_QueryState(req *api_ctrt_v0.QueryStateRequest, ctxt *ApiServiceContext) error {
	if addr, grpcContext, grpcCancel, err := makeClientGrpcContextForMsg("api.ctrt.ContainerRuntime", "v0", req, ctxt); err != nil {
		return err
	} else {
		defer grpcCancel()
		client, ok := ctxt.AddrClientMap[addr].(api_ctrt_v0.ContainerRuntimeClient)
		if !ok {
			return errors.New("no client for " + addr)
		}
		if resp, err := client.QueryState(grpcContext, req); err != nil {
			return err
		} else if handler := ctxt.RespHandlerMap["api.ctrt.ContainerRuntime/v0.QueryState"]; handler == nil {
			return nil
		} else if err := handler(resp); err != nil {
			return err
		}
	}
	return nil
}

func req_ContainerRuntime_v0_Create(req *api_ctrt_v0.CreateRequest, ctxt *ApiServiceContext) error {
	if addr, grpcContext, grpcCancel, err := makeClientGrpcContextForMsg("api.ctrt.ContainerRuntime", "v0", req, ctxt); err != nil {
		return err
	} else {
		defer grpcCancel()
		client, ok := ctxt.AddrClientMap[addr].(api_ctrt_v0.ContainerRuntimeClient)
		if !ok {
			return errors.New("no client for " + addr)
		}
		if resp, err := client.Create(grpcContext, req); err != nil {
			return err
		} else if handler := ctxt.RespHandlerMap["api.ctrt.ContainerRuntime/v0.Create"]; handler == nil {
			return nil
		} else if err := handler(resp); err != nil {
			return err
		}
	}
	return nil
}

func req_ContainerRuntime_v0_Start(req *api_ctrt_v0.StartRequest, ctxt *ApiServiceContext) error {
	if addr, grpcContext, grpcCancel, err := makeClientGrpcContextForMsg("api.ctrt.ContainerRuntime", "v0", req, ctxt); err != nil {
		return err
	} else {
		defer grpcCancel()
		client, ok := ctxt.AddrClientMap[addr].(api_ctrt_v0.ContainerRuntimeClient)
		if !ok {
			return errors.New("no client for " + addr)
		}
		if resp, err := client.Start(grpcContext, req); err != nil {
			return err
		} else if handler := ctxt.RespHandlerMap["api.ctrt.ContainerRuntime/v0.Start"]; handler == nil {
			return nil
		} else if err := handler(resp); err != nil {
			return err
		}
	}
	return nil
}

func req_ContainerRuntime_v0_Kill(req *api_ctrt_v0.KillRequest, ctxt *ApiServiceContext) error {
	if addr, grpcContext, grpcCancel, err := makeClientGrpcContextForMsg("api.ctrt.ContainerRuntime", "v0", req, ctxt); err != nil {
		return err
	} else {
		defer grpcCancel()
		client, ok := ctxt.AddrClientMap[addr].(api_ctrt_v0.ContainerRuntimeClient)
		if !ok {
			return errors.New("no client for " + addr)
		}
		if resp, err := client.Kill(grpcContext, req); err != nil {
			return err
		} else if handler := ctxt.RespHandlerMap["api.ctrt.ContainerRuntime/v0.Kill"]; handler == nil {
			return nil
		} else if err := handler(resp); err != nil {
			return err
		}
	}
	return nil
}

func req_ContainerRuntime_v0_Delete(req *api_ctrt_v0.DeleteRequest, ctxt *ApiServiceContext) error {
	if addr, grpcContext, grpcCancel, err := makeClientGrpcContextForMsg("api.ctrt.ContainerRuntime", "v0", req, ctxt); err != nil {
		return err
	} else {
		defer grpcCancel()
		client, ok := ctxt.AddrClientMap[addr].(api_ctrt_v0.ContainerRuntimeClient)
		if !ok {
			return errors.New("no client for " + addr)
		}
		if resp, err := client.Delete(grpcContext, req); err != nil {
			return err
		} else if handler := ctxt.RespHandlerMap["api.ctrt.ContainerRuntime/v0.Delete"]; handler == nil {
			return nil
		} else if err := handler(resp); err != nil {
			return err
		}
	}
	return nil
}
