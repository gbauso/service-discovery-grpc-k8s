package repository

import "github.com/gbauso/service-discovery-grpc-k8s/master/entity"

type ServiceHandlerRepository interface {
	GetAliveServices(service string) ([]string, error)
	GetByServiceId(serviceId string) ([]entity.ServiceHandler, error)
	Update(serviceHandlers ...entity.ServiceHandler) error
	Insert(serviceHandlers ...entity.ServiceHandler) error
}
