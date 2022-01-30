package main

import (
	api_os_machine_image_v0 "alt-os/api/os/machine/image/v0"
	api_os_machine_runtime_v0 "alt-os/api/os/machine/runtime/v0"
	"alt-os/exe"
	"alt-os/os/limits"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// newVmEnvironment returns a newly-instantiated VmEnvironment.
func newVmEnvironment(imagePath string, ctxt *VmRuntimeContext) VmEnvironment {
	return &_VmEnvironment{
		logger:    exe.NewLogger(ctxt.ExeLoggerConf),
		ctxt:      ctxt,
		imagePath: imagePath,
	}
}

type _VmEnvironment struct {
	logger       exe.Logger
	ctxt         *VmRuntimeContext
	vmDef        *api_os_machine_image_v0.VirtualMachine
	imagePath    string
	signalCh     <-chan int
	returnCodeCh chan<- int
}

func (vmEnv *_VmEnvironment) Run(signalCh <-chan int, returnCodeCh chan<- int) error {
	vmEnv.signalCh = signalCh
	vmEnv.returnCodeCh = returnCodeCh
	go runVm(vmEnv)
	return nil
}

type _TrimmingReader struct {
	r io.Reader
}

func (r _TrimmingReader) Read(data []byte) (int, error) {
	n, err := r.r.Read(data)
	if err != nil {
		return n, err
	}
	outData := bytes.Trim(data, " \x00\n")
	n = copy(data[:len(outData)], outData)
	return n, nil
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

	// Prepare input/output and control mechanisms.
	var decoder *json.Decoder
	resumeEvent := &QmpEvent{}
	initEvent := &QmpInit{}
	outBuff := bytes.NewBuffer(nil)
	errBuff := bytes.NewBuffer(nil)
	inBuff := bytes.NewBuffer(nil)

	absImageDir, _ := filepath.Abs(vmEnv.imagePath)
	absImageDir = filepath.Clean(absImageDir)
	vmDefName := filepath.Join(absImageDir, "vm-def.json")
	biosCodeName := filepath.Join(absImageDir, "bios-code.fd")
	biosVarsName := filepath.Join(absImageDir, "bios-vars.fd")
	imageRootName := filepath.Join(absImageDir, "root")

	// Load the serialized vm definition.
	vmEnv.vmDef = &api_os_machine_image_v0.VirtualMachine{}
	if f, err := os.Open(vmDefName); err != nil {
		vmEnv.logger.WithFields(exe.Fields{
			"err": err.Error(),
		}).Error("failed to open vm def")
		close(controlCh)
		return
	} else {
		decoder := json.NewDecoder(f)
		err := decoder.Decode(vmEnv.vmDef)
		f.Close()
		if err != nil {
			vmEnv.logger.WithFields(exe.Fields{
				"err": err.Error(),
			}).Error("failed to decode vm def")
			close(controlCh)
			return
		}
	}
	memoryMib := vmEnv.vmDef.Memory >> 20
	vmEnv.logger.WithFields(exe.Fields{
		"image-dir":  vmEnv.vmDef.ImageDir,
		"processors": vmEnv.vmDef.Processors,
		"memory-mib": memoryMib,
	}).Info("Loaded vm definition")

	sockNames := [...]string{"com1.sock", "com2.sock", "com3.sock", "com4.sock"}
	for i, name := range sockNames {
		sockNames[i] = filepath.Join(absImageDir, name)
		os.RemoveAll(sockNames[i])
	}
	os.MkdirAll(absImageDir, 0755)

	ioParams := &ioServiceParams{
		vmEnv: vmEnv,
	}
	for i, name := range sockNames {
		if sock, err := net.Listen("unix", name); err == nil {
			defer sock.Close()
			defer os.RemoveAll(name)
			ioParams.comSocks[i] = sock
		}
	}

	// QEMU settings based on qemu wiki: https://wiki.qemu.org/Features/VT-d
	args := []string{"-display", "none", "-m", fmt.Sprintf("%d", memoryMib),
		"-smp", fmt.Sprintf("%d", vmEnv.vmDef.Processors),
		"-chardev", "stdio,mux=on,id=charctl", "-mon", "charctl,mode=control",
		"-chardev", "socket,mux=on,id=charcom1,path=" + sockNames[0],
		"-chardev", "socket,mux=on,id=charcom2,path=" + sockNames[1],
		"-chardev", "socket,mux=on,id=charcom3,path=" + sockNames[2],
		"-chardev", "socket,mux=on,id=charcom4,path=" + sockNames[3],
		"-serial", "chardev:charcom1",
		"-serial", "chardev:charcom2",
		"-serial", "chardev:charcom3",
		"-serial", "chardev:charcom4",
		"-device", "virtio-blk-pci,drive=bootdisk,bootindex=0",
	}

	qemuCmd := ""
	switch vmEnv.vmDef.ArchType {
	case api_os_machine_image_v0.ArchType_ARCH_AMD64:
		qemuCmd = "qemu-system-x86_64"
		args = append(args, "-netdev", "user,id=net0")
		args = append(args, "-device", "intel-iommu,intremap=on,caching-mode=on,device-iotlb=on")
		args = append(args, "-device", "ioh3420,id=pcie.1,chassis=1")
		args = append(args, "-device", "virtio-net-pci,bus=pcie.1,netdev=net0,disable-legacy=on,"+
			"disable-modern=off,iommu_platform=on,ats=on")
		args = append(args, "-machine", "q35,kernel-irqchip=split")
		if runtime.GOARCH == "amd64" {
			args = append(args, "-cpu", "host", "-enable-kvm")
		} else {
			args = append(args, "-cpu", "max")
		}
	case api_os_machine_image_v0.ArchType_ARCH_AARCH64:
		qemuCmd = "qemu-system-aarch64"
		args = append(args, "-machine", "virt")
		if runtime.GOARCH == "aarch64" {
			args = append(args, "-cpu", "host", "-enable-kvm")
		} else {
			args = append(args, "-cpu", "max")
		}
	}

	args = append(args,
		"-drive", "format=raw,if=pflash,unit=0,readonly=on,file="+biosCodeName,
		"-drive", "format=raw,if=pflash,unit=1,file="+biosVarsName,
	)

	var qmpParams *qmpServiceParams
	args = append(args, "-drive", "format=raw,unit=2,if=none,id=bootdisk,file=fat:rw:"+imageRootName)
	cmd := exec.Command(qemuCmd, args...)
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
	// vmEnv.logger.Info(strings.Join(cmd.Args, " "))

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
				str = strings.Trim(str, " \x00\n")
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

	// Service the VM IO in another goroutine.
	go ioService(ioParams)

	// Service the VM QMP messages.
	qmpParams = &qmpServiceParams{
		encoder:    json.NewEncoder(inBuff),
		decoder:    json.NewDecoder(_TrimmingReader{r: outBuff}),
		controlCh:  controlCh,
		vmEnv:      vmEnv,
		resumeTime: time.Unix(int64(resumeEvent.Timestamp.Seconds), 0),
		verMajor:   initEvent.Qmp.Version.Qemu.Major,
		verMinor:   initEvent.Qmp.Version.Qemu.Minor,
		verMicro:   initEvent.Qmp.Version.Qemu.Micro,
	}
	qmpParams.resumeTime.Add(time.Duration(resumeEvent.Timestamp.Microseconds) * time.Microsecond)
	qmpParams.resumeTime = qmpParams.resumeTime.UTC()
	if err := qmpService(qmpParams); err != nil {
		vmEnv.logger.WithFields(exe.Fields{
			"err": err.Error(),
		}).Error("QMP servicing error")
		goto killVm
	}

killVm:
	close(controlCh)
	cmd.Process.Kill()
}
