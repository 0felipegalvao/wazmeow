package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator defines the interface for validation
type Validator interface {
	// Validate validates a struct and returns validation errors
	Validate(s interface{}) error
	// ValidateField validates a single field
	ValidateField(field interface{}, tag string) error
	// RegisterValidation registers a custom validation function
	RegisterValidation(tag string, fn ValidationFunc) error
}

// ValidationFunc represents a custom validation function
type ValidationFunc func(fl FieldLevel) bool

// FieldLevel contains all the information and helper functions
// to validate a field
type FieldLevel interface {
	// Top returns the top level struct, if any
	Top() reflect.Value
	// Parent returns the parent struct, if any
	Parent() reflect.Value
	// Field returns the field's value
	Field() reflect.Value
	// FieldName returns the field's name
	FieldName() string
	// StructFieldName returns the struct field's name
	StructFieldName() string
	// Param returns the param for the tag
	Param() string
	// GetTag returns the validation tag with the given name
	GetTag() string
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return e.Message
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// PlaygroundValidator implements Validator using go-playground/validator
type PlaygroundValidator struct {
	validator *validator.Validate
}

// New creates a new validator instance
func New() Validator {
	v := validator.New()
	
	// Register custom tag name function to use json tags
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	pv := &PlaygroundValidator{validator: v}
	
	// Register custom validations
	pv.registerCustomValidations()
	
	return pv
}

// Validate validates a struct
func (pv *PlaygroundValidator) Validate(s interface{}) error {
	err := pv.validator.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrors ValidationErrors
	
	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validatorErrors {
			validationErrors = append(validationErrors, ValidationError{
				Field:   e.Field(),
				Tag:     e.Tag(),
				Value:   fmt.Sprintf("%v", e.Value()),
				Message: pv.getErrorMessage(e),
			})
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return err
}

// ValidateField validates a single field
func (pv *PlaygroundValidator) ValidateField(field interface{}, tag string) error {
	err := pv.validator.Var(field, tag)
	if err == nil {
		return nil
	}

	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validatorErrors {
			return ValidationError{
				Field:   e.Field(),
				Tag:     e.Tag(),
				Value:   fmt.Sprintf("%v", e.Value()),
				Message: pv.getErrorMessage(e),
			}
		}
	}

	return err
}

// RegisterValidation registers a custom validation function
func (pv *PlaygroundValidator) RegisterValidation(tag string, fn ValidationFunc) error {
	return pv.validator.RegisterValidation(tag, func(fl validator.FieldLevel) bool {
		return fn(&playgroundFieldLevel{fl: fl})
	})
}

// playgroundFieldLevel wraps validator.FieldLevel to implement our FieldLevel interface
type playgroundFieldLevel struct {
	fl validator.FieldLevel
}

func (pfl *playgroundFieldLevel) Top() reflect.Value {
	return pfl.fl.Top()
}

func (pfl *playgroundFieldLevel) Parent() reflect.Value {
	return pfl.fl.Parent()
}

func (pfl *playgroundFieldLevel) Field() reflect.Value {
	return pfl.fl.Field()
}

func (pfl *playgroundFieldLevel) FieldName() string {
	return pfl.fl.FieldName()
}

func (pfl *playgroundFieldLevel) StructFieldName() string {
	return pfl.fl.StructFieldName()
}

func (pfl *playgroundFieldLevel) Param() string {
	return pfl.fl.Param()
}

func (pfl *playgroundFieldLevel) GetTag() string {
	return pfl.fl.GetTag()
}

// getErrorMessage returns a human-readable error message for validation errors
func (pv *PlaygroundValidator) getErrorMessage(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be a valid number", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "session_name":
		return fmt.Sprintf("%s must be a valid session name (3-50 characters, alphanumeric, spaces, hyphens, underscores only)", field)
	case "phone_number":
		return fmt.Sprintf("%s must be a valid phone number", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// registerCustomValidations registers custom validation functions
func (pv *PlaygroundValidator) registerCustomValidations() {
	// Session name validation
	pv.RegisterValidation("session_name", func(fl FieldLevel) bool {
		value := fl.Field().String()
		if len(value) < 3 || len(value) > 50 {
			return false
		}
		
		// Check for valid characters (alphanumeric, spaces, hyphens, underscores)
		for _, char := range value {
			if !isValidSessionNameChar(char) {
				return false
			}
		}
		
		return true
	})

	// Phone number validation (basic)
	pv.RegisterValidation("phone_number", func(fl FieldLevel) bool {
		value := fl.Field().String()
		if len(value) < 10 || len(value) > 15 {
			return false
		}
		
		// Must start with + and contain only digits after that
		if !strings.HasPrefix(value, "+") {
			return false
		}
		
		for _, char := range value[1:] {
			if char < '0' || char > '9' {
				return false
			}
		}
		
		return true
	})
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

// NoopValidator is a validator that does nothing (useful for testing)
type NoopValidator struct{}

func (nv *NoopValidator) Validate(s interface{}) error                              { return nil }
func (nv *NoopValidator) ValidateField(field interface{}, tag string) error        { return nil }
func (nv *NoopValidator) RegisterValidation(tag string, fn ValidationFunc) error   { return nil }
