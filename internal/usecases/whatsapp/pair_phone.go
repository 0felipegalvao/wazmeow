package whatsapp

import (
	"context"
	"regexp"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"
	"wazmeow/pkg/logger"
	"wazmeow/pkg/validator"
)

// PairPhoneUseCase handles phone number pairing for WhatsApp authentication
type PairPhoneUseCase struct {
	sessionRepo session.Repository
	waManager   whatsapp.Manager
	logger      logger.Logger
	validator   validator.Validator
}

// NewPairPhoneUseCase creates a new pair phone use case
func NewPairPhoneUseCase(sessionRepo session.Repository, waManager whatsapp.Manager, logger logger.Logger, validator validator.Validator) *PairPhoneUseCase {
	return &PairPhoneUseCase{
		sessionRepo: sessionRepo,
		waManager:   waManager,
		logger:      logger,
		validator:   validator,
	}
}

// PairPhoneRequest represents the request to pair with a phone number
type PairPhoneRequest struct {
	SessionID   session.SessionID `json:"session_id"`
	PhoneNumber string            `json:"phone_number" validate:"required,phone_number"`
}

// PairPhoneResponse represents the response from pairing with a phone number
type PairPhoneResponse struct {
	SessionID   session.SessionID `json:"session_id"`
	PhoneNumber string            `json:"phone_number"`
	Message     string            `json:"message"`
	Success     bool              `json:"success"`
}

// Execute pairs a session with a phone number
func (uc *PairPhoneUseCase) Execute(ctx context.Context, req PairPhoneRequest) (*PairPhoneResponse, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		uc.logger.ErrorWithError("validation failed for pair phone", err, logger.Fields{
			"session_id":   req.SessionID.String(),
			"phone_number": req.PhoneNumber,
		})
		return nil, err
	}

	// Additional phone number validation
	if !isValidPhoneNumber(req.PhoneNumber) {
		uc.logger.WarnWithFields("invalid phone number format", logger.Fields{
			"session_id":   req.SessionID.String(),
			"phone_number": req.PhoneNumber,
		})
		return nil, whatsapp.ErrInvalidPhoneNumber
	}

	// Get session from repository
	sess, err := uc.sessionRepo.GetByID(ctx, req.SessionID)
	if err != nil {
		uc.logger.ErrorWithError("failed to get session", err, logger.Fields{
			"session_id": req.SessionID.String(),
		})
		return nil, err
	}

	// Check if session is in a valid state for pairing
	if sess.IsConnected() {
		uc.logger.WarnWithFields("session already connected", logger.Fields{
			"session_id": sess.ID().String(),
			"status":     sess.Status().String(),
		})
		return nil, session.ErrSessionAlreadyConnected
	}

	// Get WhatsApp client
	waClient, err := uc.waManager.GetClient(sess.ID())
	if err != nil {
		// Create client if it doesn't exist
		waClient, err = uc.waManager.CreateClient(sess.ID())
		if err != nil {
			uc.logger.ErrorWithError("failed to create WhatsApp client", err, logger.Fields{
				"session_id": sess.ID().String(),
			})
			return nil, err
		}
	}

	// Check if already authenticated
	if waClient.IsAuthenticated() {
		uc.logger.InfoWithFields("session already authenticated", logger.Fields{
			"session_id": sess.ID().String(),
			"jid":        waClient.GetJID(),
		})
		return &PairPhoneResponse{
			SessionID:   sess.ID(),
			PhoneNumber: req.PhoneNumber,
			Message:     "Session already authenticated",
			Success:     true,
		}, nil
	}

	// Attempt to pair with phone number
	err = waClient.PairPhone(ctx, req.PhoneNumber)
	if err != nil {
		uc.logger.ErrorWithError("failed to pair with phone number", err, logger.Fields{
			"session_id":   sess.ID().String(),
			"phone_number": req.PhoneNumber,
		})
		return &PairPhoneResponse{
			SessionID:   sess.ID(),
			PhoneNumber: req.PhoneNumber,
			Message:     "Failed to pair with phone number",
			Success:     false,
		}, err
	}

	uc.logger.InfoWithFields("phone pairing initiated successfully", logger.Fields{
		"session_id":   sess.ID().String(),
		"phone_number": req.PhoneNumber,
	})

	return &PairPhoneResponse{
		SessionID:   sess.ID(),
		PhoneNumber: req.PhoneNumber,
		Message:     "Pairing code sent. Check your WhatsApp mobile app for the pairing code.",
		Success:     true,
	}, nil
}

