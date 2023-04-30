package interfaces

import "github.com/gbauso/service-discovery-grpc-k8s/agent/entity"

type DiscoveryClient interface {
	RegisterService(svc *entity.Service) error
	UnRegisterService(svc *entity.Service) error
}

type HealthCheckClient interface {
	WatchService(service, correlationId string) error
}

type ReflectionClient interface {
	GetImplementedServices(svc *entity.Service) ([]string, error)
}
