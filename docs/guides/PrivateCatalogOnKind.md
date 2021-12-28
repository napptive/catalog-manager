# Deploying a private catalog on Kind

This document contains instructions on how to deploy a private catalog on Kind.

## Creating a kind cluster

```
cat <<EOF | kind create cluster --name napptive-catalog --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
kubeadmConfigPatches:
- |
  kind: ClusterConfiguration
  apiServer:
    certSANs:
    - napptive.local
    - localhost
    - 127.0.0.1
    extraArgs:
      "service-node-port-range": "30000-40000"
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
  #catalog-manager gRPC
  - containerPort: 37060
    hostPort: 37060
    listenAddress: "0.0.0.0"
    protocol: tcp
  #catalog-manager Admin API gRPC
  - containerPort: 37062
    hostPort: 37062
    listenAddress: "0.0.0.0"
    protocol: tcp
- role: worker
EOF
```

## Preparing the environment

```
kubectl create ns napptive
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

## Deploying postgres

Clone the repository and generate the files

```bash
cd <your_development_path>
git clone git@github.com:napptive/rdbms.git
cd rdbms
TARGET_DOCKER_REGISTRY=napptive VERSION=latest TARGET_K8S_NAMESPACE=napptive make k8s-kind
kubectl create -f build/k8s
```

## Creating the Kubernetes entities

First create the target yaml files:

```bash
TARGET_DOCKER_REGISTRY=napptive VERSION=latest TARGET_K8S_NAMESPACE=napptive make k8s-kind
```

Deploy them on the cluster

```bash
kubectl create -f build/k8s
```

At this point all entities are being created and trying to connect to each other. Bear in mind that ElasticSearch tends to take some time to startup on kind, so you may need to wait for it even a couple of minutes.

## Testing the catalog

Use the [catalog-cli](https://github.com/napptive/catalog-cli)

```bash
$ catalog --catalogAddress localhost --catalogPort 37060 --useTLS=false list
APPLICATION    NAME
```
