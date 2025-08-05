package routes

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

// GetSwaggerHandler returns the httpSwagger.WrapHandler
func GetSwaggerHandler() http.HandlerFunc {
	return httpSwagger.WrapHandler
}
