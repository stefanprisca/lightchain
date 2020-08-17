The klight (k8s lightchain) operator is responsible for:

1. listening to pods with the `klight.networkId` label
2. take the ip from `lightpeer` containers inside those pods
3. join the `lightpeer` containers to the required network.