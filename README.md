# lightchain - WIP

A lightweight blockchain network which achieves distributed state management through a simple p2p communication protocol based on gRPC.

__WIP__ this project is under development and is subject to changes. Neither documentation nor code are in a stable format.

# Vision

The goal of this project is to provide the simplest way to persist a small data object (state), synchronize it across multiple nodes and provide a common history over it. At the moment, all data is stored using the local file systems. 
Ideally, the client application does not need to know about the underlying peer network. The lightchain becomes invisible, and is responsible for taking a local state and synchronizing it over to the other peers of the network. Moreover, the completely distributed p2p platform means there is no central storage, the network scaling with the application itself. With features such as self-recovery and easy join, the network can maintain the states and automatically rejoin failed instances.
This achieves a 'stateless' behavior, and allows applications to communicate with each other as if they were on the same node, while simplifying client applications by not requiring any library knowledge and being transparent about the way data is stored. 

In short, the goal is to provide:

- lightweight state management
- multi-node support, with auto-recovery and easy join
- automatic scalability through a p2p network
- 'stateless', transparent state storage
- guaranteed global order of events and changes
- same global view of the current state

Note however that the storage is not indexed in any way and provides no complex query capabilities. All knowledge of what is stored and how it should be retrieved belongs to the applications themselves. This allows the lightchain to focus on keeping data in sync, without caring about what is actually stored. This might change as the service evolves.

# Use cases

The goal can be translated to the following use-cases:


## State Storage and Access

As a user, I want to be able to easily store some state of my application, and I expect it to be accessible from all other instances or applications. I expect that my states are available under unique IDs, knowing what I look for and what I store and keeping this information in my applications.

## State Persistence

As a user, I expect that what I store is persisted outside my applications, and can be recovered if the applications themselves fail.

## Transparency

As a user, I expect to easily understand where the data is stored and be able to view the stored data myself. I do not wish to interact with complex services which require APIs an libraries, and depend on these to access and store my data.

## Reliability

As a user, I expect that the service responsible for replicating my data is reliable, and I do not wish to spend time managing it. Once set-up, the service should be able to maintain itself and recover from node failures. I also expect that the data is replicated upon recovery, and that my applications have full access to it after a restart.

# Comparison with other technologies

There are various other solutions out there for sharing data, ranging from databases (SQL/NoSql), highly available file systems (like data lakes) and simple volumes (like Kubernete's PVC). While all these technologies are excellent at solving specific problems, they don't excel when it comes to simply keeping a simple state in sync between your application's instances or nodes. 

Classical databases can be a good option, but they require another whole server to run and be maintained, and as the application grows so will the query frequency. And it can be a bit troublesome to maintain a whole database just to share some info that would fit in a file. The same for highly available file systems.

On the other hand, volume shares like Kubernetes PVC are another simple option to store some data. But these don't guarantee that the data you store will be the same over all your nodes. Nor is there the guarantee of total order over the events.


# Architecture

# Getting Started

# Examples

see [k8s example](examples/kube-network)
