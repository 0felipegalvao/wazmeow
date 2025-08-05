package whats

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"
	"wazmeow/pkg/logger"
)

// Client implements whatsapp.Client using the real whatsmeow library
type Client struct {
	sessionID    session.SessionID
	eventHandler whatsapp.EventHandler
	logger       logger.Logger

	// Whatsmeow components
	container *sqlstore.Container
	device    *store.Device
	client    *whatsmeow.Client

	// QR code management
	currentQRCode    string
	currentQRBase64  string
	qrChannel        <-chan whatsmeow.QRChannelItem
	qrMonitoringDone chan bool
	isMonitoring     bool
}

// getDeviceForSession gets or creates a device for the given session
// Baseado no padrão do wuzapi para gerenciamento correto de múltiplas sessões
func getDeviceForSession(ctx context.Context, container *sqlstore.Container, sessionID session.SessionID, savedJID string, log logger.Logger) (*store.Device, error) {
	log.InfoWithFields("🔧 INICIANDO gerenciamento de dispositivo para sessão", logger.Fields{
		"session_id":    sessionID.String(),
		"saved_jid":     savedJID,
		"has_saved_jid": savedJID != "",
	})

	var device *store.Device
	var err error

	if savedJID != "" {
		log.InfoWithFields("📱 JID salvo encontrado - tentando recuperar dispositivo existente", logger.Fields{
			"session_id": sessionID.String(),
			"jid":        savedJID,
		})

		// Parse JID from database
		jid, ok := parseJID(savedJID)
		if ok {
			log.InfoWithFields("✅ JID parseado com sucesso - buscando dispositivo", logger.Fields{
				"session_id": sessionID.String(),
				"jid":        savedJID,
				"parsed_jid": jid.String(),
			})

			// Try to get existing device by JID
			device, err = container.GetDevice(ctx, jid)
			if err != nil {
				log.WarnWithFields("⚠️ FALHA ao recuperar dispositivo existente - criando novo", logger.Fields{
					"session_id": sessionID.String(),
					"jid":        savedJID,
					"error":      err.Error(),
				})
				device = container.NewDevice()
			} else {
				deviceIDStr := "nil"
				if device.ID != nil {
					deviceIDStr = device.ID.String()
				}
				log.InfoWithFields("🎉 SUCESSO: Dispositivo existente recuperado", logger.Fields{
					"session_id": sessionID.String(),
					"jid":        savedJID,
					"device_id":  deviceIDStr,
				})
			}
		} else {
			log.ErrorWithFields("💥 ERRO: Formato de JID inválido - criando novo dispositivo", logger.Fields{
				"session_id":  sessionID.String(),
				"invalid_jid": savedJID,
			})
			device = container.NewDevice()
		}
	} else {
		log.InfoWithFields("🆕 Nenhum JID salvo - criando novo dispositivo", logger.Fields{
			"session_id": sessionID.String(),
		})
		device = container.NewDevice()
	}

	deviceIDStr := "nil"
	if device.ID != nil {
		deviceIDStr = device.ID.String()
	}
	log.InfoWithFields("✅ DISPOSITIVO configurado para sessão", logger.Fields{
		"session_id":    sessionID.String(),
		"device_id":     deviceIDStr,
		"had_saved_jid": savedJID != "",
		"is_new_device": savedJID == "" || device.ID == nil,
	})

	return device, nil
}

// parseJID parses a JID string into types.JID
func parseJID(jidStr string) (types.JID, bool) {
	if jidStr == "" {
		return types.JID{}, false
	}

	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return types.JID{}, false
	}

	return jid, true
}

