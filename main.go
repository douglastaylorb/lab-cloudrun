package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type ViaCEPResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
	Erro        bool   `json:"erro"`
}

type WeatherAPIResponse struct {
	Location struct {
		Name    string `json:"name"`
		Region  string `json:"region"`
		Country string `json:"country"`
	} `json:"location"`
	Current struct {
		TempC float64 `json:"temp_c"`
		TempF float64 `json:"temp_f"`
	} `json:"current"`
}

type TemperatureResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

const (
	viaCEPBaseURL     = "https://viacep.com.br/ws/%s/json/"
	weatherAPIBaseURL = "http://api.weatherapi.com/v1/current.json"
	weatherAPIKey     = "cole_sua_chave_aqui"
)

func main() {
	e := echo.New()

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type")
			return next(c)
		}
	})

	e.GET("/weather", handleWeatherRequest)

	e.Logger.Fatal(e.Start(":8080"))
}

func handleWeatherRequest(c echo.Context) error {
	cep := c.QueryParam("cep")
	if cep == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "CEP não fornecido"})
	}

	if !isValidCEP(cep) {
		return c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Message: "invalid zipcode"})
	}

	cep = cleanCEP(cep)
	fmt.Printf("Buscando localização para CEP: %s\n", cep)

	location, err := getLocationByCEP(cep)
	if err != nil {
		fmt.Printf("Erro ao buscar localização: %v\n", err)
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, ErrorResponse{Message: "can not find zipcode"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Erro ao buscar localização"})
	}

	fmt.Printf("Localização encontrada: %s\n", location)

	temperature, err := getTemperatureByLocation(location)
	if err != nil {
		fmt.Printf("Erro ao buscar temperatura: %v\n", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Erro ao buscar temperatura"})
	}

	fmt.Printf("Temperatura obtida com sucesso\n")
	return c.JSON(http.StatusOK, temperature)
}

func isValidCEP(cep string) bool {
	cepRegex := regexp.MustCompile(`^\d{8}$`)
	return cepRegex.MatchString(cleanCEP(cep))
}

func cleanCEP(cep string) string {
	re := regexp.MustCompile(`[^\d]`)
	return re.ReplaceAllString(cep, "")
}

func getLocationByCEP(cep string) (string, error) {
	url := strings.ReplaceAll(viaCEPBaseURL, "%s", cep)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer requisição para ViaCEP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ViaCEP retornou status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta do ViaCEP: %w", err)
	}

	var viaCEPResp ViaCEPResponse
	if err := json.Unmarshal(bodyBytes, &viaCEPResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar JSON do ViaCEP: %w (body: %s)", err, string(bodyBytes))
	}

	if viaCEPResp.Erro {
		return "", echo.NewHTTPError(http.StatusNotFound, "not found")
	}

	if viaCEPResp.Localidade == "" {
		return "", fmt.Errorf("localidade não encontrada para o CEP %s", cep)
	}

	return viaCEPResp.Localidade, nil
}

func getTemperatureByLocation(location string) (*TemperatureResponse, error) {
	encodedLocation := url.QueryEscape(location)
	url := fmt.Sprintf("%s?key=%s&q=%s&aqi=no", weatherAPIBaseURL, weatherAPIKey, encodedLocation)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição para WeatherAPI: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o corpo da resposta da WeatherAPI: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WeatherAPI retornou status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var weatherResp WeatherAPIResponse
	if err := json.Unmarshal(bodyBytes, &weatherResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON da WeatherAPI: %w (JSON original: %s)", err, string(bodyBytes))
	}

	tempC := weatherResp.Current.TempC
	tempF := celsiusToFahrenheit(tempC)
	tempK := celsiusToKelvin(tempC)

	return &TemperatureResponse{
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}, nil
}

func celsiusToFahrenheit(celsius float64) float64 {
	return celsius*1.8 + 32
}

func celsiusToKelvin(celsius float64) float64 {
	return celsius + 273
}
