package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/http/dto"
	sessionUC "wazmeow/internal/usecases/session"
	whatsappUC "wazmeow/internal/usecases/whatsapp"
	"wazmeow/pkg/errors"
	"wazmeow/pkg/logger"
	"wazmeow/pkg/validator"
)

// SessionHandler handles session-related HTTP requests
type SessionHandler struct {
	createUC     *sessionUC.CreateUseCase
	connectUC    *sessionUC.ConnectUseCase
	disconnectUC *sessionUC.DisconnectUseCase
	listUC       *sessionUC.ListUseCase
	deleteUC     *sessionUC.DeleteUseCase
	resolveUC    *sessionUC.ResolveUseCase
	setProxyUC   *sessionUC.SetProxyUseCase

	// WhatsApp use cases
	generateQRUC *whatsappUC.GenerateQRUseCase
	pairPhoneUC  *whatsappUC.PairPhoneUseCase

	logger    logger.Logger
	validator validator.Validator
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(
	createUC *sessionUC.CreateUseCase,
	connectUC *sessionUC.ConnectUseCase,
	disconnectUC *sessionUC.DisconnectUseCase,
	listUC *sessionUC.ListUseCase,
	deleteUC *sessionUC.DeleteUseCase,
	resolveUC *sessionUC.ResolveUseCase,
	setProxyUC *sessionUC.SetProxyUseCase,
	generateQRUC *whatsappUC.GenerateQRUseCase,
	pairPhoneUC *whatsappUC.PairPhoneUseCase,
	logger logger.Logger,
	validator validator.Validator,
) *SessionHandler {
	return &SessionHandler{
		createUC:     createUC,
		connectUC:    connectUC,
		disconnectUC: disconnectUC,
		listUC:       listUC,
		deleteUC:     deleteUC,
		resolveUC:    resolveUC,
		setProxyUC:   setProxyUC,
		generateQRUC: generateQRUC,
		pairPhoneUC:  pairPhoneUC,
		logger:       logger,
		validator:    validator,
	}
}

// CreateSession handles POST /sessions/add
// @Summary Criar nova sessão WhatsApp
// @Description Cria uma nova sessão WhatsApp com configuração opcional de proxy. A sessão é criada no estado 'disconnected' e pode ser conectada posteriormente.
// @Description
// @Description **Exemplos de uso:**
// @Description - Sessão simples: `{"name": "minha-sessao"}`
// @Description - Sessão com proxy HTTP: `{"name": "sessao-proxy", "proxy_host": "78.24.204.134", "proxy_port": 62122, "proxy_type": "http", "username": "user", "password": "pass"}`
// @Description - Sessão com proxy SOCKS5: `{"name": "sessao-socks5", "proxy_host": "78.24.204.134", "proxy_port": 62123, "proxy_type": "socks5"}`
// @Tags Sessions
// @Accept json
// @Produce json
// @Param request body dto.CreateSessionRequest true "Dados da sessão"
// @Success 201 {object} dto.SuccessResponse{data=dto.SessionResponse} "Sessão criada com sucesso"
// @Failure 400 {object} dto.ErrorResponse "Dados inválidos (nome muito curto, proxy inválido, etc.)"
// @Failure 409 {object} dto.ErrorResponse "Sessão com este nome já existe"
// @Failure 422 {object} dto.ErrorResponse "Dados válidos mas incompatíveis (ex: username sem password)"
// @Failure 500 {object} dto.ErrorResponse "Erro interno do servidor"
// @Security ApiKeyAuth
// @Router /sessions/add [post]
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Normalize and validate request
	req.Normalize()

	// Validate proxy parameters if provided
	if req.HasProxy() && !req.ProxyType.IsValid() {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid proxy type. Must be 'http' or 'socks5'", nil)
		return
	}

	// Execute use case
	ucReq := sessionUC.CreateRequest{Name: req.Name}
	result, err := h.createUC.Execute(r.Context(), ucReq)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Configure proxy if provided
	if req.HasProxy() {
		setProxyReq := sessionUC.SetProxyRequest{
			SessionID: result.Session.ID(),
			ProxyHost: req.ProxyHost,
			ProxyPort: req.ProxyPort,
			ProxyType: req.ProxyType.String(),
			Username:  req.Username,
			Password:  req.Password,
		}

		_, err := h.setProxyUC.Execute(r.Context(), setProxyReq)
		if err != nil {
			h.logger.ErrorWithError("failed to configure proxy during session creation", err, logger.Fields{
				"session_id": result.Session.ID().String(),
				"proxy_host": req.ProxyHost,
			})
			// Don't fail the session creation, just log the error
		} else {
			// Fetch updated session to include proxy configuration in response
			resolveReq := sessionUC.ResolveRequest{Identifier: session.SessionIdentifierFromID(result.Session.ID())}
			resolveResult, err := h.resolveUC.Execute(r.Context(), resolveReq)
			if err == nil {
				result.Session = resolveResult.Session
			}
		}
	}

	// Convert to HTTP response (this will include proxy_config if configured)
	response := dto.ToSessionResponse(result.Session)
	h.writeSuccessResponse(w, http.StatusCreated, "Session created successfully", response)
}

