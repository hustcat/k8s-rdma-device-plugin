package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const VfNetDevicePath = "/sys/class/net/%s/device/virtfn%d/net"
const SriovFile = "/sys/class/net/%s/device/sriov_numvfs"

func GetVfNetDevice(master string) ([]string, error) {
	var netDeviceList []string

	sriovFile := fmt.Sprintf(SriovFile, master)
	if _, err := os.Lstat(sriovFile); err != nil {
		return nil, fmt.Errorf("failed to open the sriov_numfs of device %q: %v", master, err)
	}

	data, err := ioutil.ReadFile(sriovFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read the sriov_numfs of device %q: %v", master, err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no data in the file %q", sriovFile)
	}

	sriovNumfs := strings.TrimSpace(string(data))
	vfTotal, err := strconv.Atoi(sriovNumfs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sriov_numfs(byte value) to int of device %q: %v", master, err)
	}

	if vfTotal <= 0 {
		return nil, fmt.Errorf("no virtual function in the device %q: %v", master)
	}

	for vf := 0; vf < vfTotal; vf++ {
		devName, err := getVFDeviceName(master, vf)
		if err != nil {
			return netDeviceList, err
		}
		netDeviceList = append(netDeviceList, devName)
	}

	return netDeviceList, nil
}

func GetAllNetDevice() ([]string, error) {
	// TODO:
	return nil, nil
}

func getVFDeviceName(master string, vf int) (string, error) {
	vfDir := fmt.Sprintf(VfNetDevicePath, master, vf)
	if _, err := os.Lstat(vfDir); err != nil {
		return "", fmt.Errorf("failed to open the virtfn%d dir of the device %q: %v", vf, master, err)
	}

	infos, err := ioutil.ReadDir(vfDir)
	if err != nil {
		return "", fmt.Errorf("failed to read the virtfn%d dir of the device %q: %v", vf, master, err)
	}

	if len(infos) != 1 {
		return "", fmt.Errorf("no network device in directory %s", vfDir)
	}
	return infos[0].Name(), nil
}
