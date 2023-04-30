package server

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gbauso/service-discovery-grpc-k8s/grpc_gen"
	"github.com/gbauso/service-discovery-grpc-k8s/master/entity"
	"github.com/gbauso/service-discovery-grpc-k8s/master/repository/stub"
)

func Test_GetServiceHandlers_Success_ShoudReturnListOfServices(t *testing.T) {
	// Arrange
	fakeReponse := []string{"fake1", "fake2"}
	getAliveServicesFn := func(service string) ([]string, error) {
		return fakeReponse, nil
	}

	repository := &stub.ServiceHandlerRepositoryStub{GetAliveServicesFn: getAliveServicesFn}
	server := NewServer(repository)
	request := &grpc_gen.DiscoverySearchRequest{ServiceDefinition: "svc1"}

	// Act
	response, err := server.GetServiceHandlers(context.Background(), request)

	if err != nil || !reflect.DeepEqual(response.Services, fakeReponse) {
		t.Error(err)
	}
}

func Test_GetServiceHandlers_Fail_ShoudReturnAnError(t *testing.T) {
	// Arrange
	getAliveServicesFn := func(service string) ([]string, error) {
		return nil, errors.New("ErrorOnRepository")
	}

	repository := &stub.ServiceHandlerRepositoryStub{GetAliveServicesFn: getAliveServicesFn}
	server := NewServer(repository)
	request := &grpc_gen.DiscoverySearchRequest{ServiceDefinition: "svc1"}

	// Act
	response, err := server.GetServiceHandlers(context.Background(), request)

	if err == nil || response != nil {
		t.Error()
	}
}

func Test_RegisterServiceHandlers_Success_ShoudReturnListOfServices(t *testing.T) {
	// Arrange
	insertFn := func(serviceHandlers ...entity.ServiceHandler) error {
		return nil
	}

	repository := &stub.ServiceHandlerRepositoryStub{InsertFn: insertFn}
	server := NewServer(repository)
	request := &grpc_gen.RegisterServiceHandlersRequest{Service: "svc", ServiceId: "12", Handlers: []string{"host1"}}

	// Act
	response, err := server.RegisterServiceHandlers(context.Background(), request)

	if err != nil || response == nil {
		t.Error(err)
	}
}

func Test_RegisterServiceHandlers_Fail_ShoudReturnAnError(t *testing.T) {
	// Arrange
	insertFn := func(serviceHandlers ...entity.ServiceHandler) error {
		return errors.New("ErrorOnRepository")
	}

	repository := &stub.ServiceHandlerRepositoryStub{InsertFn: insertFn}
	server := NewServer(repository)
	request := &grpc_gen.RegisterServiceHandlersRequest{Service: "svc", ServiceId: "12", Handlers: []string{"host1"}}

	// Act
	response, err := server.RegisterServiceHandlers(context.Background(), request)

	if err == nil || response != nil {
		t.Error()
	}
}

func Test_UnregisterService_Success_ShoudReturnListOfServices(t *testing.T) {
	// Arrange
	getServiceByIdFn := func(serviceId string) ([]entity.ServiceHandler, error) {
		return []entity.ServiceHandler{}, nil
	}

	updateFn := func(serviceHandlers ...entity.ServiceHandler) error {
		return nil
	}

	repository := &stub.ServiceHandlerRepositoryStub{UpdateFn: updateFn, GetByServiceIdFn: getServiceByIdFn}
	server := NewServer(repository)
	request := &grpc_gen.UnregisterServiceRequest{ServiceId: "12"}

	// Act
	response, err := server.UnregisterService(context.Background(), request)

	if err != nil || response == nil {
		t.Error(err)
	}
}

func Test_UnregisterService_Fail_ShoudReturnAnError(t *testing.T) {
	// Arrange
	getServiceByIdFn := func(serviceId string) ([]entity.ServiceHandler, error) {
		return nil, errors.New("ErrorOnRepository")
	}

	updateFn := func(serviceHandlers ...entity.ServiceHandler) error {
		return errors.New("ErrorOnRepository")
	}

	repository := &stub.ServiceHandlerRepositoryStub{UpdateFn: updateFn, GetByServiceIdFn: getServiceByIdFn}
	server := NewServer(repository)
	request := &grpc_gen.UnregisterServiceRequest{ServiceId: "12"}

	// Act
	response, err := server.UnregisterService(context.Background(), request)

	if err == nil || response != nil {
		t.Error()
	}
}
