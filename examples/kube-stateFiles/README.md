# k8s pod example

This is the simplest configuration for running a lightchain inside k8s.
The example includes two pods, one writing the current date, and the other reading it. It illustrates how to connect the pods using lightpod sidecars, and the `klight.networkId` label.

# building the example

The example uses the `klightpeer` version of the lightpeer, which is a wrapper over the lightpeer interface listening for changes to a `stateFile` and pushing them on the network. See [klightpeer](../../src/klight/klightpeer/README.md) for details. You also need to build this before moving on with the example. You'll also need the klight [controller](../../src/klight/controller/README.md).

In order to run, one can follow the makefile commands:

```
make install-klight-operator
make deploy
```

# Deep dive

In order to connect the pods together, the `klight controller` is needed. This is a k8s controller listening for pods with `klight.networkId` labels, and connecting them to the same network.
_Note_ that the controller will connect the pods by issuing gRpc commands to the `klightpeer` sidecar running inside the pod. This container is matched by name, so it is important to keep the  `klightpeer` name for it, and expose the `9081` port for gRpc communication.

Once both pods are in the same network, the `klightpeers` start communicating with each other. In the current setup, each `klightpeer` listens to the files in the `statePath` directory, and pushes the changes to the network. So, when the _writer_ container writes to the html file, the `klightpeer` sidecar will pick up the changes, and push them to the network. The `klightpeer` sidecar on the reader side receives them, and writes them to the shared directory for the _reader_. From there, the date is displayed in the _reader_ console. If everything works fine, the date should change every 2 seconds.

_Note_ that the communication from the _writer_/_reader_ and their `klightpeer` sidecars is done through a shared volume. This can also happen by using sockets, or directly through the `lightpeer` gRPC API.
