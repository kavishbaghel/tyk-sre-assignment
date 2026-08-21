# Tyk SRE Assignment

A lightweight tool written in Go to interact with and operate Kubernetes clusters as per the requirements of the Tyk SRE Assignment. The tool provides a set of operational capabilities through a simple HTTP API, with a focus on Kubernetes workload health, connectivity, and network control.


## Project Setup

Location: https://github.com/kavishbaghel/tyk-sre-assignment/tree/main/golang

In order to build the project run:
```
go mod tidy & go build
```

To run it against a real Kubernetes API server:
```
./tyk-sre-assignment --kubeconfig '/path/to/your/kube/conf' --address ":8080"
```

To execute unit tests:
```
go test -v ./...
```