// NewClient creates a new WhatsApp client using whatsmeow with proper multi-session support
func NewClient(sessionID session.SessionID, container *sqlstore.Container, savedJID string, proxyURL string, log logger.Logger) (whatsapp.Client, error) {
	log.InfoWithFields("🏗️ CRIANDO novo cliente WhatsApp", logger.Fields{
		"session_id":    sessionID.String(),
		"saved_jid":     savedJID,
		"has_saved_jid": savedJID != "",
	})

	ctx := context.Background()

	// Get device for this session using saved JID (if any)
	log.InfoWithFields("🔧 Configurando dispositivo para sessão", logger.Fields{
		"session_id": sessionID.String(),
	})

	device, err := getDeviceForSession(ctx, container, sessionID, savedJID, log)
	if err != nil {
		log.ErrorWithFields("💥 FALHA CRÍTICA: Erro ao configurar dispositivo", logger.Fields{
			"session_id": sessionID.String(),
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get device for session: %w", err)
	}

	deviceIDStr := "nil"
	if device.ID != nil {
		deviceIDStr = device.ID.String()
	}
	log.InfoWithFields("✅ Dispositivo configurado - criando cliente whatsmeow", logger.Fields{
		"session_id": sessionID.String(),
		"device_id":  deviceIDStr,
	})

	// Create whatsmeow client
	client := whatsmeow.NewClient(device, nil)

	// Configure proxy if provided
	if proxyURL != "" {
		log.InfoWithFields("🌐 Configurando proxy para WhatsApp WebSocket", logger.Fields{
			"session_id": sessionID.String(),
			"proxy_url":  proxyURL,
		})

		parsedURL, err := url.Parse(proxyURL)
		if err != nil {
			log.ErrorWithFields("❌ Erro ao fazer parse da URL do proxy", logger.Fields{
				"session_id": sessionID.String(),
				"proxy_url":  proxyURL,
				"error":      err.Error(),
			})
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}

		// Configure proxy for WhatsApp WebSocket connections
		client.SetProxy(http.ProxyURL(parsedURL))

		log.InfoWithFields("✅ Proxy configurado com sucesso no cliente WhatsApp", logger.Fields{
			"session_id": sessionID.String(),
			"proxy_url":  proxyURL,
		})
	}

	whatsmeowClient := &Client{
		sessionID:        sessionID,
		logger:           log,
		container:        container,
		device:           device,
		client:           client,
		qrMonitoringDone: make(chan bool, 1),
		isMonitoring:     false,
	}

	// Set up event handler
	client.AddEventHandler(whatsmeowClient.handleEvent)

	log.InfoWithFields("🎉 CLIENTE WhatsApp criado com sucesso", logger.Fields{
		"session_id":   sessionID.String(),
		"device_id":    deviceIDStr,
		"has_store_id": client.Store.ID != nil,
	})

	return whatsmeowClient, nil
}

// handleEvent handles events from whatsmeow
func (c *Client) handleEvent(evt interface{}) {
	// Get event description and additional fields
	eventDesc, additionalFields := c.getEventDescription(evt)

	// Merge session info with additional fields
	logFields := logger.Fields{
		"session_id": c.sessionID.String(),
		"event_type": fmt.Sprintf("%T", evt),
	}
	for k, v := range additionalFields {
		logFields[k] = v
	}

	// Add payload to the main log fields for JSON file logging
	eventJSONPretty, err := json.MarshalIndent(evt, "", "  ")
	if err == nil {
		logFields["payload"] = json.RawMessage(eventJSONPretty)
	}

	// Log the event info with descriptive message (now includes payload)
	c.logger.InfoWithFields(eventDesc, logFields)

	switch v := evt.(type) {
	case *events.Connected:
		c.logger.InfoWithFields("🌐 WhatsApp CONECTADO", logger.Fields{
			"session_id":       c.sessionID.String(),
			"is_authenticated": c.client.Store.ID != nil,
		})

		// Trigger connected event if handler is set
		if c.eventHandler != nil {
			jid := ""
			if c.client.Store.ID != nil {
				jid = c.client.Store.ID.String()
			}
			c.eventHandler.OnConnected(c.sessionID, jid)
		}

	case *events.Disconnected:
		c.logger.WarnWithFields("🔌 WhatsApp DESCONECTADO", logger.Fields{
			"session_id":       c.sessionID.String(),
			"is_authenticated": c.client.Store.ID != nil,
		})

		// Trigger disconnected event if handler is set
		if c.eventHandler != nil {
			c.eventHandler.OnDisconnected(c.sessionID, "connection lost")
		}

	case *events.LoggedOut:
		c.logger.ErrorWithFields("🚪 WhatsApp LOGOUT - sessão invalidada", logger.Fields{
			"session_id": c.sessionID.String(),
			"reason":     v.Reason.String(),
		})

		// Clear authentication state
		c.currentQRCode = ""
		c.currentQRBase64 = ""

		// Trigger disconnected event if handler is set
		if c.eventHandler != nil {
			c.eventHandler.OnDisconnected(c.sessionID, fmt.Sprintf("logged out: %s", v.Reason.String()))
		}

	case *events.QR:
		c.logger.InfoWithFields("📱 QR codes recebidos via EVENTOS - exibindo automaticamente", logger.Fields{
			"session_id":  c.sessionID.String(),
			"codes_count": len(v.Codes),
		})

		// O whatsmeow já gerencia a renovação automática via canal QR
		// Aqui só exibimos o primeiro código como fallback se o canal não funcionar
		if len(v.Codes) > 0 {
			c.logger.InfoWithFields("📱 Exibindo primeiro QR code do evento", logger.Fields{
				"session_id":  c.sessionID.String(),
				"code_length": len(v.Codes[0]),
			})
			c.handleQRCodeEvent(v.Codes[0])

			// Trigger QR event if handler is set
			if c.eventHandler != nil {
				c.eventHandler.OnQRCode(c.sessionID, v.Codes[0])
			}
		}

	case *events.PairSuccess:
		c.logger.InfoWithFields("🎉 PAREAMENTO BEM-SUCEDIDO", logger.Fields{
			"session_id": c.sessionID.String(),
			"jid":        v.ID.String(),
		})

		// Clear QR code state since we're now authenticated
		c.currentQRCode = ""
		c.currentQRBase64 = ""

		// Trigger authentication event if handler is set
		if c.eventHandler != nil {
			c.eventHandler.OnAuthenticated(c.sessionID, v.ID.String())
		}

	case *events.StreamError:
		c.logger.ErrorWithFields("💥 ERRO de STREAM", logger.Fields{
			"session_id": c.sessionID.String(),
			"code":       v.Code,
		})

		// Trigger error event if handler is set
		if c.eventHandler != nil {
			c.eventHandler.OnError(c.sessionID, fmt.Errorf("stream error: code=%s", v.Code))
		}

	case *events.ConnectFailure:
		c.logger.ErrorWithFields("💥 FALHA na CONEXÃO", logger.Fields{
			"session_id": c.sessionID.String(),
			"reason":     v.Reason.String(),
		})

		// Trigger error event if handler is set
		if c.eventHandler != nil {
			c.eventHandler.OnError(c.sessionID, fmt.Errorf("connection failure: %s", v.Reason.String()))
		}

	default:
		// Handle other events as needed - payload already logged above
	}
}

// getEventDescription returns a descriptive message and additional fields for each event type
func (c *Client) getEventDescription(evt interface{}) (string, logger.Fields) {
	switch e := evt.(type) {
	case *events.Message:
		msgType := "texto"
		content := ""

		if e.Message.GetConversation() != "" {
			content = e.Message.GetConversation()
			if len(content) > 50 {
				content = content[:50] + "..."
			}
		} else if e.Message.GetImageMessage() != nil {
			msgType = "imagem"
			if caption := e.Message.GetImageMessage().GetCaption(); caption != "" {
				content = caption
			}
		} else if e.Message.GetVideoMessage() != nil {
			msgType = "vídeo"
			if caption := e.Message.GetVideoMessage().GetCaption(); caption != "" {
				content = caption
			}
		} else if e.Message.GetAudioMessage() != nil {
			msgType = "áudio"
		} else if e.Message.GetDocumentMessage() != nil {
			msgType = "documento"
			if fileName := e.Message.GetDocumentMessage().GetFileName(); fileName != "" {
				content = fileName
			}
		} else if e.Message.GetStickerMessage() != nil {
			msgType = "sticker"
		}

		direction := "📤 MENSAGEM ENVIADA"
		if !e.Info.IsFromMe {
			direction = "📥 MENSAGEM RECEBIDA"
		}

		fields := logger.Fields{
			"chat":         e.Info.Chat,
			"sender":       e.Info.Sender,
			"message_type": msgType,
			"message_id":   e.Info.ID,
			"is_group":     e.Info.IsGroup,
		}

		if content != "" {
			fields["content"] = content
		}

		return direction, fields

	case *events.Receipt:
		receiptType := "entregue"
		if e.Type != "" {
			receiptType = string(e.Type)
		}

		return "📋 CONFIRMAÇÃO DE ENTREGA", logger.Fields{
			"chat":        e.Chat,
			"sender":      e.Sender,
			"type":        receiptType,
			"message_ids": len(e.MessageIDs),
		}

	case *events.Connected:
		return "🌐 WHATSAPP CONECTADO", logger.Fields{}

	case *events.Disconnected:
		return "🔌 WHATSAPP DESCONECTADO", logger.Fields{
			"reason": "connection_lost",
		}

	case *events.LoggedOut:
		return "🚪 SESSÃO DESLOGADA", logger.Fields{
			"reason": e.Reason.String(),
		}

	case *events.QR:
		return "📱 QR CODE GERADO", logger.Fields{
			"codes": len(e.Codes),
		}

	case *events.PairSuccess:
		return "✅ PAREAMENTO CONCLUÍDO", logger.Fields{
			"jid": e.ID.String(),
		}

	case *events.OfflineSyncCompleted:
		return "🔄 SINCRONIZAÇÃO OFFLINE CONCLUÍDA", logger.Fields{
			"count": e.Count,
		}

	case *events.OfflineSyncPreview:
		return "👀 PRÉVIA DE SINCRONIZAÇÃO OFFLINE", logger.Fields{}

	case *events.PushName:
		return "👤 NOME ATUALIZADO", logger.Fields{
			"jid":       e.JID.String(),
			"push_name": e.NewPushName,
		}

	case *events.GroupInfo:
		action := "informações atualizadas"
		if e.JoinReason != "" {
			action = "membro adicionado"
		}

		return "👥 GRUPO - " + strings.ToUpper(action), logger.Fields{
			"group_jid": e.JID.String(),
			"name":      e.Name,
		}

	case *events.Presence:
		status := "online"
		if e.Unavailable {
			status = "offline"
		}

		return "👁️ PRESENÇA ATUALIZADA", logger.Fields{
			"jid":    e.From.String(),
			"status": status,
		}

	case *events.ChatPresence:
		return "💬 PRESENÇA NO CHAT", logger.Fields{
			"chat":  e.Chat.String(),
			"state": string(e.State),
		}

	case *events.HistorySync:
		syncType := "unknown"
		if e.Data.SyncType != nil {
			syncType = e.Data.SyncType.String()
		}
		return "📚 SINCRONIZAÇÃO DE HISTÓRICO", logger.Fields{
			"type":          syncType,
			"conversations": len(e.Data.Conversations),
		}

	default:
		return "📨 EVENTO WHATSAPP", logger.Fields{}
	}
}

// Connect establishes connection to WhatsApp
func (c *Client) Connect(ctx context.Context) (*whatsapp.ConnectionResult, error) {
	c.logger.InfoWithFields("🔄 INICIANDO conexão com WhatsApp", logger.Fields{
		"session_id":      c.sessionID.String(),
		"store_id_exists": c.client.Store.ID != nil,
		"is_connected":    c.client.IsConnected(),
	})

	result := &whatsapp.ConnectionResult{
		Status:    whatsapp.StatusConnected,
		Timestamp: time.Now(),
	}

	// Check if already logged in
	if c.client.Store.ID == nil {
		c.logger.InfoWithFields("📱 Nenhum ID armazenado - novo login necessário", logger.Fields{
			"session_id": c.sessionID.String(),
		})

		// No ID stored, new login - seguir padrão exato do código de referência
		c.logger.InfoWithFields("🔍 Obtendo canal QR...", logger.Fields{
			"session_id": c.sessionID.String(),
		})

		qrChan, err := c.client.GetQRChannel(context.Background())
		if err != nil {
			// This error means that we're already logged in, so ignore it.
			if !errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
				c.logger.ErrorWithFields("💥 FALHA: Não foi possível obter canal QR", logger.Fields{
					"session_id": c.sessionID.String(),
					"error":      err.Error(),
				})
				return nil, fmt.Errorf("failed to get QR channel: %w", err)
			}

			// Se já está logado, tratar como autenticado
			c.logger.InfoWithFields("✅ Já autenticado (ErrQRStoreContainsID)", logger.Fields{
				"session_id": c.sessionID.String(),
				"jid":        c.client.Store.ID.String(),
			})
			result.Status = whatsapp.StatusAuthenticated
			result.JID = c.client.Store.ID.String()
		} else {
			// Conectar DEPOIS de obter o canal QR
			c.logger.InfoWithFields("🌐 Conectando ao WhatsApp...", logger.Fields{
				"session_id": c.sessionID.String(),
			})

			err = c.client.Connect()
			if err != nil {
				c.logger.ErrorWithFields("💥 FALHA: Erro na conexão com WhatsApp", logger.Fields{
					"session_id": c.sessionID.String(),
					"error":      err.Error(),
				})
				return nil, fmt.Errorf("failed to connect: %w", err)
			}

			c.logger.InfoWithFields("✅ Conexão estabelecida - processando QR codes", logger.Fields{
				"session_id":   c.sessionID.String(),
				"is_connected": c.client.IsConnected(),
			})

			// Processar QR codes de forma assíncrona para não travar o endpoint
			go c.processQRChannel(qrChan)

			result.Status = whatsapp.StatusAuthenticating
		}
	} else {
		c.logger.InfoWithFields("✅ ID já existe - reconectando sessão autenticada", logger.Fields{
			"session_id": c.sessionID.String(),
			"jid":        c.client.Store.ID.String(),
		})

		// Already logged in, just connect
		result.Status = whatsapp.StatusAuthenticated
		result.JID = c.client.Store.ID.String()

		c.logger.InfoWithFields("🌐 Reconectando cliente autenticado...", logger.Fields{
			"session_id": c.sessionID.String(),
			"jid":        result.JID,
		})

		err := c.client.Connect()
		if err != nil {
			c.logger.ErrorWithFields("💥 FALHA: Erro na reconexão de cliente autenticado", logger.Fields{
				"session_id": c.sessionID.String(),
				"jid":        result.JID,
				"error":      err.Error(),
			})
			return nil, fmt.Errorf("failed to connect: %w", err)
		}

		c.logger.InfoWithFields("✅ Cliente autenticado reconectado com sucesso", logger.Fields{
			"session_id":   c.sessionID.String(),
			"jid":          result.JID,
			"is_connected": c.client.IsConnected(),
		})
	}

	c.logger.InfoWithFields("🎉 CONEXÃO CONCLUÍDA", logger.Fields{
		"session_id":       c.sessionID.String(),
		"status":           result.Status.String(),
		"jid":              result.JID,
		"is_connected":     c.client.IsConnected(),
		"is_authenticated": c.client.Store.ID != nil,
	})

	return result, nil
}

// Disconnect closes the WhatsApp connection
func (c *Client) Disconnect(ctx context.Context) error {
	c.logger.InfoWithFields("disconnecting from WhatsApp", logger.Fields{
		"session_id": c.sessionID.String(),
	})

	c.client.Disconnect()
	return nil
}

// IsConnected returns true if connected to WhatsApp
func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

// GetConnectionStatus returns the current connection status
func (c *Client) GetConnectionStatus() whatsapp.ConnectionStatus {
	if !c.client.IsConnected() {
		return whatsapp.StatusDisconnected
	}

	if c.client.Store.ID == nil {
		return whatsapp.StatusAuthenticating
	}

	return whatsapp.StatusAuthenticated
}

// GenerateQR generates a QR code for authentication
func (c *Client) GenerateQR(ctx context.Context) (string, error) {
	c.logger.InfoWithFields("🔍 SOLICITAÇÃO de geração de QR code", logger.Fields{
		"session_id":      c.sessionID.String(),
		"store_id_exists": c.client.Store.ID != nil,
		"is_monitoring":   c.isMonitoring,
		"has_current_qr":  c.currentQRCode != "",
		"is_connected":    c.client.IsConnected(),
	})

	if c.client.Store.ID != nil {
		c.logger.WarnWithFields("❌ ERRO: Tentativa de gerar QR para sessão já autenticada", logger.Fields{
			"session_id": c.sessionID.String(),
			"jid":        c.client.Store.ID.String(),
		})
		return "", fmt.Errorf("already authenticated")
	}

	c.logger.InfoWithFields("📱 Gerando QR code para autenticação", logger.Fields{
		"session_id":        c.sessionID.String(),
		"is_monitoring":     c.isMonitoring,
		"has_qr":            c.currentQRCode != "",
		"qr_channel_active": c.qrChannel != nil,
	})

	// Return the current QR code in base64 if available from continuous monitoring
	if c.currentQRBase64 != "" {
		c.logger.InfoWithFields("✅ Retornando QR code base64 atual do monitoramento contínuo", logger.Fields{
			"session_id":    c.sessionID.String(),
			"qr_length":     len(c.currentQRBase64),
			"is_monitoring": c.isMonitoring,
		})
		return c.currentQRBase64, nil
	}

	// If monitoring is active but no QR code yet, return placeholder
	if c.isMonitoring {
		c.logger.InfoWithFields("⏳ Monitoramento ativo mas QR ainda não disponível - retornando placeholder", logger.Fields{
			"session_id":        c.sessionID.String(),
			"qr_channel_active": c.qrChannel != nil,
		})
		return "qr-code-will-be-provided-via-continuous-monitoring", nil
	}

	// If no monitoring is active, return error
	c.logger.ErrorWithFields("💥 ERRO CRÍTICO: Monitoramento QR não está ativo", logger.Fields{
		"session_id":        c.sessionID.String(),
		"is_connected":      c.client.IsConnected(),
		"qr_channel_active": c.qrChannel != nil,
	})
	return "", fmt.Errorf("QR monitoring not active - please connect the session first")
}

// PairPhone pairs with a phone number
func (c *Client) PairPhone(ctx context.Context, phoneNumber string) error {
	if c.client.Store.ID != nil {
		return fmt.Errorf("already authenticated")
	}

	c.logger.InfoWithFields("pairing with phone", logger.Fields{
		"session_id":   c.sessionID.String(),
		"phone_number": phoneNumber,
	})

	code, err := c.client.PairPhone(ctx, phoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		return fmt.Errorf("failed to pair phone: %w", err)
	}

	c.logger.InfoWithFields("pairing code generated", logger.Fields{
		"session_id": c.sessionID.String(),
		"code":       code,
	})

	return nil
}

// IsAuthenticated returns true if authenticated with WhatsApp
func (c *Client) IsAuthenticated() bool {
	return c.client.Store.ID != nil
}

// GetSessionID returns the session ID
func (c *Client) GetSessionID() session.SessionID {
	return c.sessionID
}

// GetJID returns the WhatsApp JID
func (c *Client) GetJID() string {
	if c.client.Store.ID == nil {
		return ""
	}
	return c.client.Store.ID.String()
}

// GetDeviceInfo returns device information
func (c *Client) GetDeviceInfo() *whatsapp.DeviceInfo {
	return &whatsapp.DeviceInfo{
		Platform:     "linux",
		AppVersion:   "2.2412.54",
		DeviceModel:  "Desktop",
		OSVersion:    "0.1",
		Manufacturer: "WazMeow",
	}
}

// SendMessage sends a text message
func (c *Client) SendMessage(ctx context.Context, to, message string) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}

	// Parse recipient JID
	recipient, err := types.ParseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient JID: %w", err)
	}

	// Send message
	_, err = c.client.SendMessage(ctx, recipient, &waE2E.Message{
		Conversation: &message,
	})

	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	c.logger.InfoWithFields("message sent", logger.Fields{
		"session_id": c.sessionID.String(),
		"to":         to,
		"message":    message,
	})

	return nil
}

