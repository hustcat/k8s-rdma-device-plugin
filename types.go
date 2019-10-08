package main

import (
	"github.com/hustcat/k8s-rdma-device-plugin/ibverbs"
)

type Device struct {
	RdmaDevice ibverbs.IbvDevice
	NetDevice  string
	NumaNode   int64
}
