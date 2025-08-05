package container

import (
	"testing"

	"wazmeow/internal/app/container"
	infraContainer "wazmeow/internal/infra/container"
)

func TestNewUseCaseContainer(t *testing.T) {
	// Create infrastructure container first
	cfg := createTestConfig()

	infraCont, err := infraContainer.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create infrastructure container: %v", err)
	}
	defer infraCont.Close()

	// Test use case container creation
	useCaseCont, err := container.NewUseCaseContainer(infraCont)
	if err != nil {
		t.Fatalf("NewUseCaseContainer() failed: %v", err)
	}

	if useCaseCont == nil {
		t.Fatal("NewUseCaseContainer() returned nil")
	}

	// Test GetSessionUseCases
	sessionUseCases := useCaseCont.GetSessionUseCases()
	
	if sessionUseCases.Create == nil {
		t.Error("SessionUseCases.Create is nil")
	}
	
	if sessionUseCases.Connect == nil {
		t.Error("SessionUseCases.Connect is nil")
	}
	
	if sessionUseCases.Disconnect == nil {
		t.Error("SessionUseCases.Disconnect is nil")
	}
	
	if sessionUseCases.List == nil {
		t.Error("SessionUseCases.List is nil")
	}
	
	if sessionUseCases.Delete == nil {
		t.Error("SessionUseCases.Delete is nil")
	}
	
	if sessionUseCases.Resolve == nil {
		t.Error("SessionUseCases.Resolve is nil")
	}
	
	if sessionUseCases.SetProxy == nil {
		t.Error("SessionUseCases.SetProxy is nil")
	}
	
	if sessionUseCases.AutoReconnect == nil {
		t.Error("SessionUseCases.AutoReconnect is nil")
	}

	// Test GetWhatsAppUseCases
	whatsappUseCases := useCaseCont.GetWhatsAppUseCases()
	
	if whatsappUseCases.GenerateQR == nil {
		t.Error("WhatsAppUseCases.GenerateQR is nil")
	}
	
	if whatsappUseCases.PairPhone == nil {
		t.Error("WhatsAppUseCases.PairPhone is nil")
	}
	
	if whatsappUseCases.SendMessage == nil {
		t.Error("WhatsAppUseCases.SendMessage is nil")
	}
}

func TestUseCaseContainer_SessionUseCases(t *testing.T) {
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

	sessionUseCases := useCaseCont.GetSessionUseCases()

	// Test that all session use cases are properly initialized
	tests := []struct {
		name    string
		useCase interface{}
	}{
		{"Create", sessionUseCases.Create},
		{"Connect", sessionUseCases.Connect},
		{"Disconnect", sessionUseCases.Disconnect},
		{"List", sessionUseCases.List},
		{"Delete", sessionUseCases.Delete},
		{"Resolve", sessionUseCases.Resolve},
		{"SetProxy", sessionUseCases.SetProxy},
		{"AutoReconnect", sessionUseCases.AutoReconnect},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.useCase == nil {
				t.Errorf("SessionUseCases.%s is nil", tt.name)
			}
		})
	}
}

func TestUseCaseContainer_WhatsAppUseCases(t *testing.T) {
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

	whatsappUseCases := useCaseCont.GetWhatsAppUseCases()

	// Test that all WhatsApp use cases are properly initialized
	tests := []struct {
		name    string
		useCase interface{}
	}{
		{"GenerateQR", whatsappUseCases.GenerateQR},
		{"PairPhone", whatsappUseCases.PairPhone},
		{"SendMessage", whatsappUseCases.SendMessage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.useCase == nil {
				t.Errorf("WhatsAppUseCases.%s is nil", tt.name)
			}
		})
	}
}
