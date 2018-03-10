# RDMA device plugin for Kubernetes

## Introduction

`k8s-rdma-device-plugin` is a [device plugin](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/resource-management/device-plugin.md) for Kubernetes to manage [RDMA](https://en.wikipedia.org/wiki/Remote_direct_memory_access) device.

RDMA(remote direct memory access) is a high performance network protocal, which has the following major advantages:

* Zero-copy 

  Applications can perform data transfer without the network software stack involvement and data is being send received directly to the buffers without being copied between the network layers.

* Kernel bypass

  Applications can perform data transfer directly from userspace without the need to perform context switches.

* No CPU involvement 

  Applications can access remote memory without consuming any CPU in the remote machine. The remote memory machine will be read without any intervention of remote process (or processor). The caches in the remote CPU(s) won't be filled with the accessed memory content.

You can read [this post](http://www.rdmamojo.com/2014/03/31/remote-direct-memory-access-rdma/) to get more information about `RDMA`.

This plugin allow you to use RDMA device in container of Kubernetes cluster. And more, We can use this plugin work with [sriov-cni](https://github.com/hustcat/sriov-cni) to provide high perfmance network connection for `distributed` application, especially `GPU` distributed application, such as `Tensorflow`,`Spark`, etc.

## Quick Start

### Build

Install libibverbs package, for CentOS:

```
# yum install libibverbs-devel -y
```

Then run `build`:

```
# ./build 
# ls bin
k8s-rdma-device-plugin
```

### Work with Kubernetes

* Preparing RDMA node

Install `ibverbs` libraries, then start `kubelet` with `--feature-gates=DevicePlugins=true`.


* Run device plugin daemon process

```
# bin/k8s-rdma-device-plugin -master eth1 -log-level debug
INFO[0000] Fetching devices.                            
DEBU[0000] RDMA device list: [{{mlx4_1 uverbs1 /sys/class/infiniband_verbs/uverbs1 /sys/class/infiniband/mlx4_1} eth2} {{mlx4_3 uverbs3 /sys/class/infiniband_verbs/uverbs3 /sys/class/infiniband/mlx4_3} eth4} {{mlx4_2 uverbs2 /sys/class/infiniband_verbs/uverbs2 /sys/class/infiniband/mlx4_2} eth3} {{mlx4_4 uverbs4 /sys/class/infiniband_verbs/uverbs4 /sys/class/infiniband/mlx4_4} eth5}] 
INFO[0000] Starting FS watcher.                         
INFO[0000] Starting OS watcher.                         
INFO[0000] Starting to serve on /var/lib/kubelet/device-plugins/rdma.sock 
INFO[0000] Registered device plugin with Kubelet
...
```

or deploy it as a daemonset:

```
# kubectl -n kube-system apply -f rdma-device-plugin.yml
# kubectl -n kube-system get pods
rdma-device-plugin-daemonset-2wbdv         1/1       Running   0          14m
rdma-device-plugin-daemonset-7pwf7         1/1       Running   0          14m
```

* Run RDMA container

```
apiVersion: v1
kind: Pod
metadata:
  name: rdma-pod
spec:
  containers:
    - name: rdma-container
      image: mellanox/mofed421_docker:noop
      securityContext:
        capabilities:
          add: ["ALL"]
      resources:
        limits:
          tencent.com/rdma: 1 # requesting 1 RDMA device
```

`Dockerfile` for `mellanox/mofed421_docker:noop`:

```
FROM mellanox/mofed421_docker:latest

CMD ["/bin/sleep", "360000"]
```

*** NOTE!!! ***

`/dev/infiniband/rdma_cm` device is needed for `uverbs` device in container for run RDMA application. However, Kubernetes is not support pass host device to container, refer to [issue 5607](https://github.com/kubernetes/kubernetes/issues/5607). We wish this can be fixed ASAP.

### Work with sriov-cni plugin

TODO: Kubernetes call DP(device plugin) when [Admit](https://github.com/kubernetes/kubernetes/blob/v1.9.3/pkg/kubelet/kubelet.go#L1998) pod, and call CNI plugin when creating sandbox container. We need a way that pass RDMA device information from DP to CNI. Refer to the [issue 32](https://github.com/hustcat/sriov-cni/issues/32).

### Work with NVIDIA GPU plugin

TODO: For high performance, we should coordinate the `k8s-rdma-device-plugin` and [nvidia device plugin](https://github.com/NVIDIA/k8s-device-plugin), and try to make RDMA devices and GPU devices allocated for the same container are located in the same PCIe switch.
