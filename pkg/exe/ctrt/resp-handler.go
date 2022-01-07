package main

import (
	api_ctrt_v0 "alt-os/api/ctrt/v0"
	"fmt"
)

// handleRespList prints the List response.
func handleRespList(resp *api_ctrt_v0.ListResponse) error {

	fmt.Println("got list response")

	return nil
}
