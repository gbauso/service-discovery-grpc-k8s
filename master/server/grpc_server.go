package server

import (
	"context"

	pb "github.com/gbauso/service-discovery-grpc-k8s/grpc_gen"
	"github.com/gbauso/service-discovery-grpc-k8s/master/entity"
	repository "github.com/gbauso/service-discovery-grpc-k8s/master/repository/interface"
)

type server struct {
	pb.UnimplementedDiscoveryServiceServer
	repo repository.ServiceHandlerRepository
}

func NewServer(repo repository.ServiceHandlerRepository) *server {
	return &server{repo: repo}
}

func (s *server) RegisterServiceHandlers(ctx context.Context, in *pb.RegisterServiceHandlersRequest) (*pb.RegisterServiceHandlersResponse, error) {
	var serviceHandlers []entity.ServiceHandler
	for _, handler := range in.Handlers {
		serviceHandler := entity.NewServiceHandler(in.Service, in.ServiceId, handler)
		serviceHandlers = append(serviceHandlers, *serviceHandler)
	}

	err := s.repo.Insert(serviceHandlers...)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterServiceHandlersResponse{}, nil
}

func (s *server) GetServiceHandlers(ctx context.Context, in *pb.DiscoverySearchRequest) (*pb.DiscoverySearchResponse, error) {
	services, err := s.repo.GetAliveServices(in.ServiceDefinition)
	if err != nil {
		return nil, err
	}

	return &pb.DiscoverySearchResponse{Services: services}, nil
}

func (s *server) UnregisterService(ctx context.Context, in *pb.UnregisterServiceRequest) (*pb.UnregisterServiceResponse, error) {
	queryResult, err := s.repo.GetByServiceId(in.ServiceId)
	if err != nil {
		return nil, err
	}

	var serviceHanders []entity.ServiceHandler
	for _, result := range queryResult {
		result.MarkAsNotAlive()
		serviceHanders = append(serviceHanders, result)
	}

	err = s.repo.Update(serviceHanders...)
	if err != nil {
		return nil, err
	}

	return &pb.UnregisterServiceResponse{}, nil
}
