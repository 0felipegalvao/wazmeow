package dto

import (
	"time"
)

// ResponseStatus represents the status of an API response
// @Description Status da resposta da API
// @Enum success error
type ResponseStatus string

const (
	// StatusSuccess indicates a successful operation
	StatusSuccess ResponseStatus = "success"
	// StatusError indicates a failed operation
	StatusError ResponseStatus = "error"
)

// String returns the string representation of ResponseStatus
func (rs ResponseStatus) String() string {
	return string(rs)
}

// SuccessResponse represents a generic success response
// @Description Resposta de sucesso padrão da API
type SuccessResponse struct {
	Success bool   `json:"success" example:"true" description:"Indica se a operação foi bem-sucedida"`
	Status  string `json:"status" example:"success" description:"Status da operação (sempre 'success' para respostas de sucesso)"`
	Message string `json:"message" example:"Operação realizada com sucesso" description:"Mensagem descritiva do resultado"`
	Data    any    `json:"data,omitempty" description:"Dados retornados pela operação (opcional)"`
}

// TypedSuccessResponse represents a typed success response for better type safety
// @Description Resposta de sucesso tipada da API
type TypedSuccessResponse[T any] struct {
	Success bool   `json:"success" example:"true" description:"Indica se a operação foi bem-sucedida"`
	Status  string `json:"status" example:"success" description:"Status da operação"`
	Message string `json:"message" example:"Operação realizada com sucesso" description:"Mensagem descritiva do resultado"`
	Data    T      `json:"data,omitempty" description:"Dados retornados pela operação (opcional)"`
}

// ErrorResponse represents a generic error response
// @Description Resposta de erro padrão da API
type ErrorResponse struct {
	Success   bool           `json:"success" example:"false" description:"Sempre false para respostas de erro"`
	Status    string         `json:"status" example:"error" description:"Status da operação (sempre 'error' para respostas de erro)"`
	Error     string         `json:"error" example:"Erro interno do servidor" description:"Mensagem de erro principal"`
	Code      string         `json:"code,omitempty" example:"INTERNAL_ERROR" description:"Código padronizado do erro para identificação programática"`
	Details   string         `json:"details,omitempty" example:"Detalhes técnicos do erro" description:"Detalhes técnicos adicionais do erro"`
	Context   map[string]any `json:"context,omitempty" description:"Contexto adicional do erro (request_id, session_id, etc.)"`
	Timestamp time.Time      `json:"timestamp" example:"2024-01-01T12:00:00Z" description:"Timestamp automático de quando o erro ocorreu"`
}

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Success bool                   `json:"success"`
	Error   string                 `json:"error"`
	Code    string                 `json:"code"`
	Fields  []ValidationFieldError `json:"fields"`
}

// ValidationFieldError represents a field validation error
type ValidationFieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// PaginationRequest represents pagination parameters
// @Description Parâmetros de paginação para listagens
type PaginationRequest struct {
	Limit  int `json:"limit" query:"limit" validate:"min=1,max=100" example:"10" description:"Número máximo de itens por página (1-100)"`
	Offset int `json:"offset" query:"offset" validate:"min=0" example:"0" description:"Número de itens a pular (para paginação)"`
	Page   int `json:"page" query:"page" validate:"min=1" example:"1" description:"Número da página (alternativa ao offset)"`
}

// PaginationResponse represents pagination metadata
// @Description Metadados de paginação para listagens
type PaginationResponse struct {
	Total       int  `json:"total" example:"100" description:"Total de itens disponíveis"`
	Limit       int  `json:"limit" example:"10" description:"Número máximo de itens por página"`
	Offset      int  `json:"offset" example:"0" description:"Número de itens pulados"`
	Page        int  `json:"page" example:"1" description:"Página atual"`
	Pages       int  `json:"pages" example:"10" description:"Total de páginas"`
	HasNext     bool `json:"has_next" example:"true" description:"Indica se há próxima página"`
	HasPrevious bool `json:"has_previous" example:"false" description:"Indica se há página anterior"`
}

// HealthResponse represents the health check response
// @Description Resposta do health check da aplicação
type HealthResponse struct {
	Status    string                 `json:"status" example:"healthy" description:"Status geral da aplicação"`
	Timestamp time.Time              `json:"timestamp" example:"2024-01-01T12:00:00Z" description:"Timestamp da verificação"`
	Version   string                 `json:"version" example:"1.0.0" description:"Versão da aplicação"`
	Uptime    string                 `json:"uptime" example:"2h30m45s" description:"Tempo de atividade da aplicação"`
	Services  map[string]interface{} `json:"services" description:"Status dos serviços individuais"`
}

// HealthStatus represents the health status of a service
// @Description Status de saúde de um serviço
// @Enum healthy unhealthy degraded unknown
type HealthStatus string

