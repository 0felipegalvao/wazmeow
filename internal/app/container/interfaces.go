package container

import (
	"context"

	"wazmeow/internal/http/server"
	"wazmeow/internal/infra/config"
	sessionUC "wazmeow/internal/usecases/session"
	whatsappUC "wazmeow/internal/usecases/whatsapp"
	"wazmeow/pkg/logger"
)

// Container defines the interface for application containers
type Container interface {
	GetLogger() logger.Logger
	GetConfig() *config.Config
	Health() error
	Close() error
	IsInitialized() bool
}

// UseCaseContainer defines the interface for use case management
type UseCaseContainer interface {
	GetSessionUseCases() SessionUseCases
	GetWhatsAppUseCases() WhatsAppUseCases
}

// HTTPContainer defines the interface for HTTP layer management
type HTTPContainer interface {
	GetServerManager() *server.ServerManager
	GetServerInfo() server.ServerInfo
	StartServer(ctx context.Context) error
}

// SessionUseCases groups all session-related use cases
type SessionUseCases struct {
	Create        *sessionUC.CreateUseCase
	Connect       *sessionUC.ConnectUseCase
	Disconnect    *sessionUC.DisconnectUseCase
	List          *sessionUC.ListUseCase
	Delete        *sessionUC.DeleteUseCase
	Resolve       *sessionUC.ResolveUseCase
	SetProxy      *sessionUC.SetProxyUseCase
	AutoReconnect *sessionUC.AutoReconnectUseCase
}

// WhatsAppUseCases groups all WhatsApp-related use cases
type WhatsAppUseCases struct {
	GenerateQR  *whatsappUC.GenerateQRUseCase
	PairPhone   *whatsappUC.PairPhoneUseCase
	SendMessage *whatsappUC.SendMessageUseCase
}