// ListSessions handles GET /sessions/list
// @Summary Listar sessões WhatsApp
// @Description Lista todas as sessões WhatsApp registradas no sistema com informações detalhadas incluindo status, configuração de proxy e timestamps.
// @Description
// @Description **Filtros disponíveis:**
// @Description - `status`: Filtra sessões por status (disconnected, connecting, connected)
// @Description
// @Description **Resposta inclui:**
// @Description - Lista de sessões com configuração completa
// @Description - Total de sessões encontradas
// @Description - Informações de proxy (se configurado)
// @Tags Sessions
// @Accept json
// @Produce json
// @Param status query string false "Filtrar por status da sessão" Enums(disconnected, connecting, connected)
// @Success 200 {object} dto.SuccessResponse{data=dto.SessionListResponse} "Lista de sessões recuperada com sucesso"
// @Failure 400 {object} dto.ErrorResponse "Parâmetros de filtro inválidos"
// @Failure 500 {object} dto.ErrorResponse "Erro interno do servidor"
// @Security ApiKeyAuth
// @Router /sessions/list [get]
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	statusStr := r.URL.Query().Get("status")

	var result *sessionUC.ListResponse
	var err error

	if statusStr != "" {
		// List by status
		status, parseErr := session.StatusFromString(statusStr)
		if parseErr != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid status parameter", parseErr)
			return
		}

		ucReq := sessionUC.ListByStatusRequest{
			Status: status,
			Limit:  0, // 0 means no limit - return all
			Offset: 0,
		}
		result, err = h.listUC.ExecuteByStatus(r.Context(), ucReq)
	} else {
		// List all
		ucReq := sessionUC.ListRequest{
			Limit:  0, // 0 means no limit - return all
			Offset: 0,
		}
		result, err = h.listUC.Execute(r.Context(), ucReq)
	}

	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Convert to HTTP response
	response := dto.ToSessionListResponse(result.Sessions, result.Total)
	h.writeSuccessResponse(w, http.StatusOK, "Sessions retrieved successfully", response)
}

// GetSession handles GET /sessions/{id}/info
// @Summary Obter detalhes da sessão
// @Description Retorna as informações detalhadas de uma sessão específica por ID ou nome, incluindo status completo
// @Tags Sessions
// @Accept json
// @Produce json
// @Param id path string true "ID da sessão (UUID) ou nome da sessão"
// @Success 200 {object} dto.SuccessResponse{data=dto.SessionResponse} "Detalhes da sessão"
// @Failure 400 {object} dto.ErrorResponse "Identificador da sessão inválido"
// @Failure 404 {object} dto.ErrorResponse "Sessão não encontrada"
// @Failure 500 {object} dto.ErrorResponse "Erro interno"
// @Security ApiKeyAuth
// @Router /sessions/{id}/info [get]
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	identifierStr := chi.URLParam(r, "id")

	// Resolve session using flexible identifier
	sess, err := h.resolveSessionByIdentifier(r, identifierStr)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Convert to HTTP response
	response := dto.ToSessionResponse(sess)
	h.writeSuccessResponse(w, http.StatusOK, "Session retrieved successfully", response)
}

