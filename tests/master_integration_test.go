//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/gbauso/service-discovery-grpc-k8s/grpc_gen"
	"google.golang.org/grpc"
)

func TestDiscoveryService(t *testing.T) {
	// Create a gRPC client connection
	conn, err := grpc.Dial(os.Getenv("MASTER_URL"), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	// Create a gRPC client instance
	client := grpc_gen.NewDiscoveryServiceClient(conn)

	// Test GetServiceHandlers RPC method
	response, err := client.GetServiceHandlers(ctx, &grpc_gen.DiscoverySearchRequest{ServiceDefinition: "handler1"})
	if err != nil {
		t.Fatalf("Failed to call GetServiceHandlers: %v", err)
	}
	if len(response.Services) != 0 {
		t.Errorf("Unexpected number of services: got %d, want 0", len(response.Services))
	}

	// Test RegisterServiceHandlers RPC method
	_, err = client.RegisterServiceHandlers(ctx, &grpc_gen.RegisterServiceHandlersRequest{
		Service:   "service1",
		ServiceId: "123",
		Handlers:  []string{"handler1"},
	})
	if err != nil {
		t.Fatalf("Failed to call RegisterServiceHandlers: %v", err)
	}

	afterRegister, err := client.GetServiceHandlers(ctx, &grpc_gen.DiscoverySearchRequest{ServiceDefinition: "handler1"})
	if err != nil {
		t.Fatalf("Failed to call GetServiceHandlers: %v", err)
	}
	if len(afterRegister.Services) != 1 {
		t.Errorf("Unexpected number of services: got %d, want 1", len(afterRegister.Services))
	}

	// Test UnregisterService RPC method
	_, err = client.UnregisterService(ctx, &grpc_gen.UnregisterServiceRequest{ServiceId: "123"})
	if err != nil {
		t.Fatalf("Failed to call UnregisterService: %v", err)
	}

	afterUnegister, err := client.GetServiceHandlers(ctx, &grpc_gen.DiscoverySearchRequest{ServiceDefinition: "handler1"})
	if err != nil {
		t.Fatalf("Failed to call GetServiceHandlers: %v", err)
	}
	if len(afterUnegister.Services) != 0 {
		t.Errorf("Unexpected number of services: got %d, want 0", len(afterUnegister.Services))
	}
}
