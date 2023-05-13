//go:build integration
// +build integration

package integration

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	hc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	ref "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	"github.com/gbauso/service-discovery-grpc-k8s/agent/client"
	"github.com/gbauso/service-discovery-grpc-k8s/agent/entity"
	"github.com/gbauso/service-discovery-grpc-k8s/agent/service"
	pb "github.com/gbauso/service-discovery-grpc-k8s/grpc_gen"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type integrationTestSuite struct {
	suite.Suite
	masterClient *grpc.ClientConn
	server       *grpc.Server
}

func createGrpcServer() *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterCityServiceServer(s, pb.UnimplementedCityServiceServer{})
	reflection.Register(s)

	healthCheckServer := health.NewServer()
	healthCheckServer.SetServingStatus("cityinformation.CityService", hc.HealthCheckResponse_SERVING)
	healthCheckServer.SetServingStatus("grpc.reflection.v1alpha.ServerReflection", hc.HealthCheckResponse_SERVING)

	hc.RegisterHealthServer(s, healthCheckServer)

	return s
}

func startServer(s *grpc.Server) error {
	lis, err := net.Listen("tcp", ":6500")
	if err != nil {
		return err
	}

	errorCh := make(chan error)

	go func() {
		if err = s.Serve(lis); err != nil {
			errorCh <- err
		}
	}()

	return nil
}

func stopServer(s *grpc.Server) {
	s.Stop()
}

func getServiceConnection() (*grpc.ClientConn, error) {
	cityInfoClient, err := grpc.Dial(":6500", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return cityInfoClient, nil
}

func getAgentService(masterConn *grpc.ClientConn, serviceConn *grpc.ClientConn) *service.AgentService {
	log := logrus.New()

	discoveryGrpcClient := pb.NewDiscoveryServiceClient(masterConn)
	healthCheckGrpcClient := hc.NewHealthClient(serviceConn)
	reflectionGrpcClient := ref.NewServerReflectionClient(serviceConn)

	discoveryClient := client.NewDiscoveryClient(discoveryGrpcClient)
	healthCheckClient := client.NewHealthCheckClient(healthCheckGrpcClient, log)
	reflectionClient := client.NewReflectionClient(reflectionGrpcClient)

	agentService := service.NewAgentService(reflectionClient, discoveryClient, healthCheckClient, log)

	return agentService
}

func TestDiscoveryAgent(t *testing.T) {
	// Create a gRPC client connection
	masterClient, err := grpc.Dial(os.Getenv("MASTER_URL"), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer masterClient.Close()

	suite.Run(t, &integrationTestSuite{masterClient: masterClient, server: createGrpcServer()})

	t.Logf("Agent Test Passed")
}

func (s *integrationTestSuite) Test_HappyFlow() {
	err := startServer(s.server)
	if err != nil {
		s.Error(err)
	}

	client, err := getServiceConnection()
	if err != nil {
		s.Error(err)
	}
	defer client.Close()

	service := getAgentService(s.masterClient, client)

	svc := entity.NewService("localhost:6500", "stub", "123")

	routineCheck := make(chan bool)
	defer close(routineCheck)

	go func() {
		time.Sleep(2 * time.Second)
		discoveryClient := pb.NewDiscoveryServiceClient(s.masterClient)
		beforeStop, _ := discoveryClient.GetServiceHandlers(context.Background(), &pb.DiscoverySearchRequest{ServiceDefinition: "cityinformation.CityService"})
		s.Len(beforeStop.Services, 1)

		stopServer(s.server)
		time.Sleep(2 * time.Second)

		afterStop, _ := discoveryClient.GetServiceHandlers(context.Background(), &pb.DiscoverySearchRequest{ServiceDefinition: "cityinformation.CityService"})
		s.Len(afterStop.Services, 0)
		routineCheck <- true
		return
	}()

	err = service.HandleService(svc)
	<-routineCheck
	s.Nil(err)
}
