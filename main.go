package main

import (
	"flag"
	"os"
	"syscall"

	"github.com/fsnotify/fsnotify"
	log "k8s.io/klog"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

var (
	MasterNetDevice string = ""
)

func main() {
	// Parse command-line arguments
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagMasterNetDev := flag.String("master", "", "Master ethernet network device for SRIOV, ex: eth1.")
	flagResourceName := flag.String("resource-name", defaultResourceName, "Define the default resource name: tencent.com/rdma.")
	log.InitFlags(flag.CommandLine)
	flag.Parse()

	defer log.Flush()

	if *flagMasterNetDev != "" {
		MasterNetDevice = *flagMasterNetDev
	}

	log.Info("Fetching devices.")

	devList, err := GetDevices(MasterNetDevice)
	if err != nil {
		log.Errorf("Error to get IB device: %v", err)
		return
	}
	if len(devList) == 0 {
		log.Info("No devices found.")
		return
	}

	log.V(1).Infof("RDMA device list: %v", devList)
	log.Info("Starting FS watcher.")
	watcher, err := newFSWatcher(pluginapi.DevicePluginPath)
	if err != nil {
		log.Errorf("Failed to created FS watcher.")
		os.Exit(1)
	}
	defer watcher.Close()

	log.Info("Starting OS watcher.")
	sigs := newOSWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	restart := true
	var devicePlugin *RdmaDevicePlugin

L:
	for {
		if restart {
			if devicePlugin != nil {
				devicePlugin.Stop()
			}

			devicePlugin = NewRdmaDevicePlugin(MasterNetDevice)
			if err := devicePlugin.Serve(*flagResourceName); err != nil {
				log.Info("Could not contact Kubelet, retrying. Did you enable the device plugin feature gate?")
			} else {
				restart = false
			}
		}

		select {
		case event := <-watcher.Events:
			if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
				log.Infof("inotify: %s created, restarting.", pluginapi.KubeletSocket)
				restart = true
			}

		case err := <-watcher.Errors:
			log.Infof("inotify: %s", err)

		case s := <-sigs:
			switch s {
			case syscall.SIGHUP:
				log.Infof("Received SIGHUP, restarting.")
				restart = true
			default:
				log.Infof("Received signal \"%v\", shutting down.", s)
				devicePlugin.Stop()
				break L
			}
		}
	}
}
