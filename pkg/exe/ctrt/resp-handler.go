package main

import (
	api_os_container_runtime_v0 "alt-os/api/os/container/runtime/v0"
	"fmt"
)

// handleRespList prints the List response.
func handleRespList(resp *api_os_container_runtime_v0.ListResponse) error {

	fmt.Println("got list response")

	return nil
}
