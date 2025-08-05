package dto_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wazmeow/internal/http/dto"
)

func TestSuccessResponse(t *testing.T) {
	t.Run("should create success response with data", func(t *testing.T) {
		// Arrange
		data := map[string]interface{}{
			"id":   "123",
			"name": "test",
		}
		message := "Operation successful"

		// Act
		response := dto.NewSuccessResponse(message, data)

		// Assert
		assert.True(t, response.Success)
		assert.Equal(t, message, response.Message)
		assert.Equal(t, data, response.Data)
	})

	t.Run("should create success response without data", func(t *testing.T) {
		// Arrange
		message := "Operation completed"

		// Act
		response := dto.NewSuccessResponse(message, nil)

		// Assert
		assert.True(t, response.Success)
		assert.Equal(t, message, response.Message)
		assert.Nil(t, response.Data)
	})

	t.Run("should marshal and unmarshal correctly", func(t *testing.T) {
		// Arrange
		data := map[string]string{"key": "value"}
		response := dto.NewSuccessResponse("test message", data)

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		// Act - Unmarshal from JSON
		var unmarshaled dto.SuccessResponse
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, response.Success, unmarshaled.Success)
		assert.Equal(t, response.Message, unmarshaled.Message)

		// Data should be unmarshaled as map[string]interface{}
		dataMap, ok := unmarshaled.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "value", dataMap["key"])
	})

	t.Run("should handle complex data structures", func(t *testing.T) {
		// Arrange
		complexData := map[string]interface{}{
			"string":  "value",
			"number":  42,
			"boolean": true,
			"array":   []string{"a", "b", "c"},
			"nested": map[string]interface{}{
				"inner": "value",
			},
		}
		response := dto.NewSuccessResponse("complex data", complexData)

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		// Act - Unmarshal from JSON
		var unmarshaled dto.SuccessResponse
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		// Assert
		assert.True(t, unmarshaled.Success)
		assert.Equal(t, "complex data", unmarshaled.Message)

		dataMap, ok := unmarshaled.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "value", dataMap["string"])
		assert.Equal(t, float64(42), dataMap["number"]) // JSON numbers become float64
		assert.Equal(t, true, dataMap["boolean"])
	})
}

func TestErrorResponse(t *testing.T) {
	t.Run("should create error response with error details", func(t *testing.T) {
		// Arrange
		message := "Operation failed"

		// Act
		response := dto.NewErrorResponse(message, "VALIDATION_ERROR", "Field validation failed")

		// Assert
		assert.False(t, response.Success)
		assert.Equal(t, message, response.Error)
		assert.Equal(t, "VALIDATION_ERROR", response.Code)
		assert.Equal(t, "Field validation failed", response.Details)
	})

	t.Run("should create error response without error details", func(t *testing.T) {
		// Arrange
		message := "Internal server error"

		// Act
		response := dto.NewErrorResponse(message, "INTERNAL_ERROR", "")

		// Assert
		assert.False(t, response.Success)
		assert.Equal(t, message, response.Error)
		assert.Equal(t, "INTERNAL_ERROR", response.Code)
		assert.Equal(t, "", response.Details)
	})

	t.Run("should marshal and unmarshal correctly", func(t *testing.T) {
		// Arrange
		response := dto.NewErrorResponse("Resource not found", "NOT_FOUND", "The requested resource was not found")

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		// Act - Unmarshal from JSON
		var unmarshaled dto.ErrorResponse
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, response.Success, unmarshaled.Success)
		assert.Equal(t, response.Error, unmarshaled.Error)
		assert.Equal(t, response.Code, unmarshaled.Code)
		assert.Equal(t, response.Details, unmarshaled.Details)
	})

	t.Run("should handle validation error details", func(t *testing.T) {
		// Arrange
		response := dto.NewErrorResponse("Validation failed", "VALIDATION_ERROR", "Multiple field validation errors")

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		// Act - Unmarshal from JSON
		var unmarshaled dto.ErrorResponse
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		// Assert
		assert.False(t, unmarshaled.Success)
		assert.Equal(t, "Validation failed", unmarshaled.Error)
		assert.Equal(t, "VALIDATION_ERROR", unmarshaled.Code)
		assert.Equal(t, "Multiple field validation errors", unmarshaled.Details)
	})
}

func TestResponseHelpers(t *testing.T) {
	t.Run("should create success response from error", func(t *testing.T) {
		// This test assumes there might be helper functions to create responses from errors
		// If such functions exist in the dto package

		// Arrange
		err := errors.New("test error")

		// Act - Create error response from error
		response := dto.NewErrorResponse(err.Error(), "GENERIC_ERROR", "")

		// Assert
		assert.False(t, response.Success)
		assert.Equal(t, "test error", response.Error)
		assert.Equal(t, "GENERIC_ERROR", response.Code)
		assert.Equal(t, "", response.Details)
	})

	t.Run("should handle empty message", func(t *testing.T) {
		// Arrange & Act
		successResponse := dto.NewSuccessResponse("", nil)
		errorResponse := dto.NewErrorResponse("", "", "")

		// Assert
		assert.True(t, successResponse.Success)
		assert.Empty(t, successResponse.Message)

		assert.False(t, errorResponse.Success)
		assert.Empty(t, errorResponse.Error)
	})

	t.Run("should maintain response structure consistency", func(t *testing.T) {
		// Arrange
		successResponse := dto.NewSuccessResponse("success", map[string]string{"key": "value"})
		errorResponse := dto.NewErrorResponse("error", "ERROR", "Generic error")

		// Act - Marshal both responses
		successJSON, err := json.Marshal(successResponse)
		require.NoError(t, err)
		errorJSON, err := json.Marshal(errorResponse)
		require.NoError(t, err)

		// Act - Unmarshal to generic maps to check structure
		var successMap map[string]interface{}
		var errorMap map[string]interface{}

		err = json.Unmarshal(successJSON, &successMap)
		require.NoError(t, err)
		err = json.Unmarshal(errorJSON, &errorMap)
		require.NoError(t, err)

		// Assert - Both should have the same keys
		successKeys := make([]string, 0, len(successMap))
		errorKeys := make([]string, 0, len(errorMap))

		for key := range successMap {
			successKeys = append(successKeys, key)
		}
		for key := range errorMap {
			errorKeys = append(errorKeys, key)
		}

		// Success response should have success, message, data fields
		expectedSuccessKeys := []string{"success", "message", "data"}
		for _, key := range expectedSuccessKeys {
			assert.Contains(t, successKeys, key)
		}

		// Error response should have success, error, code, details fields
		expectedErrorKeys := []string{"success", "error", "code", "details"}
		for _, key := range expectedErrorKeys {
			assert.Contains(t, errorKeys, key)
		}
	})

	t.Run("should handle nil data and error correctly", func(t *testing.T) {
		// Arrange
		successResponse := dto.NewSuccessResponse("success", nil)
		errorResponse := dto.NewErrorResponse("error", "ERROR", "")

		// Act - Marshal to JSON
		successJSON, err := json.Marshal(successResponse)
		require.NoError(t, err)
		errorJSON, err := json.Marshal(errorResponse)
		require.NoError(t, err)

		// Assert - Check that responses are properly serialized
		assert.Contains(t, string(successJSON), `"success":true`)
		assert.Contains(t, string(successJSON), `"message":"success"`)
		assert.Contains(t, string(errorJSON), `"success":false`)
		assert.Contains(t, string(errorJSON), `"error":"error"`)
	})
}