// ConnectSession handles POST /sessions/{id}/connect
// @Summary Conectar sessão WhatsApp
// @Description Inicia o processo de conexão de uma sessão WhatsApp. Se a sessão não estiver autenticada, gera um QR Code para escaneamento.
// @Description
// @Description **Fluxo de conexão:**
// @Description 1. Sessão não autenticada: Retorna QR Code para escaneamento
// @Description 2. Sessão autenticada: Conecta diretamente ao WhatsApp
// @Description 3. Sessão já conectada: Retorna erro 409
// @Description
// @Description **Identificadores aceitos:**
// @Description - UUID da sessão: `4ee6195b-6a0f-4c85-a4ee-673ee15f14c8`
// @Description - Nome da sessão: `minha-sessao`
// @Tags Sessions
// @Accept json
// @Produce json
// @Param id path string true "ID da sessão (UUID) ou nome da sessão" example("minha-sessao")
// @Success 200 {object} dto.SuccessResponse{data=dto.ConnectSessionResponse} "Processo de conexão iniciado (QR Code gerado ou sessão conectada)"
// @Failure 400 {object} dto.ErrorResponse "Identificador da sessão inválido ou malformado"
// @Failure 404 {object} dto.ErrorResponse "Sessão não encontrada com o identificador fornecido"
// @Failure 409 {object} dto.ErrorResponse "Sessão já está conectada"
// @Failure 500 {object} dto.ErrorResponse "Erro interno do servidor ou falha na conexão WhatsApp"
// @Security ApiKeyAuth
// @Router /sessions/{id}/connect [post]
func (h *SessionHandler) ConnectSession(w http.ResponseWriter, r *http.Request) {
	identifierStr := chi.URLParam(r, "id")

	// Resolve session using flexible identifier
	sess, err := h.resolveSessionByIdentifier(r, identifierStr)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Execute use case with resolved session ID
	ucReq := sessionUC.ConnectRequest{SessionID: sess.ID()}
	result, err := h.connectUC.Execute(r.Context(), ucReq)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Convert to HTTP response
	response := &dto.ConnectSessionResponse{
		Session:   dto.ToSessionResponse(result.Session),
		QRCode:    result.QRCode,
		NeedsAuth: result.NeedsAuth,
		Message:   result.Message,
	}

	h.writeSuccessResponse(w, http.StatusOK, "Session connection processed", response)
}

// DeleteSession handles DELETE /sessions/{id}
// @Summary Deletar sessão WhatsApp
// @Description Deleta uma sessão WhatsApp específica por ID ou nome. Sempre força a deleção mesmo se conectada
// @Tags Sessions
// @Produce json
// @Param id path string true "ID da sessão (UUID) ou nome da sessão"
// @Success 200 {object} dto.SuccessResponse{data=dto.DeleteSessionResponse} "Sessão deletada"
// @Failure 400 {object} dto.ErrorResponse "Identificador da sessão inválido"
// @Failure 404 {object} dto.ErrorResponse "Sessão não encontrada"
// @Failure 500 {object} dto.ErrorResponse "Erro interno"
// @Security ApiKeyAuth
// @Router /sessions/{id} [delete]
func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	identifierStr := chi.URLParam(r, "id")

	// Resolve session using flexible identifier
	sess, err := h.resolveSessionByIdentifier(r, identifierStr)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Always force delete - no need for request body parsing
	// Execute use case with resolved session ID
	ucReq := sessionUC.DeleteRequest{
		SessionID: sess.ID(),
		Force:     true, // Always force delete
	}
	result, err := h.deleteUC.Execute(r.Context(), ucReq)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Convert to HTTP response
	response := &dto.DeleteSessionResponse{
		SessionID: result.SessionID.String(),
		Message:   result.Message,
	}

	h.writeSuccessResponse(w, http.StatusOK, "Session deleted", response)
}

