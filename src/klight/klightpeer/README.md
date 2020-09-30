# klight peer

A wrapper over the general `lightpeer` service, making it easy to share state through a `stateFile`. It uses a simple file listener to check from updates to the state file, and sharing these across the network. It also updates the state file with changes received from other peers.


# Considerations (/Limitations)

* listens to one file only, given as input argument `stateFile`
* Files always overridden by the latest change in the network. i.e. if two peers edit the same file, the latest received edit in the network is accepted as the valid one, overriding other changes.
* Does not implement any consensus algorithm. Changes are accepted based on the logic from the underlying `lightpeer` service.
    * Future releases could implement a RoundRobin ticketing system to allow a deterministic communication over the network.

# Install

The only way to install is to build the docker image, push it to a repo and use it from k8s. For example, you can build and push it to a local microk8s repository:

```
docker build -t localhost:32000/klight.peer:v0.1.0-k8s .
docker push localhost:32000/klight.peer:v0.1.0-k8s
```

See [k8s-stateFiles](../../../examples/k8s-stateFiles) for an usage example.
