package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/hustcat/k8s-rdma-device-plugin/ibverbs"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1alpha"
)

const (
	RdmaDeviceRource   = "/sys/class/infiniband/%s/device/resource"
	NetDeviceRource    = "/sys/class/net/%s/device/resource"
	RdmaDeviceNumaNode = "/sys/class/infiniband/%s/device/numa_node"
)

func GetDevices(masterNetDevice string) ([]Device, error) {
	if masterNetDevice == "" {
		return getAllRdmaDeivces()
	} else {
		return getRdmaDeivces(masterNetDevice)
	}
}

func getAllRdmaDeivces() ([]Device, error) {
	var devs []Device
	// Get all RDMA device list
	ibvDevList, err := ibverbs.IbvGetDeviceList()
	if err != nil {
		return nil, err
	}

	netDevList, err := GetAllNetDevice()
	if err != nil {
		return nil, err
	}
	for _, d := range ibvDevList {
		for _, n := range netDevList {
			dResource, err := getRdmaDeviceResoure(d.Name)
			if err != nil {
				return nil, err
			}
			nResource, err := getNetDeviceResoure(n)
			if err != nil {
				return nil, err
			}
			nn, err := getRdmaDeviceNumaNode(d.Name)
			if err != nil {
				return nil, err
			}

			// the same device
			if bytes.Compare(dResource, nResource) == 0 {
				devs = append(devs, Device{
					RdmaDevice: d,
					NetDevice:  n,
					NumaNode:   int64(nn),
				})
			}
		}
	}
	return devs, nil
}

func getRdmaDeivces(masterNetDevice string) ([]Device, error) {
	var devs []Device
	// Get all RDMA device list
	ibvDevList, err := ibverbs.IbvGetDeviceList()
	if err != nil {
		return nil, err
	}

	netDevList, err := GetVfNetDevice(masterNetDevice)
	if err != nil {
		return nil, err
	}

	for _, d := range ibvDevList {
		for _, n := range netDevList {
			dResource, err := getRdmaDeviceResoure(d.Name)
			if err != nil {
				return nil, err
			}
			nResource, err := getNetDeviceResoure(n)
			if err != nil {
				return nil, err
			}
			nn, err := getRdmaDeviceNumaNode(d.Name)
			if err != nil {
				return nil, err
			}

			// the same device
			if bytes.Compare(dResource, nResource) == 0 {
				devs = append(devs, Device{
					RdmaDevice: d,
					NetDevice:  n,
					NumaNode:   int64(nn),
				})
			}
		}
	}
	return devs, nil
}

func getRdmaDeviceResoure(name string) ([]byte, error) {
	resourceFile := fmt.Sprintf(RdmaDeviceRource, name)
	data, err := ioutil.ReadFile(resourceFile)
	return data, err
}

func getNetDeviceResoure(name string) ([]byte, error) {
	resourceFile := fmt.Sprintf(NetDeviceRource, name)
	data, err := ioutil.ReadFile(resourceFile)
	return data, err
}

func getRdmaDeviceNumaNode(name string) (int, error) {
	numaNodeFile := fmt.Sprintf(RdmaDeviceNumaNode, name)
	data, err := ioutil.ReadFile(numaNodeFile)
	return strconv.Atoi(string(data))
}

func deviceExists(devs []*pluginapi.Device, id string) bool {
	for _, d := range devs {
		if d.ID == id {
			return true
		}
	}
	return false
}

func watchXIDs(ctx context.Context, devs []*pluginapi.Device, xids chan<- *pluginapi.Device) {
	for {
		select {
		case <-ctx.Done():
			return
		}

		// TODO: check RDMA device healthy status
	}
}