// SendImage sends an image message
func (c *Client) SendImage(ctx context.Context, to, imagePath, caption string) error {
	return fmt.Errorf("image sending not implemented yet")
}

// SendDocument sends a document message
func (c *Client) SendDocument(ctx context.Context, to, documentPath, filename string) error {
	return fmt.Errorf("document sending not implemented yet")
}

// SetEventHandler sets the event handler
func (c *Client) SetEventHandler(handler whatsapp.EventHandler) {
	c.eventHandler = handler
}

// RemoveEventHandler removes the event handler
func (c *Client) RemoveEventHandler() {
	c.eventHandler = nil
}

// Close closes the client
func (c *Client) Close() error {
	c.logger.InfoWithFields("Closing WhatsApp client", logger.Fields{
		"session_id": c.sessionID.String(),
	})

	// Stop QR monitoring if active
	c.stopQRMonitoring()

	// Disconnect from WhatsApp
	c.client.Disconnect()

	c.logger.InfoWithFields("WhatsApp client closed", logger.Fields{
		"session_id": c.sessionID.String(),
	})

	return nil
}

// processQRChannel processes QR channel synchronously (baseado no código de referência)
// Processa QR codes de forma síncrona seguindo o padrão exato do código que funciona
func (c *Client) processQRChannel(qrChan <-chan whatsmeow.QRChannelItem) {
	c.logger.InfoWithFields("🔄 Processando QR channel", logger.Fields{
		"session_id": c.sessionID.String(),
	})

	// Track if connection was established successfully
	connectionEstablished := false

	// Loop direto como no código de referência
	for evt := range qrChan {
		c.logger.InfoWithFields("📨 QR event received", logger.Fields{
			"session_id": c.sessionID.String(),
			"event":      evt.Event,
		})

		switch evt.Event {
		case "code":
			// Display QR code in terminal
			c.displayQRCodeInTerminal(evt.Code, "qr-code")

			// Store encoded/embedded base64 QR
			c.handleQRCodeEvent(evt.Code)

			c.logger.InfoWithFields("📱 QR code processado", logger.Fields{
				"session_id":  c.sessionID.String(),
				"code_length": len(evt.Code),
			})

		case "timeout":
			c.logger.WarnWithFields("⏰ QR timeout", logger.Fields{
				"session_id": c.sessionID.String(),
			})
			c.handleQRTimeoutEvent()
			return // Encerrar após timeout

		case "success":
			c.logger.InfoWithFields("🎉 QR pairing successful!", logger.Fields{
				"session_id": c.sessionID.String(),
			})
			connectionEstablished = true
			c.handleQRSuccessEvent()
			return // Encerrar após sucesso

		default:
			c.logger.InfoWithFields("📋 Other login event", logger.Fields{
				"session_id": c.sessionID.String(),
				"event":      evt.Event,
			})
		}
	}

	c.logger.InfoWithFields("🔚 QR channel closed", logger.Fields{
		"session_id":             c.sessionID.String(),
		"connection_established": connectionEstablished,
	})

	// Validate if channel was closed without establishing connection
	if !connectionEstablished {
		c.logger.WarnWithFields("⚠️ QR channel fechado sem estabelecer conexão - mudando status para disconnected", logger.Fields{
			"session_id": c.sessionID.String(),
		})
		c.handleQRChannelClosedWithoutConnection()
	}
}

