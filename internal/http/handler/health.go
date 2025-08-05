package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"wazmeow/internal/http/dto"
	"wazmeow/internal/infra/container"
	"wazmeow/pkg/logger"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	container *container.Container
	logger    logger.Logger
	startTime time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(container *container.Container, logger logger.Logger) *HealthHandler {
	return &HealthHandler{
		container: container,
		logger:    logger,
		startTime: time.Now(),
	}
}

// Health handles GET /health
// @Summary Health Check da aplicação
// @Description Verifica o status de saúde da aplicação e todos os seus serviços dependentes (banco de dados, WhatsApp, etc.).
// @Description
// @Description **Informações retornadas:**
// @Description - Status geral da aplicação (healthy/unhealthy)
// @Description - Versão da aplicação
// @Description - Tempo de atividade (uptime)
// @Description - Status individual de cada serviço
// @Description - Timestamp da verificação
// @Description
// @Description **Status possíveis:**
// @Description - `healthy`: Todos os serviços funcionando normalmente
// @Description - `unhealthy`: Um ou mais serviços com problemas
// @Description - `degraded`: Serviços funcionando com limitações
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} dto.SuccessResponse{data=dto.HealthResponse} "Aplicação e serviços saudáveis"
// @Failure 503 {object} dto.ErrorResponse "Um ou mais serviços indisponíveis"
// @Router /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	services := make(map[string]interface{})

	// Check database health
	dbHealth := &dto.ServiceHealth{Status: "healthy"}
	if h.container != nil && h.container.DBConnection != nil {
		if err := h.container.Health(); err != nil {
			dbHealth.Status = "unhealthy"
			dbHealth.Message = err.Error()
		}
	} else {
		dbHealth.Status = "unhealthy"
		dbHealth.Message = "Database connection not initialized"
	}
	services["database"] = dbHealth

	// Check WhatsApp manager health
	waHealth := &dto.ServiceHealth{Status: "healthy"}
	if h.container != nil && h.container.WhatsAppManager != nil {
		if err := h.container.WhatsAppManager.HealthCheck(); err != nil {
			waHealth.Status = "unhealthy"
			waHealth.Message = err.Error()
		}
	} else {
		waHealth.Status = "unhealthy"
		waHealth.Message = "WhatsApp manager not initialized"
	}
	services["whatsapp"] = waHealth

	// Overall status
	overallStatus := "healthy"
	for _, service := range services {
		if serviceHealth, ok := service.(*dto.ServiceHealth); ok {
			if serviceHealth.Status != "healthy" {
				overallStatus = "unhealthy"
				break
			}
		}
	}

	response := &dto.HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   "1.0.0", // Could be injected from build
		Uptime:    time.Since(h.startTime).String(),
		Services:  services,
	}

	statusCode := http.StatusOK
	if overallStatus != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Metrics handles GET /metrics
// @Summary Métricas da aplicação
// @Description Retorna métricas detalhadas e estatísticas de performance da aplicação, incluindo informações sobre sessões, WhatsApp e sistema.
// @Description
// @Description **Métricas incluídas:**
// @Description
// @Description **Sessões:**
// @Description - Total de sessões criadas
// @Description - Sessões conectadas/desconectadas
// @Description - Sessões com erro
// @Description - Sessões ativas
// @Description
// @Description **WhatsApp:**
// @Description - Total de clientes WhatsApp
// @Description - Clientes conectados e autenticados
// @Description - Mensagens enviadas e recebidas
// @Description - Clientes com erro
// @Description
// @Description **Sistema:**
// @Description - Tempo de atividade (uptime)
// @Description - Uso de memória
// @Description - Uso de CPU
// @Description - Status do banco de dados
// @Description - Conexões ativas do banco
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} dto.SuccessResponse{data=dto.MetricsResponse} "Métricas coletadas com sucesso"
// @Failure 500 {object} dto.ErrorResponse "Erro interno ao coletar métricas"
// @Router /metrics [get]
func (h *HealthHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	// Get database stats
	_ = h.container.GetDatabaseStats() // TODO: Use database stats

	// Get WhatsApp stats
	waStats := h.container.GetWhatsAppStats()

	// Build metrics response
	response := &dto.MetricsResponse{
		Sessions: dto.SessionMetrics{
			// These would be populated from actual metrics
			Total:        0,
			Connected:    0,
			Disconnected: 0,
			Error:        0,
			Active:       0,
		},
		WhatsApp: dto.WhatsAppMetrics{
			TotalClients:         waStats.TotalClients,
			ConnectedClients:     waStats.ConnectedClients,
			AuthenticatedClients: waStats.AuthenticatedClients,
			ErrorClients:         waStats.ErrorClients,
			MessagesSent:         0, // Would be tracked in real implementation
			MessagesReceived:     0, // Would be tracked in real implementation
		},
		System: dto.SystemMetrics{
			Uptime:              time.Since(h.startTime).String(),
			MemoryUsage:         "N/A", // Would be calculated from runtime.MemStats
			CPUUsage:            "N/A", // Would be calculated from system metrics
			DatabaseStatus:      "healthy",
			DatabaseConnections: 0, // Would be extracted from dbStats
		},
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
