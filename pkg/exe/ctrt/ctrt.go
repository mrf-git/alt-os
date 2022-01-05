package main

import (
	api_ctrt_v0 "alt-os/api/ctrt/v0"
	"alt-os/exe"
	"errors"
	"fmt"
)

// handleApiMessages processes each message in order and returns only when all
// tasks have stopped. Therefore, if a message starts hosting a new runtime
// without a later message exiting it, the function will not return until the
// runtime exits by some other means.
func handleApiMessages(ctxt *CtrtContext) {
	for _, apiMsg := range ctxt.Messages {
		fmt.Println(apiMsg.Kind)
		fmt.Println(apiMsg.Version)

		switch msg := apiMsg.Def.(type) {
		default:
			exe.Fatal("handling api messages", errors.New("invalid message kind: "+apiMsg.Kind), ctxt.ExeContext)

		case *api_ctrt_v0.HostConfiguration:
			fmt.Println(msg.GoString())
		}

		// TODO
	}
}
