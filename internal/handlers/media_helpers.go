package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"strings"

	"github.com/nfnt/resize"
	"github.com/vincent-petithory/dataurl"
)

// MediaHelper provides utility functions for media handling
type MediaHelper struct{}

// NewMediaHelper creates a new media helper instance
func NewMediaHelper() *MediaHelper {
	return &MediaHelper{}
}

// DecodeDataURL decodes a data URL and returns the raw data
func (m *MediaHelper) DecodeDataURL(dataURL string) ([]byte, string, error) {
	parsed, err := dataurl.DecodeString(dataURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode data URL: %w", err)
	}
	
	return parsed.Data, parsed.MediaType.String(), nil
}

// ValidateImageFormat validates if the data URL is a valid image format
func (m *MediaHelper) ValidateImageFormat(dataURL string) error {
	if !strings.HasPrefix(dataURL, "data:image/") {
		return fmt.Errorf("invalid image format: must start with 'data:image/'")
	}
	return nil
}

// ValidateAudioFormat validates if the data URL is a valid audio format
func (m *MediaHelper) ValidateAudioFormat(dataURL string) error {
	if !strings.HasPrefix(dataURL, "data:audio/ogg") {
		return fmt.Errorf("invalid audio format: must start with 'data:audio/ogg'")
	}
	return nil
}

// ValidateDocumentFormat validates if the data URL is a valid document format
func (m *MediaHelper) ValidateDocumentFormat(dataURL string) error {
	if !strings.HasPrefix(dataURL, "data:application/octet-stream") {
		return fmt.Errorf("invalid document format: must start with 'data:application/octet-stream'")
	}
	return nil
}

// ValidateVideoFormat validates if the data URL is a valid video format
func (m *MediaHelper) ValidateVideoFormat(dataURL string) error {
	if !strings.HasPrefix(dataURL, "data:video/") {
		return fmt.Errorf("invalid video format: must start with 'data:video/'")
	}
	return nil
}

// DetectMimeType detects the MIME type of the given data
func (m *MediaHelper) DetectMimeType(data []byte) string {
	return http.DetectContentType(data)
}

// GenerateThumbnail generates a 72x72 thumbnail for images
func (m *MediaHelper) GenerateThumbnail(imageData []byte) ([]byte, error) {
	// Decode the image
	reader := bytes.NewReader(imageData)
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize to 72x72 thumbnail using Lanczos resampling
	thumbnail := resize.Thumbnail(72, 72, img, resize.Lanczos3)

	// Encode back to bytes
	var buf bytes.Buffer
	
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 75})
	case "png":
		err = png.Encode(&buf, thumbnail)
	default:
		// Default to JPEG for unknown formats
		err = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 75})
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateVideoThumbnail generates a thumbnail for video (placeholder implementation)
func (m *MediaHelper) GenerateVideoThumbnail(videoData []byte) ([]byte, error) {
	// For now, return empty bytes as video thumbnail generation is complex
	// In a real implementation, you would use ffmpeg or similar
	return []byte{}, nil
}

// ValidateFileSize validates if the file size is within limits
func (m *MediaHelper) ValidateFileSize(data []byte, maxSizeMB int) error {
	sizeMB := len(data) / (1024 * 1024)
	if sizeMB > maxSizeMB {
		return fmt.Errorf("file size %dMB exceeds limit of %dMB", sizeMB, maxSizeMB)
	}
	return nil
}

// CreateTempFile creates a temporary file with the given data
func (m *MediaHelper) CreateTempFile(data []byte, prefix, suffix string) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", prefix+"*"+suffix)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Reset file pointer to beginning
	if _, err := tmpFile.Seek(0, 0); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to seek temp file: %w", err)
	}

	return tmpFile, nil
}

// CleanupTempFile removes a temporary file
func (m *MediaHelper) CleanupTempFile(file *os.File) {
	if file != nil {
		filename := file.Name()
		file.Close()
		os.Remove(filename)
	}
}

// GetAudioMimeType returns the standard audio MIME type for WhatsApp
func (m *MediaHelper) GetAudioMimeType() string {
	return "audio/ogg; codecs=opus"
}

// IsValidCoordinate validates latitude and longitude values
func (m *MediaHelper) IsValidCoordinate(lat, lng float64) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("invalid latitude: %f (must be between -90 and 90)", lat)
	}
	if lng < -180 || lng > 180 {
		return fmt.Errorf("invalid longitude: %f (must be between -180 and 180)", lng)
	}
	return nil
}

// FormatVCard creates a vCard string for contact sharing
func (m *MediaHelper) FormatVCard(name, phone string) string {
	return fmt.Sprintf(`BEGIN:VCARD
VERSION:3.0
FN:%s
TEL:%s
END:VCARD`, name, phone)
}