// handleQRCodeEvent handles new QR code events - inicial ou renovação automática
// Baseado na implementação do zmeow QRCodeManager.handleQRCode
func (c *Client) handleQRCodeEvent(qrCode string) {
	isRenewal := c.currentQRCode != ""
	eventType := "initial"
	if isRenewal {
		eventType = "auto-renewal"
	}

	c.logger.InfoWithFields("🔄 Processing QR code event", logger.Fields{
		"session_id":  c.sessionID.String(),
		"code_length": len(qrCode),
		"type":        eventType,
		"is_renewal":  isRenewal,
	})

	// Store the current QR code
	c.currentQRCode = qrCode

	// Generate base64 encoded QR code
	image, err := qrcode.Encode(qrCode, qrcode.Medium, 256)
	if err != nil {
		c.logger.ErrorWithFields("❌ Failed to encode QR code", logger.Fields{
			"session_id": c.sessionID.String(),
			"error":      err.Error(),
			"type":       eventType,
		})
		return
	}

	base64QR := "data:image/png;base64," + base64.StdEncoding.EncodeToString(image)
	c.currentQRBase64 = base64QR

	// Display QR code in terminal (sempre exibir, mesmo renovações)
	c.displayQRCodeInTerminal(qrCode, eventType)

	// Trigger QR event if handler is set
	if c.eventHandler != nil {
		c.eventHandler.OnQRCode(c.sessionID, qrCode)
	}

	// TODO: Persistir no banco de dados
	// TODO: Enviar webhook se configurado

	c.logger.InfoWithFields("✅ QR code event processed successfully", logger.Fields{
		"session_id": c.sessionID.String(),
		"qr_length":  len(base64QR),
		"type":       eventType,
	})
}

