package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHandleWeatherRequest(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		cep            string
		expectedStatus int
	}{
		{
			name:           "CEP válido",
			cep:            "29108790",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "CEP inválido",
			cep:            "123",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "CEP não fornecido",
			cep:            "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/weather?cep="+tt.cep, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handleWeatherRequest(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestIsValidCEP(t *testing.T) {
	tests := []struct {
		name     string
		cep      string
		expected bool
	}{
		{"CEP válido com hífen", "01310-100", true},
		{"CEP válido sem hífen", "01310100", true},
		{"CEP inválido - muito curto", "123", false},
		{"CEP inválido - muito longo", "123456789", false},
		{"CEP inválido - com letras", "0131010A", false},
		{"CEP vazio", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidCEP(tt.cep)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemperatureConversions(t *testing.T) {
	t.Run("Celsius para Fahrenheit", func(t *testing.T) {
		result := celsiusToFahrenheit(0)
		assert.Equal(t, 32.0, result)

		result = celsiusToFahrenheit(100)
		assert.Equal(t, 212.0, result)
	})

	t.Run("Celsius para Kelvin", func(t *testing.T) {
		result := celsiusToKelvin(0)
		assert.Equal(t, 273.0, result)

		result = celsiusToKelvin(-273)
		assert.Equal(t, 0.0, result)
	})
}
