package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

// FormatWhatsAppJID formats a phone number to WhatsApp JID format
func FormatWhatsAppJID(phone string) string {
	// Remove any non-numeric characters except +
	formatted := strings.ReplaceAll(phone, " ", "")
	formatted = strings.ReplaceAll(formatted, "-", "")
	formatted = strings.ReplaceAll(formatted, "(", "")
	formatted = strings.ReplaceAll(formatted, ")", "")
	formatted = strings.ReplaceAll(formatted, ".", "")

	// Add @s.whatsapp.net if not present
	if !strings.Contains(formatted, "@") {
		// Remove + if present at the beginning
		formatted = strings.TrimPrefix(formatted, "+")
		formatted = formatted + "@s.whatsapp.net"
	}

	return formatted
}

// GenerateMessageID generates a unique message ID for WhatsApp messages
func GenerateMessageID() string {
	// Generate a random 8-byte ID
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return "msg_" + strings.ReplaceAll(time.Now().Format("20060102_150405"), "_", "")
	}

	return "msg_" + hex.EncodeToString(bytes)
}

// TruncateMessage truncates a message to a specified length for logging purposes
func TruncateMessage(message string, maxLength int) string {
	if len(message) <= maxLength {
		return message
	}
	return message[:maxLength] + "..."
}
