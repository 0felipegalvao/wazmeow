package dto

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"

	"wazmeow/pkg/validator"
)

// DTOValidator provides validation methods for DTOs
type DTOValidator struct {
	validator validator.Validator
}

// NewDTOValidator creates a new DTO validator
func NewDTOValidator(v validator.Validator) *DTOValidator {
	return &DTOValidator{validator: v}
}

// ValidateCreateSessionRequest validates a create session request
func (dv *DTOValidator) ValidateCreateSessionRequest(req *CreateSessionRequest) error {
	// Normalize first
	req.Normalize()

	// Basic validation using struct tags
	if err := dv.validator.Validate(req); err != nil {
		return err
	}

	// Custom validation logic
	if req.HasProxy() {
		if err := dv.validateProxyConfig(req.ProxyHost, req.ProxyPort, req.ProxyType, req.Username, req.Password); err != nil {
			return err
		}
	}

	return nil
}

// ValidateProxySetRequest validates a proxy set request
func (dv *DTOValidator) ValidateProxySetRequest(req *ProxySetRequest) error {
	// Normalize first
	req.Normalize()

	// Basic validation using struct tags
	if err := dv.validator.Validate(req); err != nil {
		return err
	}

	// Custom validation logic
	if req.HasProxy() {
		if err := dv.validateProxyConfig(req.ProxyHost, req.ProxyPort, req.ProxyType, req.Username, req.Password); err != nil {
			return err
		}
	}

	return nil
}

// ValidatePairPhoneRequest validates a pair phone request
func (dv *DTOValidator) ValidatePairPhoneRequest(req *PairPhoneRequest) error {
	// Basic validation using struct tags
	if err := dv.validator.Validate(req); err != nil {
		return err
	}

	// Custom phone number validation
	if err := dv.validatePhoneNumber(req.PhoneNumber); err != nil {
		return err
	}

	return nil
}

// ValidatePaginationRequest validates a pagination request
func (dv *DTOValidator) ValidatePaginationRequest(req *PaginationRequest) error {
	// Normalize first
	req.Normalize()

	// Basic validation using struct tags
	if err := dv.validator.Validate(req); err != nil {
		return err
	}

	return nil
}

// validateProxyConfig validates proxy configuration
func (dv *DTOValidator) validateProxyConfig(host string, port int, proxyType ProxyType, username, password string) error {
	// Validate host
	if err := dv.validateHost(host); err != nil {
		return NewValidationError("proxy_host", "invalid_host", host, "Invalid proxy host: "+err.Error())
	}

	// Validate port
	if port <= 0 || port > 65535 {
		return NewValidationError("proxy_port", "invalid_port", fmt.Sprintf("%d", port), "Proxy port must be between 1 and 65535")
	}

	// Validate proxy type
	if !proxyType.IsValid() {
		return NewValidationError("proxy_type", "invalid_type", proxyType.String(), "Proxy type must be 'http' or 'socks5'")
	}

	// Validate credentials if provided
	if username != "" && password == "" {
		return NewValidationError("password", "required_with_username", "", "Password is required when username is provided")
	}

	if password != "" && username == "" {
		return NewValidationError("username", "required_with_password", "", "Username is required when password is provided")
	}

	return nil
}

// validateHost validates a host (IP or hostname)
func (dv *DTOValidator) validateHost(host string) error {
	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	// Try to parse as IP first
	if ip := net.ParseIP(host); ip != nil {
		return nil
	}

	// Validate as hostname
	if err := dv.validateHostname(host); err != nil {
		return err
	}

	return nil
}

// validateHostname validates a hostname according to RFC standards
func (dv *DTOValidator) validateHostname(hostname string) error {
	if len(hostname) == 0 || len(hostname) > 253 {
		return fmt.Errorf("hostname length must be between 1 and 253 characters")
	}

	// Check for valid hostname pattern
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !hostnameRegex.MatchString(hostname) {
		return fmt.Errorf("invalid hostname format")
	}

	return nil
}

// validatePhoneNumber validates a phone number
func (dv *DTOValidator) validatePhoneNumber(phoneNumber string) error {
	if phoneNumber == "" {
		return NewValidationError("phone_number", "required", "", "Phone number is required")
	}

	// Remove spaces and special characters for validation
	cleaned := strings.ReplaceAll(phoneNumber, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	// Must start with + and contain only digits after that
	if !strings.HasPrefix(cleaned, "+") {
		return NewValidationError("phone_number", "invalid_format", phoneNumber, "Phone number must start with +")
	}

	// Check if the rest are digits
	digits := cleaned[1:]
	if len(digits) < 10 || len(digits) > 15 {
		return NewValidationError("phone_number", "invalid_length", phoneNumber, "Phone number must have between 10 and 15 digits")
	}

	for _, char := range digits {
		if char < '0' || char > '9' {
			return NewValidationError("phone_number", "invalid_characters", phoneNumber, "Phone number can only contain digits after +")
		}
	}

	return nil
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", ve.Field, ve.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, tag, value, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Tag:     tag,
		Value:   value,
		Message: message,
	}
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}

	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Error())
	}

	return strings.Join(messages, "; ")
}

// ToValidationErrorResponse converts validation errors to response
func (ve ValidationErrors) ToValidationErrorResponse() *ValidationErrorResponse {
	fields := make([]ValidationFieldError, len(ve))
	for i, err := range ve {
		fields[i] = ValidationFieldError(err)
	}

	return NewValidationErrorResponse(fields)
}

// ProxyURLValidator validates proxy URLs
type ProxyURLValidator struct{}

// NewProxyURLValidator creates a new proxy URL validator
func NewProxyURLValidator() *ProxyURLValidator {
	return &ProxyURLValidator{}
}

// Validate validates a proxy URL
func (puv *ProxyURLValidator) Validate(proxyURL string) error {
	if proxyURL == "" {
		return nil // Empty URL is valid (no proxy)
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL format: %w", err)
	}

	// Validate scheme
	switch parsedURL.Scheme {
	case "http", "https", "socks5":
		// Valid schemes
	default:
		return fmt.Errorf("unsupported proxy scheme: %s (supported: http, https, socks5)", parsedURL.Scheme)
	}

	// Validate host
	if parsedURL.Host == "" {
		return fmt.Errorf("proxy URL must include host")
	}

	// Validate port if specified
	if parsedURL.Port() != "" {
		// Port validation is handled by url.Parse
	}

	return nil
}

// SessionNameValidator validates session names
type SessionNameValidator struct{}

// NewSessionNameValidator creates a new session name validator
func NewSessionNameValidator() *SessionNameValidator {
	return &SessionNameValidator{}
}

// Validate validates a session name
func (snv *SessionNameValidator) Validate(name string) error {
	if name == "" {
		return NewValidationError("name", "required", "", "Session name is required")
	}

	if len(name) < 3 {
		return NewValidationError("name", "min_length", name, "Session name must be at least 3 characters long")
	}

	if len(name) > 50 {
		return NewValidationError("name", "max_length", name, "Session name must be at most 50 characters long")
	}

	// Check for valid characters (alphanumeric, spaces, hyphens, underscores)
	for _, char := range name {
		if !isValidSessionNameChar(char) {
			return NewValidationError("name", "invalid_characters", name, "Session name can only contain letters, numbers, spaces, hyphens, and underscores")
		}
	}

	return nil
}

// isValidSessionNameChar checks if a character is valid for session names
func isValidSessionNameChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == ' ' ||
		char == '-' ||
		char == '_'
}
