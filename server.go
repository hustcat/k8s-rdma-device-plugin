package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"google.golang.org/grpc"
	log "k8s.io/klog"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1alpha"
)

const (
	defaultResourceName = "tencent.com/rdma"
	serverSock          = pluginapi.DevicePluginPath + "rdma.sock"
)

// RdmaDevicePlugin implements the Kubernetes device plugin API
type RdmaDevicePlugin struct {
	devs []*pluginapi.Device
	// ID => Device
	devices         map[string]Device
	socket          string
	masterNetDevice string

	stop   chan interface{}
	health chan *pluginapi.Device

	server *grpc.Server
}

// NewRdmaDevicePlugin returns an initialized RdmaDevicePlugin
func NewRdmaDevicePlugin(master string) *RdmaDevicePlugin {
	devices, err := GetDevices(master)
	if err != nil {
		log.Errorf("Error to get RDMA devices: %v", err)
		return nil
	}

	var devs []*pluginapi.Device
	devMap := make(map[string]Device)
	for _, device := range devices {
		id := device.RdmaDevice.Name
		devs = append(devs, &pluginapi.Device{
			ID:     id,
			Health: pluginapi.Healthy,
			Topology: &pluginapi.TopologyInfo{
				Nodes: []*pluginapi.NUMANode{
					&pluginapi.NUMANode{
						ID: device.NumaNode,
					},
				},
			},
		})
		devMap[id] = device
	}

	return &RdmaDevicePlugin{
		masterNetDevice: master,
		socket:          serverSock,
		devs:            devs,
		devices:         devMap,
		stop:            make(chan interface{}),
		health:          make(chan *pluginapi.Device),
	}
}

// dial establishes the gRPC communication with the registered device plugin.
func dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}

	return c, nil
}

// Start starts the gRPC server of the device plugin
func (m *RdmaDevicePlugin) Start() error {
	err := m.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", m.socket)
	if err != nil {
		return err
	}

	m.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(m.server, m)

	go m.server.Serve(sock)

	// Wait for server to start by launching a blocking connection
	conn, err := dial(m.socket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()

	go m.healthcheck()

	return nil
}

// Stop stops the gRPC server
func (m *RdmaDevicePlugin) Stop() error {
	if m.server == nil {
		return nil
	}

	m.server.Stop()
	m.server = nil
	close(m.stop)

	return m.cleanup()
}

// Register registers the device plugin for the given resourceName with Kubelet.
func (m *RdmaDevicePlugin) Register(kubeletEndpoint, resourceName string) error {
	conn, err := dial(kubeletEndpoint, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(m.socket),
		ResourceName: resourceName,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}
	return nil
}

// ListAndWatch lists devices and update that list according to the health status
func (m *RdmaDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	s.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})

	for {
		select {
		case <-m.stop:
			return nil
		case d := <-m.health:
			// FIXME: there is no way to recover from the Unhealthy state.
			d.Health = pluginapi.Unhealthy
			s.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})
		}
	}
}

func (m *RdmaDevicePlugin) unhealthy(dev *pluginapi.Device) {
	m.health <- dev
}

// Allocate which return list of devices.
func (m *RdmaDevicePlugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	devs := m.devs
	response := pluginapi.AllocateResponse{}

	log.V(1).Infof("Request IDs: %v", r.DevicesIDs)
	var devicesList []*pluginapi.DeviceSpec
	for _, id := range r.DevicesIDs {
		if !deviceExists(devs, id) {
			return nil, fmt.Errorf("invalid allocation request: unknown device: %s", id)
		}

		var devPath string
		if dev, ok := m.devices[id]; ok {
			// TODO: to function
			devPath = fmt.Sprintf("/dev/infiniband/%s", dev.RdmaDevice.DevName)
		} else {
			continue
		}

		ds := &pluginapi.DeviceSpec{
			ContainerPath: devPath,
			HostPath:      devPath,
			Permissions:   "rw",
		}
		devicesList = append(devicesList, ds)
	}

	// for /dev/infiniband/rdma_cm
	rdma_cm_paths := []string{
		"/dev/infiniband/rdma_cm",
	}
	for _, dev := range rdma_cm_paths {
		devicesList = append(devicesList, &pluginapi.DeviceSpec{
			ContainerPath: dev,
			HostPath:      dev,
			Permissions:   "rw",
		})
	}

	response.Devices = devicesList

	return &response, nil
}

func (m *RdmaDevicePlugin) cleanup() error {
	if err := os.Remove(m.socket); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (m *RdmaDevicePlugin) healthcheck() {
	ctx, cancel := context.WithCancel(context.Background())

	xids := make(chan *pluginapi.Device)
	go watchXIDs(ctx, m.devs, xids)

	for {
		select {
		case <-m.stop:
			cancel()
			return
		case dev := <-xids:
			m.unhealthy(dev)
		}
	}
}

// Serve starts the gRPC server and register the device plugin to Kubelet
func (m *RdmaDevicePlugin) Serve(resourceName string) error {
	err := m.Start()
	if err != nil {
		log.Errorf("Could not start device plugin: %v", err)
		return err
	}
	log.Infof("Starting to serve on %s", m.socket)

	err = m.Register(pluginapi.KubeletSocket, resourceName)
	if err != nil {
		log.Errorf("Could not register device plugin: %v", err)
		m.Stop()
		return err
	}
	log.Infof("Registered device plugin with Kubelet")

	return nil
}
