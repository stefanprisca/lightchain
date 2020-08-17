__DEPRECATED__

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

If everything works fine, the persister should write down a "Hello World" message with a timestamp every one minute, and the reader will display it. Inspecting the reader logs should print the messages.

# Walkthrough

This example uses two kubernetes stateful sets to deploy a *persister* application and a *reader* application.
Both of these communicate through a lightchain network, where the *persister* writes a message and the *reader* reads and displays it.

## Starting a lightpeer
In order to connect to the lightchain network, both applications run a sidecar container from the `lightpeer` image:

```
- name: lightpeer
        image: localhost:32000/lightpeer:example
        imagePullPolicy: Always
        command: ["./lightpeer"]
        args: ["-repo=$(LP_BLOCKREPO)", "-host=0.0.0.0"]
        resources:
          limits:
            cpu: 100m
            memory: 100Mi
          requests:
            cpu: 50m
            memory: 50Mi
        env:
          - name: LP_BLOCKREPO
            value: "/blockRepository"
        ports:
        - containerPort: 9081 # gRPC port.
        volumeMounts:
        - name: reader-lightpeer-sto
          mountPath: /blockRepository
```
The configuration required for the `lightpeer` is pretty simple: we need a volume mount to act as storage, and to configure the host on *0.0.0.0* so it listens to connections from the other containers.

The application then only needs to create a grpc client connection to the lightpeer, using it's address:
```go
var conn *grpc.ClientConn
conn, err := grpc.Dial(lpAddress, grpc.WithInsecure())
```

For this application, the `lpAddress` is passed through the environment variable `LP_ADDRESS`. Since the containers run in the same pod, this will simply be *:9081*.

## Forming the network

As a design decision, the client application is responsible for connecting the peers together and forming the actual network. This decision was made in order to simplify the logic contained in the peers themselves, and to allow applications to configure the network however they see fit.

In order to connect two `lightpeers` together, simply use the `JoinNetwork` request. For this example, the *reader* will connect to the *persister* peer in the following way:
```go
client := pb.NewLightpeerClient(conn)
ctx := context.Background()
log.Println("Trying to join persister at address %s", *persisterAddress)
_, err = client.JoinNetwork(ctx, &pb.JoinRequest{
    Address: *persisterAddress,
})
if err != nil {
    log.Fatalf("did not connect: %s", err)
}
```  
The `persisterAddress` from above corresponds to the address of the *persister-0* pod running as part of the persister stateful set: *persister-0.persister.lightchain.svc.cluster.local:9081*.

