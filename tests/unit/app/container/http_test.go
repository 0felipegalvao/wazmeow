package container

import (
	"testing"

	"wazmeow/internal/app/container"
	infraContainer "wazmeow/internal/infra/container"
)

func TestNewHTTPContainer(t *testing.T) {
	// Create infrastructure container
	cfg := createTestConfig()

	infraCont, err := infraContainer.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create infrastructure container: %v", err)
	}
	defer infraCont.Close()

	// Create use case container
	useCaseCont, err := container.NewUseCaseContainer(infraCont)
	if err != nil {
		t.Fatalf("Failed to create use case container: %v", err)
	}

	// Test HTTP container creation
	httpCont, err := container.NewHTTPContainer(infraCont, useCaseCont, cfg)
	if err != nil {
		t.Fatalf("NewHTTPContainer() failed: %v", err)
	}

	if httpCont == nil {
		t.Fatal("NewHTTPContainer() returned nil")
	}

	// Test GetServerManager
	serverManager := httpCont.GetServerManager()
	if serverManager == nil {
		t.Error("GetServerManager() returned nil")
	}

	// Test GetServerInfo
	serverInfo := httpCont.GetServerInfo()
	if serverInfo.Address == "" {
		t.Error("GetServerInfo() returned empty address")
	}
}

func TestHTTPContainer_GetServerManager(t *testing.T) {
	// Create infrastructure container
	cfg := createTestConfig()

	infraCont, err := infraContainer.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create infrastructure container: %v", err)
	}
	defer infraCont.Close()

	// Create use case container
	useCaseCont, err := container.NewUseCaseContainer(infraCont)
	if err != nil {
		t.Fatalf("Failed to create use case container: %v", err)
	}

	// Create HTTP container
	httpCont, err := container.NewHTTPContainer(infraCont, useCaseCont, cfg)
	if err != nil {
		t.Fatalf("Failed to create HTTP container: %v", err)
	}

	serverManager := httpCont.GetServerManager()
	if serverManager == nil {
		t.Error("GetServerManager() returned nil")
	}
}

func TestHTTPContainer_GetServerInfo(t *testing.T) {
	// Create infrastructure container
	cfg := createTestConfig()

	infraCont, err := infraContainer.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create infrastructure container: %v", err)
	}
	defer infraCont.Close()

	// Create use case container
	useCaseCont, err := container.NewUseCaseContainer(infraCont)
	if err != nil {
		t.Fatalf("Failed to create use case container: %v", err)
	}

	// Create HTTP container
	httpCont, err := container.NewHTTPContainer(infraCont, useCaseCont, cfg)
	if err != nil {
		t.Fatalf("Failed to create HTTP container: %v", err)
	}

	serverInfo := httpCont.GetServerInfo()
	
	// Check that server info has expected fields
	if serverInfo.Address == "" {
		t.Error("GetServerInfo() returned empty address")
	}
	
	// The address should be in format "host:port"
	expectedAddress := "localhost:8080"
	if serverInfo.Address != expectedAddress {
		t.Errorf("Expected address %s, got %s", expectedAddress, serverInfo.Address)
	}
}
