package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	logger "github.com/sirupsen/logrus"

	"github.com/gbauso/service-discovery-grpc-k8s/agent/client"
	"github.com/gbauso/service-discovery-grpc-k8s/agent/client/interceptors"
	"github.com/gbauso/service-discovery-grpc-k8s/agent/client/interfaces"
	"github.com/gbauso/service-discovery-grpc-k8s/agent/entity"
	"github.com/gbauso/service-discovery-grpc-k8s/agent/service"
	discovery "github.com/gbauso/service-discovery-grpc-k8s/grpc_gen"
	uuid "github.com/nu7hatch/gouuid"
	"google.golang.org/grpc"
	hc "google.golang.org/grpc/health/grpc_health_v1"
	reflection "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

var (
	masterNodeUrl = flag.String("master-node", "", "The master node url")
	serviceUrl    = flag.String("service-url", "", "The service url")
	serviceName   = flag.String("service", "", "The service name")
	logPath       = flag.String("log-path", "/tmp/discovery_agent-%.log", "Log path")
)

func main() {
	log := logger.New()
	id, _ := uuid.NewV4()
	log.SetFormatter(&logger.JSONFormatter{})
	flag.Parse()
	fileName := fmt.Sprintf(*logPath, id)
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Errorf("error opening file: %v", err)
		panic(err)
	}
	defer f.Close()
	wrt := io.MultiWriter(os.Stdout, f)

	log.SetOutput(wrt)

	correlationInterceptor := interceptors.NewCorrelationInterceptor()

	serviceConn, err := grpc.Dial(*serviceUrl, grpc.WithInsecure(), grpc.WithUnaryInterceptor(correlationInterceptor.ClientInterceptor))
	if err != nil {
		log.Errorf("error when connect to %s: %v", *serviceUrl, err)
		panic(err)
	}
	defer serviceConn.Close()

	masterConn, err := grpc.Dial(*masterNodeUrl, grpc.WithInsecure(), grpc.WithUnaryInterceptor(correlationInterceptor.ClientInterceptor))
	if err != nil {
		log.Errorf("error when connect to %s: %v", *masterNodeUrl, err)
		panic(err)
	}
	defer masterConn.Close()

	svc := entity.NewService(*serviceUrl, *serviceName, id.String())

	discoveryGrpcClient := discovery.NewDiscoveryServiceClient(masterConn)
	healthCheckGrpcClient := hc.NewHealthClient(serviceConn)
	reflectionGrpcClient := reflection.NewServerReflectionClient(serviceConn)

	var discoveryClient interfaces.DiscoveryClient = client.NewDiscoveryClient(discoveryGrpcClient)
	var healthCheckClient interfaces.HealthCheckClient = client.NewHealthCheckClient(healthCheckGrpcClient, log)
	var reflectionClient interfaces.ReflectionClient = client.NewReflectionClient(reflectionGrpcClient)

	agentService := service.NewAgentService(reflectionClient, discoveryClient, healthCheckClient, log)
	agentService.HandleService(svc)
}
