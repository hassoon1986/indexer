// Package common provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/algorand/oapi-codegen DO NOT EDIT.
package common

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Returns 200 if healthy.
	// (GET /health)
	MakeHealthCheck(ctx echo.Context) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// MakeHealthCheck converts echo context to params.
func (w *ServerInterfaceWrapper) MakeHealthCheck(ctx echo.Context) error {

	validQueryParams := map[string]bool{
		"pretty": true,
	}

	// Check for unknown query parameters.
	for name, _ := range ctx.QueryParams() {
		if _, ok := validQueryParams[name]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unknown parameter detected: %s", name))
		}
	}

	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.MakeHealthCheck(ctx)
	return err
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}, si ServerInterface, m ...echo.MiddlewareFunc) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET("/health", wrapper.MakeHealthCheck, m...)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"AwvMh/6aAAA=",
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file.
func GetSwagger() (*openapi3.Swagger, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error loading Swagger: %s", err)
	}
	return swagger, nil
}
