package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"golang.org/x/net/context"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1alpha1"

	"github.com/hustcat/k8s-rdma-device-plugin/ibverbs"
)

const RdmaDeviceRource = "/sys/class/infiniband/%s/device/resource"
const NetDeviceRource = "/sys/class/net/%s/device/resource"

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

			// the same device
			if bytes.Compare(dResource, nResource) == 0 {
				devs = append(devs, Device{
					RdmaDevice: d,
					NetDevice:  n,
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

			// the same device
			if bytes.Compare(dResource, nResource) == 0 {
				devs = append(devs, Device{
					RdmaDevice: d,
					NetDevice:  n,
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