// Helper methods

// resolveSessionByIdentifier resolves a session using the flexible identifier
func (h *SessionHandler) resolveSessionByIdentifier(r *http.Request, identifierStr string) (*session.Session, error) {
	// Validate input
	if identifierStr == "" {
		h.logger.WarnWithFields("empty session identifier provided", logger.Fields{
			"request_path": r.URL.Path,
		})
		return nil, session.ErrInvalidSessionIdentifier
	}

	// Create SessionIdentifier with automatic type detection
	identifier, err := session.NewSessionIdentifier(identifierStr)
	if err != nil {
		h.logger.ErrorWithError("invalid session identifier format", err, logger.Fields{
			"identifier":     identifierStr,
			"request_path":   r.URL.Path,
			"request_method": r.Method,
		})
		return nil, err
	}

	// Use resolve use case to get the session
	ucReq := sessionUC.ResolveRequest{Identifier: identifier}
	result, err := h.resolveUC.Execute(r.Context(), ucReq)
	if err != nil {
		h.logger.ErrorWithError("failed to resolve session", err, logger.Fields{
			"identifier":      identifierStr,
			"identifier_type": identifier.Type().String(),
			"request_path":    r.URL.Path,
			"request_method":  r.Method,
		})
		return nil, err
	}

	h.logger.InfoWithFields("session resolved successfully", logger.Fields{
		"session_id":      result.Session.ID().String(),
		"session_name":    result.Session.Name(),
		"identifier":      identifierStr,
		"identifier_type": result.IdentifierType,
		"request_path":    r.URL.Path,
	})

	return result.Session, nil
}

func (h *SessionHandler) writeSuccessResponse(w http.ResponseWriter, statusCode int, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := dto.NewSuccessResponse(message, data)
	json.NewEncoder(w).Encode(response)
}

func (h *SessionHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var details string
	if err != nil {
		details = err.Error()
	}

	response := dto.NewErrorResponse(message, "", details)
	json.NewEncoder(w).Encode(response)

	h.logger.ErrorWithError("HTTP error response", err, logger.Fields{
		"status_code": statusCode,
		"message":     message,
	})
}

func (h *SessionHandler) handleUseCaseError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		h.writeErrorResponse(w, appErr.GetHTTPStatus(), appErr.Message, err)
		return
	}

	// Handle domain errors
	switch err {
	case session.ErrSessionNotFound:
		h.writeErrorResponse(w, http.StatusNotFound, "Session not found", err)
	case session.ErrSessionAlreadyExists:
		h.writeErrorResponse(w, http.StatusConflict, "Session already exists", err)
	case session.ErrSessionAlreadyConnected:
		h.writeErrorResponse(w, http.StatusConflict, "Session already connected", err)
	case session.ErrSessionNotConnected:
		h.writeErrorResponse(w, http.StatusBadRequest, "Session not connected", err)
	case session.ErrSessionInvalidState:
		h.writeErrorResponse(w, http.StatusBadRequest, "Session in invalid state", err)
	default:
		h.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", err)
	}
}

