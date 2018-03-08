FROM golang:1.10-stretch

RUN apt-get update && \
    apt-get install -y libibverbs-dev git && \
    mkdir -p /go/src/github.com/hustcat/ && \
    cd /go/src/github.com/hustcat && \
    git clone https://github.com/hustcat/k8s-rdma-device-plugin && \
    cd k8s-rdma-device-plugin && ./build && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/hustcat/k8s-rdma-device-plugin/bin
ENTRYPOINT ["./k8s-rdma-device-plugin"]