// ValidatePhoneRequest represents the request to validate a phone number
type ValidatePhoneRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
}

// ValidatePhoneResponse represents the response from validating a phone number
type ValidatePhoneResponse struct {
	PhoneNumber string `json:"phone_number"`
	IsValid     bool   `json:"is_valid"`
	Message     string `json:"message"`
}

// ExecuteValidatePhone validates a phone number format
func (uc *PairPhoneUseCase) ExecuteValidatePhone(ctx context.Context, req ValidatePhoneRequest) (*ValidatePhoneResponse, error) {
	isValid := isValidPhoneNumber(req.PhoneNumber)
	
	response := &ValidatePhoneResponse{
		PhoneNumber: req.PhoneNumber,
		IsValid:     isValid,
	}

	if isValid {
		response.Message = "Phone number format is valid"
	} else {
		response.Message = "Phone number format is invalid. Must start with + followed by country code and number (10-15 digits total)"
	}

	uc.logger.InfoWithFields("phone number validation completed", logger.Fields{
		"phone_number": req.PhoneNumber,
		"is_valid":     isValid,
	})

	return response, nil
}

// isValidPhoneNumber validates phone number format
func isValidPhoneNumber(phoneNumber string) bool {
	// Phone number should start with + and have 10-15 digits total
	// Example: +1234567890, +551234567890
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{9,14}$`)
	return phoneRegex.MatchString(phoneNumber)
}

// FormatPhoneRequest represents the request to format a phone number
type FormatPhoneRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	CountryCode string `json:"country_code,omitempty"`
}

// FormatPhoneResponse represents the response from formatting a phone number
type FormatPhoneResponse struct {
	OriginalNumber  string `json:"original_number"`
	FormattedNumber string `json:"formatted_number"`
	IsValid         bool   `json:"is_valid"`
	Message         string `json:"message"`
}

// ExecuteFormatPhone formats a phone number to international format
func (uc *PairPhoneUseCase) ExecuteFormatPhone(ctx context.Context, req FormatPhoneRequest) (*FormatPhoneResponse, error) {
	originalNumber := req.PhoneNumber
	formattedNumber := formatPhoneNumber(req.PhoneNumber, req.CountryCode)
	isValid := isValidPhoneNumber(formattedNumber)

	response := &FormatPhoneResponse{
		OriginalNumber:  originalNumber,
		FormattedNumber: formattedNumber,
		IsValid:         isValid,
	}

	if isValid {
		response.Message = "Phone number formatted successfully"
	} else {
		response.Message = "Unable to format phone number to valid international format"
	}

	uc.logger.InfoWithFields("phone number formatting completed", logger.Fields{
		"original_number":  originalNumber,
		"formatted_number": formattedNumber,
		"country_code":     req.CountryCode,
		"is_valid":         isValid,
	})

	return response, nil
}

// formatPhoneNumber formats a phone number to international format
func formatPhoneNumber(phoneNumber, countryCode string) string {
	// Remove all non-digit characters except +
	phoneRegex := regexp.MustCompile(`[^\d+]`)
	cleaned := phoneRegex.ReplaceAllString(phoneNumber, "")

	// If already starts with +, return as is
	if len(cleaned) > 0 && cleaned[0] == '+' {
		return cleaned
	}

	// If starts with 00, replace with +
	if len(cleaned) >= 2 && cleaned[:2] == "00" {
		return "+" + cleaned[2:]
	}

	// If country code provided, prepend it
	if countryCode != "" {
		// Remove + from country code if present
		countryCode = regexp.MustCompile(`[^\d]`).ReplaceAllString(countryCode, "")
		
		// Remove leading 0 from phone number if present
		if len(cleaned) > 0 && cleaned[0] == '0' {
			cleaned = cleaned[1:]
		}
		
		return "+" + countryCode + cleaned
	}

	// Default: assume it's already in correct format, just add +
	if len(cleaned) > 0 && cleaned[0] != '+' {
		return "+" + cleaned
	}

	return cleaned
}
