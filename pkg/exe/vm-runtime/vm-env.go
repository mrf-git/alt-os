package main

import (
	api_os_machine_runtime_v0 "alt-os/api/os/machine/runtime/v0"
	"alt-os/exe"
	"alt-os/os/code"
	"alt-os/os/limits"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"strings"
	"time"
)

// VmEnvironment encapsulates the runtime environment for a single
// virtual machine to run virtualized code.
type VmEnvironment interface {
	// Run creates a hardware-virtualized, privilege level 0
	// environment in a new goroutine and starts running the virtual
	// machine code in it. Sends the code returned by main to
	// returnCodeCh when the virtual machine exits.
	Run(signalCh <-chan int, returnCodeCh chan<- int) error
}

// newVmEnvironment returns a newly-instantiated VmEnvironment for the
// specified code to run in.
func newVmEnvironment(exeCode code.ExecutableCode, ctxt *VmRuntimeContext) VmEnvironment {
	return &_VmEnvironment{
		logger:  exe.NewLogger(ctxt.ExeLoggerConf),
		ctxt:    ctxt,
		exeCode: exeCode,
	}
}

type _VmEnvironment struct {
	logger       exe.Logger
	ctxt         *VmRuntimeContext
	exeCode      code.ExecutableCode
	signalCh     <-chan int
	returnCodeCh chan<- int
}

func (vmEnv *_VmEnvironment) Run(signalCh <-chan int, returnCodeCh chan<- int) error {
	vmEnv.signalCh = signalCh
	vmEnv.returnCodeCh = returnCodeCh
	go runVm(vmEnv)
	return nil
}

// runVm initializes and runs the virtual machine environment to completion.
func runVm(vmEnv *_VmEnvironment) {

	controlCh := make(chan QmpControlCommandType, limits.MAX_PROCESS_SIGNALS)

	go func() {
		for {
			intSig, ok := <-vmEnv.signalCh
			if !ok {
				goto stopHandlingKillSignals
			}
			sig := api_os_machine_runtime_v0.KillSignal(intSig)
			switch sig {
			default:
				vmEnv.logger.WithFields(exe.Fields{
					"signal": sig,
				}).Warn("unrecognized signal")
			case api_os_machine_runtime_v0.KillSignal_SIGTERM:
				controlCh <- _QMP_CONTROL_SHUTDOWN
				goto stopHandlingKillSignals
			}
		}
	stopHandlingKillSignals:
	}()

	var decoder *json.Decoder
	resumeEvent := &QmpEvent{}
	initEvent := &QmpInit{}
	outBuff := bytes.NewBuffer(nil)
	errBuff := bytes.NewBuffer(nil)
	inBuff := bytes.NewBuffer(nil)

	// QEMU settings based on qemu wiki: https://wiki.qemu.org/Features/VT-d
	args := []string{"-display", "none", "-cpu", "host", "-enable-kvm",
		"-machine", "q35,kernel-irqchip=split",
		"-device", "intel-iommu,intremap=on,caching-mode=on,device-iotlb=on",
		"-netdev", "user,id=net0", "-device", "ioh3420,id=pcie.1,chassis=1",
		"-device", "virtio-net-pci,bus=pcie.1,netdev=net0,disable-legacy=on," +
			"disable-modern=off,iommu_platform=on,ats=on",
		"-chardev", "stdio,mux=on,id=charctl", "-mon", "charctl,mode=control"}

	var serviceParams *qmpServiceParams
	args = append(args, "-drive", "format=raw,file=workspace/alpine.iso")
	cmd := exec.Command("qemu-system-x86_64", args...)
	cmd.Stdout = outBuff
	cmd.Stdin = inBuff
	cmd.Stderr = errBuff
	go func() {
		if err := cmd.Run(); err != nil {
			vmEnv.returnCodeCh <- -1
		} else {
			vmEnv.returnCodeCh <- 0
		}
	}()

	// Read the initialization event from qemu.
	resumeEventStr := ""
	initEventStr := ""
	readResume := false
	readInit := false
	for {
		if !readResume && outBuff.Len() > 0 {
			if str, err := outBuff.ReadString('\n'); err != nil && !errors.Is(err, io.EOF) {
				vmEnv.logger.WithFields(exe.Fields{
					"err": err.Error(),
				}).Error("failed to read resume")
				goto killVm
			} else {
				resumeEventStr += str
				if ind := strings.Index(resumeEventStr, "\n"); ind >= 0 {
					initEventStr += string(resumeEventStr[ind+1:])
					resumeEventStr = string(resumeEventStr[:ind])
					readResume = true
				}
			}
		}
		if readResume && !readInit && outBuff.Len() > 0 {
			if str, err := outBuff.ReadString('\n'); err != nil && !errors.Is(err, io.EOF) {
				vmEnv.logger.WithFields(exe.Fields{
					"err": err.Error(),
				}).Error("failed to read init")
				goto killVm
			} else {
				initEventStr += str
				if ind := strings.Index(initEventStr, "\n"); ind >= 0 {
					readInit = true
				}
			}
		}
		if readResume && readInit {
			break
		}
		if errBuff.Len() > 0 {
			if str, err := errBuff.ReadString('\n'); err == nil || errors.Is(err, io.EOF) {
				str = strings.Trim(resumeEventStr, " \x00\n")
				if str != "" {
					vmEnv.logger.Error(str)
				}
			}
		}
	}
	resumeEventStr = strings.Trim(resumeEventStr, " \x00\n")
	initEventStr = strings.Trim(initEventStr, " \x00\n")

	// Decode the resume and init.
	decoder = json.NewDecoder(strings.NewReader(resumeEventStr))
	if err := decoder.Decode(resumeEvent); err != nil {
		vmEnv.logger.WithFields(exe.Fields{
			"err": err.Error(),
		}).Error("failed to decode resume")
		goto killVm
	}
	decoder = json.NewDecoder(strings.NewReader(initEventStr))
	if err := decoder.Decode(initEvent); err != nil {
		vmEnv.logger.WithFields(exe.Fields{
			"err": err.Error(),
		}).Error("failed to decode init")
		goto killVm
	}
	if resumeEvent.Event != "RESUME" {
		vmEnv.logger.WithFields(exe.Fields{
			"event": resumeEvent.Event,
		}).Error("not a resume event")
		goto killVm
	}

	// Service the VM.
	serviceParams = &qmpServiceParams{
		encoder:    json.NewEncoder(inBuff),
		decoder:    json.NewDecoder(outBuff),
		controlCh:  controlCh,
		vmEnv:      vmEnv,
		resumeTime: time.Unix(int64(resumeEvent.Timestamp.Seconds), 0),
		verMajor:   initEvent.Qmp.Version.Qemu.Major,
		verMinor:   initEvent.Qmp.Version.Qemu.Minor,
		verMicro:   initEvent.Qmp.Version.Qemu.Micro,
	}
	serviceParams.resumeTime.Add(time.Duration(resumeEvent.Timestamp.Microseconds) * time.Microsecond)
	serviceParams.resumeTime = serviceParams.resumeTime.UTC()
	if err := qmpService(serviceParams); err != nil {
		vmEnv.logger.WithFields(exe.Fields{
			"err": err.Error(),
		}).Error("QMP servicing error")
		goto killVm
	}

killVm:
	close(controlCh)
	cmd.Process.Kill()
}