// handleQRTimeoutEvent handles QR code timeout events
func (c *Client) handleQRTimeoutEvent() {
	c.logger.WarnWithFields("⏰ QR code timeout - limpando estado", logger.Fields{
		"session_id":    c.sessionID.String(),
		"had_qr_code":   c.currentQRCode != "",
		"had_qr_base64": c.currentQRBase64 != "",
	})

	// Clear QR code state
	previousQRCode := c.currentQRCode
	c.currentQRCode = ""
	c.currentQRBase64 = ""

	// Trigger timeout event if handler is set
	if c.eventHandler != nil {
		c.logger.InfoWithFields("📢 Disparando evento de timeout para handler", logger.Fields{
			"session_id": c.sessionID.String(),
		})
		c.eventHandler.OnError(c.sessionID, fmt.Errorf("QR code timeout"))
	}

	c.logger.InfoWithFields("🧹 QR code state cleared after timeout", logger.Fields{
		"session_id":         c.sessionID.String(),
		"previous_qr_length": len(previousQRCode),
	})
}

// handleQRSuccessEvent handles successful QR code authentication
func (c *Client) handleQRSuccessEvent() {
	c.logger.InfoWithFields("QR authentication successful", logger.Fields{
		"session_id": c.sessionID.String(),
	})

	// Clear QR code state
	c.currentQRCode = ""
	c.currentQRBase64 = ""

	// Get JID from authenticated client
	jid := ""
	if c.client.Store.ID != nil {
		jid = c.client.Store.ID.String()
	}

	// Trigger authentication event if handler is set
	// The event handler should save the JID to the database
	if c.eventHandler != nil && jid != "" {
		c.eventHandler.OnAuthenticated(c.sessionID, jid)
	}

	c.logger.InfoWithFields("QR success handled - session authenticated", logger.Fields{
		"session_id": c.sessionID.String(),
		"jid":        jid,
	})
}