const (
	// HealthStatusHealthy indicates the service is healthy
	HealthStatusHealthy HealthStatus = "healthy"
	// HealthStatusUnhealthy indicates the service is unhealthy
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	// HealthStatusDegraded indicates the service is degraded
	HealthStatusDegraded HealthStatus = "degraded"
	// HealthStatusUnknown indicates the service status is unknown
	HealthStatusUnknown HealthStatus = "unknown"
)

// String returns the string representation of HealthStatus
func (hs HealthStatus) String() string {
	return string(hs)
}

// ServiceHealth represents the health status of a service
// @Description Status de saúde de um serviço individual
type ServiceHealth struct {
	Status    HealthStatus   `json:"status" example:"healthy" description:"Status atual do serviço (healthy, unhealthy, degraded, unknown)"`
	Message   string         `json:"message,omitempty" example:"Service is running normally" description:"Mensagem descritiva do status atual"`
	Details   map[string]any `json:"details,omitempty" description:"Detalhes específicos do serviço (versão, configuração, etc.)"`
	Timestamp time.Time      `json:"timestamp" example:"2024-01-01T12:00:00Z" description:"Timestamp da última verificação de saúde"`
	Metrics   map[string]any `json:"metrics,omitempty" description:"Métricas específicas do serviço (latência, throughput, etc.)"`
}

// MetricsResponse represents application metrics
// @Description Métricas completas da aplicação
type MetricsResponse struct {
	Sessions  SessionMetrics  `json:"sessions" description:"Métricas das sessões"`
	WhatsApp  WhatsAppMetrics `json:"whatsapp" description:"Métricas do WhatsApp"`
	System    SystemMetrics   `json:"system" description:"Métricas do sistema"`
	Timestamp time.Time       `json:"timestamp" example:"2024-01-01T12:00:00Z" description:"Timestamp da coleta das métricas"`
}

// SessionMetrics represents session-related metrics
// @Description Métricas relacionadas às sessões WhatsApp
type SessionMetrics struct {
	Total        int `json:"total" example:"10" description:"Total de sessões"`
	Connected    int `json:"connected" example:"5" description:"Sessões conectadas"`
	Disconnected int `json:"disconnected" example:"3" description:"Sessões desconectadas"`
	Error        int `json:"error" example:"1" description:"Sessões com erro"`
	Active       int `json:"active" example:"4" description:"Sessões ativas"`
}

// WhatsAppMetrics represents WhatsApp-related metrics
// @Description Métricas relacionadas ao WhatsApp
type WhatsAppMetrics struct {
	TotalClients         int `json:"total_clients" example:"5" description:"Total de clientes WhatsApp"`
	ConnectedClients     int `json:"connected_clients" example:"3" description:"Clientes conectados"`
	AuthenticatedClients int `json:"authenticated_clients" example:"2" description:"Clientes autenticados"`
	ErrorClients         int `json:"error_clients" example:"1" description:"Clientes com erro"`
	MessagesSent         int `json:"messages_sent" example:"150" description:"Total de mensagens enviadas"`
	MessagesReceived     int `json:"messages_received" example:"75" description:"Total de mensagens recebidas"`
}

