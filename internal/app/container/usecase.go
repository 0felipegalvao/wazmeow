package container

import (
	"fmt"

	"wazmeow/internal/infra/container"
	sessionUC "wazmeow/internal/usecases/session"
	whatsappUC "wazmeow/internal/usecases/whatsapp"
	"wazmeow/pkg/logger"
)

// useCaseContainer implements UseCaseContainer interface
type useCaseContainer struct {
	sessionUseCases  SessionUseCases
	whatsappUseCases WhatsAppUseCases
	logger           logger.Logger
	isInitialized    bool
}

// NewUseCaseContainer creates a new use case container
func NewUseCaseContainer(infraContainer *container.Container) (UseCaseContainer, error) {
	uc := &useCaseContainer{
		logger: infraContainer.Logger,
	}

	if err := uc.initialize(infraContainer); err != nil {
		return nil, fmt.Errorf("failed to initialize use case container: %w", err)
	}

	return uc, nil
}

// initialize sets up all use cases
func (uc *useCaseContainer) initialize(infraContainer *container.Container) error {
	logger := infraContainer.Logger
	validator := infraContainer.Validator

	// Initialize session use cases
	uc.sessionUseCases = SessionUseCases{
		Create: sessionUC.NewCreateUseCase(
			infraContainer.SessionRepo,
			logger,
			validator,
		),
		Connect: sessionUC.NewConnectUseCase(
			infraContainer.SessionRepo,
			infraContainer.WhatsAppManager,
			logger,
		),
		Disconnect: sessionUC.NewDisconnectUseCase(
			infraContainer.SessionRepo,
			infraContainer.WhatsAppManager,
			logger,
		),
		List: sessionUC.NewListUseCase(
			infraContainer.SessionRepo,
			logger,
		),
		Delete: sessionUC.NewDeleteUseCase(
			infraContainer.SessionRepo,
			infraContainer.WhatsAppManager,
			logger,
		),
		Resolve: sessionUC.NewResolveUseCase(
			infraContainer.SessionRepo,
			logger,
		),
		SetProxy: sessionUC.NewSetProxyUseCase(
			infraContainer.SessionRepo,
			logger,
			validator,
		),
		AutoReconnect: sessionUC.NewAutoReconnectUseCase(
			infraContainer.SessionRepo,
			infraContainer.WhatsAppManager,
			logger,
		),
	}

	// Initialize WhatsApp use cases
	uc.whatsappUseCases = WhatsAppUseCases{
		GenerateQR: whatsappUC.NewGenerateQRUseCase(
			infraContainer.SessionRepo,
			infraContainer.WhatsAppManager,
			logger,
		),
		PairPhone: whatsappUC.NewPairPhoneUseCase(
			infraContainer.SessionRepo,
			infraContainer.WhatsAppManager,
			logger,
			validator,
		),
		SendMessage: whatsappUC.NewSendMessageUseCase(
			infraContainer.SessionRepo,
			infraContainer.WhatsAppManager,
			logger,
			validator,
		),
	}

	uc.isInitialized = true
	logger.Info("Use case container initialized successfully")
	return nil
}

// GetSessionUseCases returns session use cases
func (uc *useCaseContainer) GetSessionUseCases() SessionUseCases {
	return uc.sessionUseCases
}

// GetWhatsAppUseCases returns WhatsApp use cases
func (uc *useCaseContainer) GetWhatsAppUseCases() WhatsAppUseCases {
	return uc.whatsappUseCases
}
