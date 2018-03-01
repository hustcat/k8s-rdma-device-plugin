// +build linux

package ibverbs

// #cgo LDFLAGS: -libverbs
// #include <stdlib.h>
// #include <infiniband/verbs.h>
import "C"

import (
	"fmt"
	"unsafe"
)

func IbvGetDeviceList() ([]IbvDevice, error) {
	var ibvDevList []IbvDevice
	var c_num C.int
	var c_ptrdevice *C.struct_ibv_device

	c_devList := C.ibv_get_device_list(&c_num)

	if c_devList == nil {
		return nil, fmt.Errorf("failed to get IB devices list")
	}

	ptrSize := unsafe.Sizeof(c_ptrdevice)
	ptr := uintptr(unsafe.Pointer(c_devList))
	for i := 0; i <= int(c_num); i++ {
		c_ptrdevice = *(**C.struct_ibv_device)(unsafe.Pointer(ptr))
		if c_ptrdevice == nil {
			break
		}

		c_device := *c_ptrdevice
		device := IbvDevice{
			Name:       C.GoString(&c_device.name[0]),
			DevName:    C.GoString(&c_device.dev_name[0]),
			DevPath:    C.GoString(&c_device.dev_path[0]),
			IbvDevPath: C.GoString(&c_device.ibdev_path[0]),
		}
		ibvDevList = append(ibvDevList, device)
		ptr += ptrSize
	}

	C.ibv_free_device_list(c_devList)
	return ibvDevList, nil
}
