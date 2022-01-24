package main

import (
	"alt-os/exe"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

// QmpCommand represents a qmp command message.
type QmpCommand struct {
	Id        int         `json:"id"`
	Execute   string      `json:"execute"`
	Arguments interface{} `json:"arguments,omitempty"`
}

// QmpResponse represents a qmp response message.
type QmpResponse struct {
	Id     int         `json:"id"`
	Return interface{} `json:"return,omitempty"`
	Error  struct {
		Class string `json:"class"`
		Desc  string `json:"desc"`
	} `json:"error,omitempty"`
}

// QmpEvent represents a qmp event message.
type QmpEvent struct {
	Event     string `json:"event"`
	Timestamp struct {
		Seconds      int `json:"seconds"`
		Microseconds int `json:"microseconds"`
	} `json:"timestamp"`
}

// QmpInit represents a qmp init message.
type QmpInit struct {
	Qmp struct {
		Capabilities []string `json:"capabilities"`
		Version      struct {
			Package string `json:"package"`
			Qemu    struct {
				Major int `json:"major"`
				Minor int `json:"minor"`
				Micro int `json:"micro"`
			} `json:"qemu"`
		} `json:"version"`
	} `json:"QMP"`
}

// qmpServiceParams holds parameters for the qmpService method.
type qmpServiceParams struct {
	verMajor   int
	verMinor   int
	verMicro   int
	resumeTime time.Time
	encoder    *json.Encoder
	decoder    *json.Decoder
	controlCh  <-chan QmpControlCommandType
	vmEnv      *_VmEnvironment
}

// QmpControlCommandType represents the type of control command sent to a vm.
type QmpControlCommandType int

const (
	_ QmpControlCommandType = iota
	// Completely shut down the virtual machine.
	_QMP_CONTROL_SHUTDOWN
)

// qmpService services QMP messages read from the specified json decoder and
// issues commands via the specified encoder.
func qmpService(params *qmpServiceParams) error {
	logger := params.vmEnv.logger

	errCommand := errors.New("error servicing QMP command")

	do := func(command *QmpCommand) *QmpResponse {
		params.encoder.Encode(command)
		response := &QmpResponse{}
		for {
			if err := params.decoder.Decode(response); err != nil && !errors.Is(err, io.EOF) {
				logger.Error(err.Error())
				return nil
			} else {
				break
			}
		}
		return response
	}

	var response *QmpResponse
	command := &QmpCommand{}

	// Negotiate capabilities (none needed).
	command.Id = 0
	command.Execute = "qmp_capabilities"
	response = do(command)
	if response.Error.Class != "" {
		return errCommand
	}

	logger.WithFields(exe.Fields{
		"qemu-version": fmt.Sprintf("%d.%d.%d", params.verMajor,
			params.verMinor, params.verMicro),
		"resume-time": params.resumeTime,
	}).Info("Starting QMP servicing")

	for {
		control, ok := <-params.controlCh
		if !ok {
			break
		}

		_ = control
		// TODO handle control message
	}

	return nil
}
