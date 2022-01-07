package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ApiServiceMessage interface represents the common values of all service messages.
type ApiServiceMessage interface {
	GetHostname() string
	GetPort() int32
	GetTimeout() int32
}

// ApiServiceContext holds runtime context information for OS API services.
type ApiServiceContext struct {
	MessageQueue      []*ApiProtoMessage
	AddrGrpcServerMap map[string]*grpc.Server
	AddrGrpcConnMap   map[string]*grpc.ClientConn
	AddrKindVerMap    map[string]string
	AddrClientMap     map[string]interface{}
	AddrServerMap     map[string]interface{}
	AddrStopSignalMap map[string]func()
	AddrStopWaitMap   map[string]func() error
	RespHandlerMap    map[string]func(interface{}) error
	KindImplMap       map[string]interface{}
	ServerWg          *sync.WaitGroup
}

// InitContext initializes a new context for api services.
func InitContext(kindImplMap map[string]interface{}, respHandlerMap map[string]func(interface{}) error) *ApiServiceContext {
	return &ApiServiceContext{
		AddrGrpcServerMap: make(map[string]*grpc.Server),
		AddrGrpcConnMap:   make(map[string]*grpc.ClientConn),
		AddrKindVerMap:    make(map[string]string),
		AddrClientMap:     make(map[string]interface{}),
		AddrServerMap:     make(map[string]interface{}),
		AddrStopSignalMap: make(map[string]func()),
		AddrStopWaitMap:   make(map[string]func() error),
		RespHandlerMap:    respHandlerMap,
		KindImplMap:       kindImplMap,
		ServerWg:          &sync.WaitGroup{},
	}
}

// ServiceMessages processes each message in the queue and returns only when all
// tasks have stopped. Therefore, if a message starts serving a new server
// without a later message unserving it, the function will not return until the
// server stops by some other means.
func ServiceMessages(ctxt *ApiServiceContext) error {
	return serviceMessageKinds(ctxt)
}

// makeClientGrpcContextForMsg creates a new client for the api message if one doesn't
// already exist, then initializes a new grpc context for a grpc call.
func makeClientGrpcContextForMsg(kind, version string, msg ApiServiceMessage,
	ctxt *ApiServiceContext) (string, context.Context, context.CancelFunc, error) {

	addr := fmt.Sprintf("%s:%d", msg.GetHostname(), msg.GetPort())
	if _, ok := ctxt.AddrClientMap[addr]; !ok {
		if err := newClient(kind, version, addr, ctxt); err != nil {
			return "", nil, nil, err
		}
	}
	grpcContext, grpcCancel := context.WithTimeout(context.Background(),
		time.Duration(msg.GetTimeout())*time.Second)
	return addr, grpcContext, grpcCancel, nil
}

// newServer creates a new gRPC server and stores it in the
// context map with listening address as the key to the map.
func newServer(kind, version string, msg ApiServiceMessage, ctxt *ApiServiceContext) error {
	addr := fmt.Sprintf("%s:%d", msg.GetHostname(), msg.GetPort())
	chStop := make(chan bool, 1)
	if _, ok := ctxt.AddrServerMap[addr]; ok {
		return errors.New("already exists: " + addr)
	}
	if listener, err := net.Listen("tcp", addr); err != nil {
		return err
	} else {
		ctxt.AddrGrpcServerMap[addr] = grpc.NewServer()
		ctxt.AddrKindVerMap[addr] = kind + "/" + version
		ctxt.AddrStopSignalMap[addr] = func() { chStop <- true }
		if err := makeServerKind(addr, ctxt); err != nil {
			return err
		}
		ctxt.ServerWg.Add(1)
		go func() {
			if err := ctxt.AddrGrpcServerMap[addr].Serve(listener); err != nil {
				fmt.Println("Error serving " + addr + ": " + err.Error())
			}
			ctxt.ServerWg.Done()
		}()
		ctxt.AddrStopWaitMap[addr] = func() error {
			stop, ok := <-chStop
			if stop && ok {
				if err := closeClient(addr, ctxt); err != nil {
					return err
				}
				if err := stopServer(addr, ctxt); err != nil {
					return err
				}
			}
			return nil
		}
	}
	return nil
}

// newClient creates a new gRPC client for the specified
// address and stores it in the context map.
func newClient(kind, version, addr string, ctxt *ApiServiceContext) error {
	if _, ok := ctxt.AddrClientMap[addr]; ok {
		return errors.New("already exists: " + addr)
	}
	creds := insecure.NewCredentials() // No TLS, localhost assumed.
	if conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds)); err != nil {
		return err
	} else {
		if kindVer, ok := ctxt.AddrKindVerMap[addr]; ok && kindVer != kind+"/"+version {
			return errors.New("client kind mismatch")
		}
		ctxt.AddrGrpcConnMap[addr] = conn
		ctxt.AddrKindVerMap[addr] = kind + "/" + version
		if err := makeClientKind(addr, ctxt); err != nil {
			return err
		}
	}
	return nil
}

// closeClient closes the client for the specified address and
// removes it from the context.
func closeClient(addr string, ctxt *ApiServiceContext) error {
	if _, ok := ctxt.AddrClientMap[addr]; !ok {
		return nil
	}
	if err := ctxt.AddrGrpcConnMap[addr].Close(); err != nil {
		return err
	}
	delete(ctxt.AddrGrpcConnMap, addr)
	delete(ctxt.AddrClientMap, addr)
	return nil
}

// stopServer stops the server at the specified address from listening and
// removes it from the context.
func stopServer(addr string, ctxt *ApiServiceContext) error {
	if _, ok := ctxt.AddrServerMap[addr]; !ok {
		return nil
	}
	ctxt.AddrGrpcServerMap[addr].GracefulStop()
	delete(ctxt.AddrServerMap, addr)
	delete(ctxt.AddrGrpcServerMap, addr)
	return nil
}