// handleQRChannelClosedWithoutConnection handles when QR channel is closed without establishing connection
func (c *Client) handleQRChannelClosedWithoutConnection() {
	c.logger.WarnWithFields("🔌 QR channel fechado sem conexão estabelecida - limpando estado e notificando", logger.Fields{
		"session_id":    c.sessionID.String(),
		"had_qr_code":   c.currentQRCode != "",
		"had_qr_base64": c.currentQRBase64 != "",
	})

	// Clear QR code state
	c.currentQRCode = ""
	c.currentQRBase64 = ""

	// Mark monitoring as inactive
	c.isMonitoring = false

	// Trigger disconnection event if handler is set
	// This will change the session status from connecting to disconnected
	if c.eventHandler != nil {
		c.logger.InfoWithFields("📢 Disparando evento de desconexão para handler", logger.Fields{
			"session_id": c.sessionID.String(),
		})
		c.eventHandler.OnDisconnected(c.sessionID, "QR channel closed without connection")
	}

	c.logger.InfoWithFields("🔚 QR channel closure handled - session marked as disconnected", logger.Fields{
		"session_id": c.sessionID.String(),
	})
}

// stopQRMonitoring stops the QR monitoring gracefully
func (c *Client) stopQRMonitoring() {
	if c.isMonitoring {
		c.logger.InfoWithFields("Stopping QR monitoring", logger.Fields{
			"session_id": c.sessionID.String(),
		})

		select {
		case c.qrMonitoringDone <- true:
			c.logger.InfoWithFields("QR monitoring stop signal sent", logger.Fields{
				"session_id": c.sessionID.String(),
			})
		default:
			c.logger.WarnWithFields("QR monitoring stop signal channel full", logger.Fields{
				"session_id": c.sessionID.String(),
			})
		}
	}
}

