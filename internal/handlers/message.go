package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"wazmeow/internal/domain"
	"wazmeow/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// MessageHandler handles HTTP requests for message operations
type MessageHandler struct {
	multiSessionManager *services.MultiSessionManager
	mediaHelper         *MediaHelper
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(multiSessionManager *services.MultiSessionManager) *MessageHandler {
	return &MessageHandler{
		multiSessionManager: multiSessionManager,
		mediaHelper:         NewMediaHelper(),
	}
}

// SendTextMessage sends a text message
func (h *MessageHandler) SendTextMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionId")

	sessionID, err := domain.ParseSessionID(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var req SendTextMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Phone == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}
	if req.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	// Get session client
	client, err := h.multiSessionManager.GetClient(sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionIDStr).Msg("Failed to get session client")
		http.Error(w, "Session not found or not connected", http.StatusNotFound)
		return
	}

	// Parse recipient JID
	recipient, err := h.parsePhoneToJID(req.Phone)
	if err != nil {
		log.Error().Err(err).Str("phone", req.Phone).Msg("Failed to parse phone number")
		http.Error(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	// Generate message ID if not provided
	messageID := req.ID
	if messageID == "" {
		messageID = client.GenerateMessageID()
	}

	// Create text message
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(req.Message),
		},
	}

	// Send message
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.SendMessage(ctx, recipient, msg, whatsmeow.SendRequestExtra{ID: messageID})
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionIDStr).
			Str("phone", req.Phone).
			Msg("Failed to send text message")
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	// Create response
	response := MessageResponse{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
		Phone:     req.Phone,
		SessionID: sessionIDStr,
	}

	log.Info().
		Str("session_id", sessionIDStr).
		Str("phone", req.Phone).
		Str("message_id", resp.ID).
		Msg("Text message sent successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// parsePhoneToJID converts a phone number to WhatsApp JID
func (h *MessageHandler) parsePhoneToJID(phone string) (types.JID, error) {
	// Remove any non-numeric characters except +
	cleanPhone := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			cleanPhone += string(char)
		}
	}

	// Remove leading + if present
	if len(phone) > 0 && phone[0] == '+' {
		// Keep the clean phone as is
	}

	// Ensure we have a valid phone number
	if len(cleanPhone) < 10 {
		return types.JID{}, fmt.Errorf("phone number too short: %s", phone)
	}

	// Create JID for individual chat
	jid := types.JID{
		User:   cleanPhone,
		Server: "s.whatsapp.net",
	}

	return jid, nil
}