// LogoutSession handles POST /sessions/{id}/logout
// @Summary Desconectar sessão (logout)
// @Description Desconecta a sessão do WhatsApp, encerrando a comunicação
// @Tags Sessions
// @Accept json
// @Produce json
// @Param id path string true "ID da sessão (UUID)"
// @Success 200 {object} dto.SuccessResponse{data=dto.DisconnectSessionResponse} "Sessão desconectada"
// @Failure 400 {object} dto.ErrorResponse "ID da sessão inválido"
// @Failure 404 {object} dto.ErrorResponse "Sessão não encontrada"
// @Failure 409 {object} dto.ErrorResponse "Sessão já desconectada"
// @Failure 500 {object} dto.ErrorResponse "Erro interno"
// @Security ApiKeyAuth
// @Router /sessions/{id}/logout [post]
func (h *SessionHandler) LogoutSession(w http.ResponseWriter, r *http.Request) {
	identifierStr := chi.URLParam(r, "id")

	// Resolve session using flexible identifier
	sess, err := h.resolveSessionByIdentifier(r, identifierStr)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Execute use case with resolved session ID
	ucReq := sessionUC.DisconnectRequest{SessionID: sess.ID()}
	result, err := h.disconnectUC.Execute(r.Context(), ucReq)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Convert to HTTP response
	response := &dto.DisconnectSessionResponse{
		Session: dto.ToSessionResponse(result.Session),
		Message: result.Message,
	}

	h.writeSuccessResponse(w, http.StatusOK, "Session disconnected", response)
}

// GenerateQR handles GET /sessions/{id}/qr
// @Summary Gerar QR Code para autenticação
// @Description Gera um QR Code para autenticação de uma sessão WhatsApp específica por ID ou nome
// @Tags Sessions
// @Accept json
// @Produce json
// @Param id path string true "ID da sessão (UUID) ou nome da sessão"
// @Success 200 {object} dto.SuccessResponse{data=dto.QRCodeResponse} "QR Code gerado"
// @Failure 400 {object} dto.ErrorResponse "Identificador da sessão inválido"
// @Failure 404 {object} dto.ErrorResponse "Sessão não encontrada"
// @Failure 409 {object} dto.ErrorResponse "Sessão já autenticada"
// @Failure 500 {object} dto.ErrorResponse "Erro interno"
// @Security ApiKeyAuth
// @Router /sessions/{id}/qr [get]
func (h *SessionHandler) GenerateQR(w http.ResponseWriter, r *http.Request) {
	identifierStr := chi.URLParam(r, "id")

	// Resolve session using flexible identifier
	sess, err := h.resolveSessionByIdentifier(r, identifierStr)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Execute use case with resolved session ID
	ucReq := whatsappUC.GenerateQRRequest{SessionID: sess.ID()}
	result, err := h.generateQRUC.Execute(r.Context(), ucReq)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Convert to HTTP response
	response := &dto.QRCodeResponse{
		SessionID: result.SessionID.String(),
		QRCode:    result.QRCode,
		Message:   result.Message,
	}

	h.writeSuccessResponse(w, http.StatusOK, "QR Code generated", response)
}

// PairPhone handles POST /sessions/{id}/pairphone
// @Summary Emparelhar telefone com sessão
// @Description Emparelha um telefone com a sessão WhatsApp por ID ou nome
// @Tags Sessions
// @Accept json
// @Produce json
// @Param id path string true "ID da sessão (UUID) ou nome da sessão"
// @Param request body dto.PairPhoneRequest true "Dados do telefone"
// @Success 200 {object} dto.SuccessResponse{data=dto.PairPhoneResponse} "Telefone emparelhado"
// @Failure 400 {object} dto.ErrorResponse "Dados inválidos"
// @Failure 404 {object} dto.ErrorResponse "Sessão não encontrada"
// @Failure 500 {object} dto.ErrorResponse "Erro interno"
// @Security ApiKeyAuth
// @Router /sessions/{id}/pairphone [post]
func (h *SessionHandler) PairPhone(w http.ResponseWriter, r *http.Request) {
	identifierStr := chi.URLParam(r, "id")

	// Resolve session using flexible identifier
	sess, err := h.resolveSessionByIdentifier(r, identifierStr)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	var req dto.PairPhoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Execute use case with resolved session ID
	ucReq := whatsappUC.PairPhoneRequest{
		SessionID:   sess.ID(),
		PhoneNumber: req.PhoneNumber,
	}
	result, err := h.pairPhoneUC.Execute(r.Context(), ucReq)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Convert to HTTP response
	response := &dto.PairPhoneResponse{
		SessionID:   result.SessionID.String(),
		PhoneNumber: result.PhoneNumber,
		Success:     result.Success,
		Message:     result.Message,
	}

	h.writeSuccessResponse(w, http.StatusOK, "Phone pairing processed", response)
}

