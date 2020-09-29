# Overview

The klight (k8s lightchain) controller is responsible for:

1. listening to pods with the `klight.networkId` label
2. take the ip from `lightpeer` containers inside those pods, and the `klightPort` port (defaulting to `9081`)
3. join the `lightpeer` containers to the required network through the gRPC interface.

# installing the controller

There is no publicly available image for the controller, so it needs to be built from scratch. This can be done by building the Docker image via the provided `/Dockerfile`, pushing it to a repository and using the `/k8s/controller.yaml` to deploy the klight controller.

Here is an example of building the docker image and deploying it to a local microk8s cluster:

```bash
# pwd: /src/klight/controller

docker build -t localhost:32000/klight.controller:v0.1.0-k8s .
docker push localhost:32000/klight.controller:v0.1.0-k8s

k apply -f k8s/controller.yaml
```

# ensuring pod access

As mentioned, the controller looks for pods with the `klight.networkId`. At the moment, this is only done for the default namespace. 
Afterwards, in order to form the networks, the controller assumes there is a `lightpeer` gRPC service running at the pod address and will try to connect to it. The connection is done by taking the pod IP address, and looking for a port exposed on the pod with the name `klightPort`. If a port is not found, it will default to `9081`.

It is important that when a pod wants to join a lightchain, it has the following:

* a `klight.networkId` tag with the network it wants to be part of
* a `klightPort` with the port exposed for the `lightpeer` gRPC service (or `9081` for the default)
* a `lightpeer` sidecar running inside the pod, or a custom `lightpeer` service, listening on the `klightPort`  

# Reconciler logic

It is the responsibility of the reconciler to join pods together based on the network tags. It is, however, not responsible for cleaning the formed networks or tracking live pods. Once pods are joined in a network, it is up to the `lightpeer`s to maintain the connections and to clean up dead peers.
The reconciler will issue health checks also to the pods to make sure they are available and to clean its own stacks, but it will not participate in the maintenance of the network itself.

Reconciling should be as stateless as possible, as k8s pods are volatile and there are no guarantees of what's up and what's down. But at the same time it needs to keep track of contact pods for each network id, such that new pods can join the network if it already exists.
This can be done by maintaining an IP stack for each network id, with the newest known pod at the top of the stack. When a new pod (podA) wants to join the network, reconciliation works as follows:
 1) stack is empty => podA is the only known one in the network, so push podA.IP to stack
 2) stack is not empty => read, without poping, the first IP on the stack and try to connect podA to that IP.
 2.1) Connection is successful => push podA.IP to the stack
 2.2) Connection unsuccessful => pop the head of the stack, cleaning up down pods, and jump to step 1.

This method should ensure that all known pods for a network are recorded, and if there is one alive on that network, then new pods will be able to join it. And since it is poping the existing nodes in case of unsuccessful connections, the stack should be pretty clean (although there can still be leftovers at the bottom which need cleaning).

# Limitations

1) The controller currently only runs on the default namespace. Future releases should make it possible to run on different namespaces.

2) The reconciler will not actively clean the dead pods from stacks. This should change