// SendImageMessage sends an image message
func (h *MessageHandler) SendImageMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionId")

	sessionID, err := domain.ParseSessionID(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var req SendImageMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Phone == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}
	if req.Image == "" {
		http.Error(w, "Image is required", http.StatusBadRequest)
		return
	}

	// Get session client
	client, err := h.multiSessionManager.GetClient(sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionIDStr).Msg("Failed to get session client")
		http.Error(w, "Session not found or not connected", http.StatusNotFound)
		return
	}

	// Parse recipient JID
	recipient, err := h.parsePhoneToJID(req.Phone)
	if err != nil {
		log.Error().Err(err).Str("phone", req.Phone).Msg("Failed to parse phone number")
		http.Error(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	// Generate message ID if not provided
	messageID := req.ID
	if messageID == "" {
		messageID = client.GenerateMessageID()
	}

	// Validate image format
	if err := h.mediaHelper.ValidateImageFormat(req.Image); err != nil {
		log.Error().Err(err).Msg("Invalid image format")
		http.Error(w, "Invalid image format", http.StatusBadRequest)
		return
	}

	// Decode data URL
	imageData, mimeType, err := h.mediaHelper.DecodeDataURL(req.Image)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode image data")
		http.Error(w, "Invalid image data", http.StatusBadRequest)
		return
	}

	// Generate thumbnail (optional - continue if fails)
	thumbnailData, err := h.mediaHelper.GenerateThumbnail(imageData)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate thumbnail, continuing without thumbnail")
		thumbnailData = []byte{} // Empty thumbnail
	}

	// Upload image to WhatsApp
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	uploaded, err := client.Upload(ctx, imageData, whatsmeow.MediaImage)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upload image")
		http.Error(w, fmt.Sprintf("Failed to upload image: %v", err), http.StatusInternalServerError)
		return
	}

	// Create image message
	msg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(imageData))),
			Caption:       proto.String(req.Caption),
			JPEGThumbnail: thumbnailData,
		},
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipient, msg, whatsmeow.SendRequestExtra{ID: messageID})
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionIDStr).
			Str("phone", req.Phone).
			Msg("Failed to send image message")
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	// Create response
	response := MessageResponse{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
		Phone:     req.Phone,
		SessionID: sessionIDStr,
	}

	log.Info().
		Str("session_id", sessionIDStr).
		Str("phone", req.Phone).
		Str("message_id", resp.ID).
		Msg("Image message sent successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SendAudioMessage sends an audio message
func (h *MessageHandler) SendAudioMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionId")

	sessionID, err := domain.ParseSessionID(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var req SendAudioMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Phone == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}
	if req.Audio == "" {
		http.Error(w, "Audio is required", http.StatusBadRequest)
		return
	}

	// Get session client
	client, err := h.multiSessionManager.GetClient(sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionIDStr).Msg("Failed to get session client")
		http.Error(w, "Session not found or not connected", http.StatusNotFound)
		return
	}

	// Parse recipient JID
	recipient, err := h.parsePhoneToJID(req.Phone)
	if err != nil {
		log.Error().Err(err).Str("phone", req.Phone).Msg("Failed to parse phone number")
		http.Error(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	// Validate audio format
	if err := h.mediaHelper.ValidateAudioFormat(req.Audio); err != nil {
		log.Error().Err(err).Msg("Invalid audio format")
		http.Error(w, "Invalid audio format: must be data:audio/ogg;base64", http.StatusBadRequest)
		return
	}

	// Decode data URL
	audioData, _, err := h.mediaHelper.DecodeDataURL(req.Audio)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode audio data")
		http.Error(w, "Invalid audio data", http.StatusBadRequest)
		return
	}

	// Generate message ID if not provided
	messageID := req.ID
	if messageID == "" {
		messageID = client.GenerateMessageID()
	}

	// Upload audio to WhatsApp
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	uploaded, err := client.Upload(ctx, audioData, whatsmeow.MediaAudio)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upload audio")
		http.Error(w, fmt.Sprintf("Failed to upload audio: %v", err), http.StatusInternalServerError)
		return
	}

	// Create audio message (PTT - Push to Talk)
	ptt := true
	msg := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(h.mediaHelper.GetAudioMimeType()),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(audioData))),
			PTT:           &ptt,
		},
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipient, msg, whatsmeow.SendRequestExtra{ID: messageID})
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionIDStr).
			Str("phone", req.Phone).
			Msg("Failed to send audio message")
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	// Create response
	response := MessageResponse{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
		Phone:     req.Phone,
		SessionID: sessionIDStr,
	}

	log.Info().
		Str("session_id", sessionIDStr).
		Str("phone", req.Phone).
		Str("message_id", resp.ID).
		Msg("Audio message sent successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SendVideoMessage sends a video message
func (h *MessageHandler) SendVideoMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionId")

	sessionID, err := domain.ParseSessionID(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var req SendVideoMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Phone == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}
	if req.Video == "" {
		http.Error(w, "Video is required", http.StatusBadRequest)
		return
	}

	// Get session client
	client, err := h.multiSessionManager.GetClient(sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionIDStr).Msg("Failed to get session client")
		http.Error(w, "Session not found or not connected", http.StatusNotFound)
		return
	}

	// Parse recipient JID
	recipient, err := h.parsePhoneToJID(req.Phone)
	if err != nil {
		log.Error().Err(err).Str("phone", req.Phone).Msg("Failed to parse phone number")
		http.Error(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	// Validate video format
	if err := h.mediaHelper.ValidateVideoFormat(req.Video); err != nil {
		log.Error().Err(err).Msg("Invalid video format")
		http.Error(w, "Invalid video format: must be data:video/*", http.StatusBadRequest)
		return
	}

	// Decode data URL
	videoData, mimeType, err := h.mediaHelper.DecodeDataURL(req.Video)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode video data")
		http.Error(w, "Invalid video data", http.StatusBadRequest)
		return
	}

	// Generate message ID if not provided
	messageID := req.ID
	if messageID == "" {
		messageID = client.GenerateMessageID()
	}

	// Upload video to WhatsApp
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second) // Longer timeout for videos
	defer cancel()

	uploaded, err := client.Upload(ctx, videoData, whatsmeow.MediaVideo)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upload video")
		http.Error(w, fmt.Sprintf("Failed to upload video: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate video thumbnail (placeholder for now)
	thumbnailData, _ := h.mediaHelper.GenerateVideoThumbnail(videoData)

	// Create video message
	msg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(videoData))),
			Caption:       proto.String(req.Caption),
			JPEGThumbnail: thumbnailData,
		},
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipient, msg, whatsmeow.SendRequestExtra{ID: messageID})
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionIDStr).
			Str("phone", req.Phone).
			Msg("Failed to send video message")
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	// Create response
	response := MessageResponse{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
		Phone:     req.Phone,
		SessionID: sessionIDStr,
	}

	log.Info().
		Str("session_id", sessionIDStr).
		Str("phone", req.Phone).
		Str("message_id", resp.ID).
		Msg("Video message sent successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SendDocumentMessage sends a document message
func (h *MessageHandler) SendDocumentMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionId")

	sessionID, err := domain.ParseSessionID(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var req SendDocumentMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Phone == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}
	if req.Document == "" {
		http.Error(w, "Document is required", http.StatusBadRequest)
		return
	}
	if req.Filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Get session client
	client, err := h.multiSessionManager.GetClient(sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionIDStr).Msg("Failed to get session client")
		http.Error(w, "Session not found or not connected", http.StatusNotFound)
		return
	}

	// Parse recipient JID
	recipient, err := h.parsePhoneToJID(req.Phone)
	if err != nil {
		log.Error().Err(err).Str("phone", req.Phone).Msg("Failed to parse phone number")
		http.Error(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	// Validate document format
	if err := h.mediaHelper.ValidateDocumentFormat(req.Document); err != nil {
		log.Error().Err(err).Msg("Invalid document format")
		http.Error(w, "Invalid document format: must be data:application/octet-stream;base64", http.StatusBadRequest)
		return
	}

	// Decode data URL
	documentData, _, err := h.mediaHelper.DecodeDataURL(req.Document)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode document data")
		http.Error(w, "Invalid document data", http.StatusBadRequest)
		return
	}

	// Generate message ID if not provided
	messageID := req.ID
	if messageID == "" {
		messageID = client.GenerateMessageID()
	}

	// Upload document to WhatsApp
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	uploaded, err := client.Upload(ctx, documentData, whatsmeow.MediaDocument)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upload document")
		http.Error(w, fmt.Sprintf("Failed to upload document: %v", err), http.StatusInternalServerError)
		return
	}

	// Determine MIME type
	mimeType := req.Mimetype
	if mimeType == "" {
		mimeType = h.mediaHelper.DetectMimeType(documentData)
	}

	// Create document message
	msg := &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(documentData))),
			FileName:      proto.String(req.Filename),
		},
	}

	// Send message
	resp, err := client.SendMessage(ctx, recipient, msg, whatsmeow.SendRequestExtra{ID: messageID})
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionIDStr).
			Str("phone", req.Phone).
			Msg("Failed to send document message")
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	// Create response
	response := MessageResponse{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
		Phone:     req.Phone,
		SessionID: sessionIDStr,
	}

	log.Info().
		Str("session_id", sessionIDStr).
		Str("phone", req.Phone).
		Str("message_id", resp.ID).
		Str("filename", req.Filename).
		Msg("Document message sent successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SendLocationMessage sends a location message
func (h *MessageHandler) SendLocationMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionId")

	sessionID, err := domain.ParseSessionID(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var req SendLocationMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Phone == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	// Validate coordinates
	if err := h.mediaHelper.IsValidCoordinate(req.Latitude, req.Longitude); err != nil {
		log.Error().Err(err).Msg("Invalid coordinates")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get session client
	client, err := h.multiSessionManager.GetClient(sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionIDStr).Msg("Failed to get session client")
		http.Error(w, "Session not found or not connected", http.StatusNotFound)
		return
	}

	// Parse recipient JID
	recipient, err := h.parsePhoneToJID(req.Phone)
	if err != nil {
		log.Error().Err(err).Str("phone", req.Phone).Msg("Failed to parse phone number")
		http.Error(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	// Generate message ID if not provided
	messageID := req.ID
	if messageID == "" {
		messageID = client.GenerateMessageID()
	}

	// Create location message
	msg := &waE2E.Message{
		LocationMessage: &waE2E.LocationMessage{
			DegreesLatitude:  proto.Float64(req.Latitude),
			DegreesLongitude: proto.Float64(req.Longitude),
			Name:             proto.String(req.Name),
			Address:          proto.String(req.Address),
		},
	}

	// Send message
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.SendMessage(ctx, recipient, msg, whatsmeow.SendRequestExtra{ID: messageID})
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionIDStr).
			Str("phone", req.Phone).
			Msg("Failed to send location message")
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	// Create response
	response := MessageResponse{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
		Phone:     req.Phone,
		SessionID: sessionIDStr,
	}

	log.Info().
		Str("session_id", sessionIDStr).
		Str("phone", req.Phone).
		Str("message_id", resp.ID).
		Float64("latitude", req.Latitude).
		Float64("longitude", req.Longitude).
		Msg("Location message sent successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SendContactMessage sends a contact message
func (h *MessageHandler) SendContactMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionId")

	sessionID, err := domain.ParseSessionID(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var req SendContactMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Phone == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}
	if req.ContactPhone == "" {
		http.Error(w, "Contact phone is required", http.StatusBadRequest)
		return
	}
	if req.ContactName == "" {
		http.Error(w, "Contact name is required", http.StatusBadRequest)
		return
	}

	// Get session client
	client, err := h.multiSessionManager.GetClient(sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionIDStr).Msg("Failed to get session client")
		http.Error(w, "Session not found or not connected", http.StatusNotFound)
		return
	}

	// Parse recipient JID
	recipient, err := h.parsePhoneToJID(req.Phone)
	if err != nil {
		log.Error().Err(err).Str("phone", req.Phone).Msg("Failed to parse phone number")
		http.Error(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	// Generate message ID if not provided
	messageID := req.ID
	if messageID == "" {
		messageID = client.GenerateMessageID()
	}

	// Create vCard
	vcard := h.mediaHelper.FormatVCard(req.ContactName, req.ContactPhone)

	// Create contact message
	msg := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: proto.String(req.ContactName),
			Vcard:       proto.String(vcard),
		},
	}

	// Send message
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.SendMessage(ctx, recipient, msg, whatsmeow.SendRequestExtra{ID: messageID})
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionIDStr).
			Str("phone", req.Phone).
			Msg("Failed to send contact message")
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	// Create response
	response := MessageResponse{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
		Phone:     req.Phone,
		SessionID: sessionIDStr,
	}

	log.Info().
		Str("session_id", sessionIDStr).
		Str("phone", req.Phone).
		Str("message_id", resp.ID).
		Str("contact_name", req.ContactName).
		Str("contact_phone", req.ContactPhone).
		Msg("Contact message sent successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