// SystemMetrics represents system-related metrics
// @Description Métricas relacionadas ao sistema
type SystemMetrics struct {
	Uptime              string `json:"uptime" example:"2h30m45s" description:"Tempo de atividade do sistema"`
	MemoryUsage         string `json:"memory_usage" example:"256MB" description:"Uso de memória"`
	CPUUsage            string `json:"cpu_usage" example:"15%" description:"Uso de CPU"`
	DatabaseStatus      string `json:"database_status" example:"healthy" description:"Status do banco de dados"`
	DatabaseConnections int    `json:"database_connections" example:"5" description:"Número de conexões ativas no banco"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(message string, data interface{}) *SuccessResponse {
	return &SuccessResponse{
		Success: true,
		Status:  StatusSuccess.String(),
		Message: message,
		Data:    data,
	}
}

// NewTypedSuccessResponse creates a new typed success response
func NewTypedSuccessResponse[T any](message string, data T) *TypedSuccessResponse[T] {
	return &TypedSuccessResponse[T]{
		Success: true,
		Status:  StatusSuccess.String(),
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(error, code, details string) *ErrorResponse {
	return &ErrorResponse{
		Success:   false,
		Status:    StatusError.String(),
		Error:     error,
		Code:      code,
		Details:   details,
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// NewErrorResponseWithContext creates a new error response with context
func NewErrorResponseWithContext(error, code, details string, context map[string]interface{}) *ErrorResponse {
	return &ErrorResponse{
		Success:   false,
		Status:    StatusError.String(),
		Error:     error,
		Code:      code,
		Details:   details,
		Context:   context,
		Timestamp: time.Now(),
	}
}

// NewValidationErrorResponse creates a new validation error response
func NewValidationErrorResponse(fields []ValidationFieldError) *ValidationErrorResponse {
	return &ValidationErrorResponse{
		Success: false,
		Error:   "Validation failed",
		Code:    "VALIDATION_ERROR",
		Fields:  fields,
	}
}

// Normalize normalizes pagination request parameters
func (pr *PaginationRequest) Normalize() {
	// Set default values
	if pr.Limit <= 0 {
		pr.Limit = 10
	}
	if pr.Limit > 100 {
		pr.Limit = 100
	}
	if pr.Offset < 0 {
		pr.Offset = 0
	}
	if pr.Page <= 0 {
		pr.Page = 1
	}

	// Convert page to offset if page is provided
	if pr.Page > 1 && pr.Offset == 0 {
		pr.Offset = (pr.Page - 1) * pr.Limit
	}
}

// GetPage returns the current page number
func (pr *PaginationRequest) GetPage() int {
	if pr.Page > 0 {
		return pr.Page
	}
	return (pr.Offset / pr.Limit) + 1
}

// CalculatePages calculates the number of pages for pagination
func (p *PaginationResponse) CalculatePages() {
	if p.Limit > 0 {
		p.Pages = (p.Total + p.Limit - 1) / p.Limit
		p.Page = (p.Offset / p.Limit) + 1
		p.HasNext = p.Page < p.Pages
		p.HasPrevious = p.Page > 1
	}
}

// NewPaginationResponse creates a new pagination response
func NewPaginationResponse(total, limit, offset int) *PaginationResponse {
	pagination := &PaginationResponse{
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	pagination.CalculatePages()
	return pagination
}

// NewPaginationResponseFromRequest creates pagination response from request
func NewPaginationResponseFromRequest(req *PaginationRequest, total int) *PaginationResponse {
	req.Normalize()
	return NewPaginationResponse(total, req.Limit, req.Offset)
}

// NewServiceHealth creates a new service health instance
func NewServiceHealth(status HealthStatus, message string) *ServiceHealth {
	return &ServiceHealth{
		Status:    status,
		Message:   message,
		Details:   make(map[string]interface{}),
		Timestamp: time.Now(),
		Metrics:   make(map[string]interface{}),
	}
}

// NewHealthyService creates a healthy service health instance
func NewHealthyService(message string) *ServiceHealth {
	return NewServiceHealth(HealthStatusHealthy, message)
}

// NewUnhealthyService creates an unhealthy service health instance
func NewUnhealthyService(message string) *ServiceHealth {
	return NewServiceHealth(HealthStatusUnhealthy, message)
}

// NewDegradedService creates a degraded service health instance
func NewDegradedService(message string) *ServiceHealth {
	return NewServiceHealth(HealthStatusDegraded, message)
}

// AddDetail adds a detail to the service health
func (sh *ServiceHealth) AddDetail(key string, value interface{}) {
	if sh.Details == nil {
		sh.Details = make(map[string]interface{})
	}
	sh.Details[key] = value
}

// AddMetric adds a metric to the service health
func (sh *ServiceHealth) AddMetric(key string, value interface{}) {
	if sh.Metrics == nil {
		sh.Metrics = make(map[string]interface{})
	}
	sh.Metrics[key] = value
}

// IsHealthy returns true if the service is healthy
func (sh *ServiceHealth) IsHealthy() bool {
	return sh.Status == HealthStatusHealthy
}

// Factory Methods for Common DTOs

// CreateSuccessResponseWithData creates a success response with typed data
func CreateSuccessResponseWithData[T any](message string, data T) *TypedSuccessResponse[T] {
	return NewTypedSuccessResponse(message, data)
}

// CreateErrorResponseFromError creates an error response from a Go error
func CreateErrorResponseFromError(err error, code string) *ErrorResponse {
	return NewErrorResponse(err.Error(), code, "")
}

// CreateValidationErrorFromFields creates a validation error response from field errors
func CreateValidationErrorFromFields(fields []ValidationFieldError) *ValidationErrorResponse {
	return NewValidationErrorResponse(fields)
}

// CreatePaginatedResponse creates a paginated response with typed data
func CreatePaginatedResponse[T any](data []T, pagination *PaginationResponse) *TypedSuccessResponse[map[string]interface{}] {
	responseData := map[string]interface{}{
		"items":      data,
		"pagination": pagination,
	}
	return NewTypedSuccessResponse("Data retrieved successfully", responseData)
}

// CreateHealthResponse creates a health response
func CreateHealthResponse(status HealthStatus, version, uptime string, services map[string]*ServiceHealth) *HealthResponse {
	servicesMap := make(map[string]interface{})
	for name, health := range services {
		servicesMap[name] = health
	}

	return &HealthResponse{
		Status:    status.String(),
		Timestamp: time.Now(),
		Version:   version,
		Uptime:    uptime,
		Services:  servicesMap,
	}
}

// CreateMetricsResponse creates a metrics response
func CreateMetricsResponse(sessions SessionMetrics, whatsapp WhatsAppMetrics, system SystemMetrics) *MetricsResponse {
	return &MetricsResponse{
		Sessions:  sessions,
		WhatsApp:  whatsapp,
		System:    system,
		Timestamp: time.Now(),
	}
}