// SetProxy handles POST /sessions/{id}/proxy/set
// @Summary Configurar proxy para sessão
// @Description Configura ou atualiza a configuração de proxy para uma sessão existente. O proxy será usado para todas as conexões WhatsApp desta sessão.
// @Description
// @Description **Tipos de proxy suportados:**
// @Description - HTTP: Proxy HTTP/HTTPS padrão
// @Description - SOCKS5: Proxy SOCKS5 com suporte a autenticação
// @Description
// @Description **Exemplos de configuração:**
// @Description - Proxy HTTP: `{"proxy_host": "78.24.204.134", "proxy_port": 62122, "proxy_type": "http"}`
// @Description - Proxy com autenticação: `{"proxy_host": "78.24.204.134", "proxy_port": 62122, "proxy_type": "http", "username": "user", "password": "pass"}`
// @Description - Proxy SOCKS5: `{"proxy_host": "78.24.204.134", "proxy_port": 62123, "proxy_type": "socks5", "username": "user", "password": "pass"}`
// @Tags Sessions
// @Accept json
// @Produce json
// @Param id path string true "ID da sessão (UUID) ou nome da sessão" example("minha-sessao")
// @Param request body dto.ProxySetRequest true "Configuração do proxy"
// @Success 200 {object} dto.SuccessResponse{data=dto.ProxySetResponse} "Proxy configurado com sucesso"
// @Failure 400 {object} dto.ErrorResponse "Dados de proxy inválidos (host, porta, tipo, etc.)"
// @Failure 404 {object} dto.ErrorResponse "Sessão não encontrada"
// @Failure 422 {object} dto.ErrorResponse "Configuração de proxy inválida (ex: username sem password)"
// @Failure 500 {object} dto.ErrorResponse "Erro interno do servidor"
// @Security ApiKeyAuth
// @Router /sessions/{id}/proxy/set [post]
func (h *SessionHandler) SetProxy(w http.ResponseWriter, r *http.Request) {
	identifierStr := chi.URLParam(r, "id")

	// Resolve session using flexible identifier
	sess, err := h.resolveSessionByIdentifier(r, identifierStr)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	var req dto.ProxySetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Se ProxyHost está vazio, remove o proxy
	if req.ProxyHost == "" {
		sess.ClearProxyURL()

		// Usar o use case para atualizar (passando host vazio)
		setProxyReq := sessionUC.SetProxyRequest{
			SessionID: sess.ID(),
			ProxyHost: "",
		}

		_, err := h.setProxyUC.Execute(r.Context(), setProxyReq)
		if err != nil {
			h.handleUseCaseError(w, err)
			return
		}

		response := &dto.ProxySetResponse{
			SessionID: sess.ID().String(),
			ProxyURL:  "",
			Success:   true,
			Message:   "Proxy removed successfully",
		}
		h.writeSuccessResponse(w, http.StatusOK, "Proxy removed", response)
		return
	}

	// Normalize request
	req.Normalize()

	// Se ProxyHost tem valor, configura o proxy usando o use case
	setProxyReq := sessionUC.SetProxyRequest{
		SessionID: sess.ID(),
		ProxyHost: req.ProxyHost,
		ProxyPort: req.ProxyPort,
		ProxyType: req.ProxyType.String(),
		Username:  req.Username,
		Password:  req.Password,
	}

	result, err := h.setProxyUC.Execute(r.Context(), setProxyReq)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	response := &dto.ProxySetResponse{
		SessionID: result.Session.ID().String(),
		ProxyURL:  result.Session.ProxyURL(),
		Success:   true,
		Message:   result.Message,
	}

	h.writeSuccessResponse(w, http.StatusOK, "Proxy configured", response)
}
