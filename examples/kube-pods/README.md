__Work In Progress__

# k8s pod example

This is the simplest configuration for running a lightchain inside k8s.
The example includes two pods, one writing the current date, and the other reading it. It illustrates how to connect the pods using lightpod sidecars, and the `klight.networkId` label.

# building the example

In order to run, one can follow the makefile commands:

1. make install-klight-operator
2. make deploy

# Deep dive

In order to connect the pods together, the `klight controller` is needed. This is a k8s controller listening for pods with `klight.networkId` labels, and connecting them to the same network.
_Note_ that the controller will connect the pods by issuing gRpc commands to the `lightpeer` sidecar running inside the pod. This container is matched by name, so it is important to keep the  `lightpeer` name for it, and expose the `9081` port for gRpc communication.

Once both pods are in the same network, the `lightpeers` start communicating with each other. In the current setup, each `lightpeer` listens to the files in the `statePath` directory, and pushes the changes to the network. So, when the _writer_ container writes to the html file, the `lightpeer` sidecar will pick up the changes, and push them to the network. The `lightpeer` sidecar on the reader side receives them, and writes them to the shared directory for the _reader_. From there, the date is displayed in the _reader_ console. If everything works fine, the date should change every 2 seconds.

_Note_ that the communication from the _writer_/_reader_ and their `lightpeer` sidecars is done through a shared volume. This can also happen by using sockets, or directly through the `lightpeer` gRPC API.