FROM mellanox/mofed421_docker

COPY bin/k8s-rdma-device-plugin /usr/local/bin/

ENTRYPOINT ["k8s-rdma-device-plugin"]
