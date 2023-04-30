# Service Discovery 
## Overview

This project is an implementation of [Client-Side Discovery Pattern](https://microservices.io/patterns/client-side-discovery.html) intended to use in gRPC projects together Kubernetes (using sidecar pattern).

### Discovery Service  
![discovery](./docs/discovery_service_sequence_diagram.svg)

## Stack

This service is written in GoLang and contains the implementation of Master(gRPC server) and Agent that invokes Reflection and HealthCheck on services and send it over to Master(register and unregister). 

Data Source: SQLite