// displayQRCodeInTerminal displays the QR code in the terminal and logs it
// Baseado na implementação do zmeow QRCodeManager.displayQRCodeInTerminal
func (c *Client) displayQRCodeInTerminal(qrCode string, eventType string) {
	c.logger.InfoWithFields("📺 Displaying QR code in terminal", logger.Fields{
		"session_id": c.sessionID.String(),
		"type":       eventType,
	})

	// Display QR code in terminal using ASCII art
	fmt.Println("\n" + strings.Repeat("=", 60))
	if eventType == "auto-renewal" {
		fmt.Printf("🔄 QR CODE RENOVADO AUTOMATICAMENTE - Sessão %s\n", c.sessionID.String())
		fmt.Println("📱 O QR code anterior expirou. Use este novo código:")
	} else {
		fmt.Printf("📱 QR CODE INICIAL - Sessão %s\n", c.sessionID.String())
		fmt.Println("🔗 Escaneie para conectar ao WhatsApp:")
	}

	fmt.Println("   1. Abra o WhatsApp no seu celular")
	fmt.Println("   2. Vá em Configurações > Aparelhos conectados")
	fmt.Println("   3. Toque em 'Conectar um aparelho'")
	fmt.Println("   4. Escaneie o QR Code abaixo")
	fmt.Println(strings.Repeat("=", 60))

	qrterminal.GenerateHalfBlock(qrCode, qrterminal.L, os.Stdout)

	fmt.Println(strings.Repeat("=", 60))
	if eventType == "auto-renewal" {
		fmt.Println("🔄 QR code renovado automaticamente (~60s)")
		fmt.Println("⏰ Este código expira em ~60 segundos e será renovado automaticamente")
	} else {
		fmt.Println("⏰ Este QR code expira em ~60 segundos")
		fmt.Println("🔄 Será renovado automaticamente se não for escaneado")
	}
	fmt.Println(strings.Repeat("=", 60) + "\n")

	c.logger.InfoWithFields("✅ QR code displayed in terminal", logger.Fields{
		"session_id": c.sessionID.String(),
		"type":       eventType,
	})
}
