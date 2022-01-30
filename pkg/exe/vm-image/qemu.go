package main

import (
	api_os_machine_image_v0 "alt-os/api/os/machine/image/v0"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// qemuCreateImage creates a virtual image directory for the specified vm in the
// given root directory.
func qemuCreateImage(def *api_os_machine_image_v0.VirtualMachine, rootDir string) error {

	absImageDir, _ := filepath.Abs(rootDir)
	absImageDir = filepath.Clean(filepath.Join(absImageDir, def.ImageDir))
	imageRootName := filepath.Join(absImageDir, "root")
	imageBootName := filepath.Join(imageRootName, "EFI", "BOOT")
	vmDefName := filepath.Join(absImageDir, "vm-def.json")
	biosCodeName := filepath.Join(absImageDir, "bios-code.fd")
	biosVarsName := filepath.Join(absImageDir, "bios-vars.fd")
	os.RemoveAll(imageRootName)
	os.MkdirAll(imageBootName, 0755)
	confBootName := filepath.Join(imageBootName, "SYS.CONF")
	imageBootName = filepath.Join(imageBootName, filepath.Base(def.EfiPath))

	debugPort := uint32(0)
	debugAddress := uint32(0)
	for _, serial := range def.Serial {
		if serial.Type == api_os_machine_image_v0.SerialType_SERIAL_STDOUT {
			debugPort = serial.Port
			debugAddress = serial.Address
			break
		}
	}

	if data, err := ioutil.ReadFile(def.EfiPath); err != nil {
		return err
	} else if err := ioutil.WriteFile(imageBootName, data, 0755); err != nil {
		return err
	}
	if debugPort != 0 {
		if f, err := os.Create(confBootName); err != nil {
			return err
		} else {
			f.WriteString(fmt.Sprintf("DebugPort=0x%04X\n", debugPort))
			f.Close()
		}
	} else if debugAddress != 0 {
		if f, err := os.Create(confBootName); err != nil {
			return err
		} else {
			f.WriteString(fmt.Sprintf("DebugAddress=0x%08X\n", debugAddress))
			f.Close()
		}
	}
	if data, err := ioutil.ReadFile(def.BiosImage); err != nil {
		return err
	} else if err := ioutil.WriteFile(biosCodeName, data, 0755); err != nil {
		return err
	}
	if data, err := ioutil.ReadFile(def.VarsImage); err != nil {
		return err
	} else if err := ioutil.WriteFile(biosVarsName, data, 0755); err != nil {
		return err
	}
	if f, err := os.Create(vmDefName); err != nil {
		return err
	} else {
		encoder := json.NewEncoder(f)
		err := encoder.Encode(def)
		f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
