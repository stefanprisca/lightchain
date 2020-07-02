# Kube Network

Simple example showing how a lightchain can be set-up for two services in kubernetes.
The set-up consists of two services: a *persister* which will save messages using a lightpeer, and a *reader* which reads the messages.

# Starting the example

One can use the Makefile to run the example in Kubernetes:
```
make build-lightpeer
make deploy-example
```

As the lightpeer image is not publicly available on a docker repo, you'll have to build it yourself. The example is built against a local microk8s instance, so the makefile will use the container registry running in microk8s at *localhost:32000*. You can change this to your own registry by modifying the variable inside the makefile, and the image names from the kubernetes files.

# Walk through