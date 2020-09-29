# klight peer

A file listener for kubernetes purposes. It listens to file changes and synchronizing the files automatically within the `ligthchain` network via the local `lightpeer`.


# Considerations (/Limitations)

* Assumes that the `lightpeer` runs on the same local network, as a sidecar to the same pod.
* listens to one file only, given as input argument `stateFile`
* Files always overridden by the latest change in the network. i.e. if two peers edit the same file, the latest received edit in the network is accepted as the valid one, overriding other changes.


# Install

The only way to install is to build the docker image, push it to a repo and use it from k8s. For example, you can build and push it to a local microk8s repository:

```
docker build -t localhost:32000/klight.peer:v0.1.0-k8s .
docker push localhost:32000/klight.peer:v0.1.0-k8s
```

See [k8s pods](../../../examples/k8s-pods) for an usage example.
