package service

import (
	"github.com/sirupsen/logrus"

	"github.com/gbauso/service-discovery-grpc-k8s/agent/client/interfaces"
	"github.com/gbauso/service-discovery-grpc-k8s/agent/entity"
)

type AgentService struct {
	reflectionClient  interfaces.ReflectionClient
	discoveryClient   interfaces.DiscoveryClient
	healthCheckClient interfaces.HealthCheckClient
	log               *logrus.Logger
}

func NewAgentService(reflectionClient interfaces.ReflectionClient,
	discoveryClient interfaces.DiscoveryClient,
	healthCheckClient interfaces.HealthCheckClient,
	log *logrus.Logger) *AgentService {

	return &AgentService{reflectionClient: reflectionClient,
		discoveryClient:   discoveryClient,
		healthCheckClient: healthCheckClient,
		log:               log}
}

func (as *AgentService) HandleService(service *entity.Service) error {
	as.log.Infof("invoking reflection method on target service: %s", service.Name)
	services, err := as.reflectionClient.GetImplementedServices(service)

	if err != nil {
		as.log.Errorf("error when called reflection method on target service: %v", err)
		return err
	}

	service.SetServices(services)

	as.log.Infof("registering implemented services on master node: %v", services)
	err = as.discoveryClient.RegisterService(service)
	if err != nil {
		as.log.Errorf("error when register services on master node: %v", err)
		return err
	}

	quit := make(chan bool)
	defer close(quit)

	go func() {
		for _, svc := range service.Services {
			as.log.Infof("starting health check on: %s", svc)
			err := as.healthCheckClient.WatchService(svc, service.Id)
			if err != nil {
				as.log.Errorf("error when invoke health checking on target service %s: %v", svc, err)
			}
			as.log.Infof("finished health check on: %s -> unregistering the service", svc)
			as.discoveryClient.UnRegisterService(service)
			quit <- true
			return
		}
	}()

	<-quit

	return nil
}
