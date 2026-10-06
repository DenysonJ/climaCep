package usecase

import (
	cep "climaCEP/internal/domain/vo"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	neturl "net/url"
)

const VIACEPURL = "http://viacep.com.br/ws/%s/json"
const WEATHERAPI = "http://api.weatherapi.com/v1/current.json"

type CepInputDTO struct {
	Cep string
}

type CepResponse struct {
	Cep        string `json:"cep"`
	Localidade string `json:"localidade"`
}

type WeatherAPIResponse struct {
	Current struct {
		LastUpdated string  `json:"last_updated"`
		TempC       float64 `json:"temp_c"`
	} `json:"current"`
}

type WeatherOutputDTO struct {
	TempCelsius    string `json:"temp_C"`
	TempFahrenheit string `json:"temp_F"`
	TempKelvin     string `json:"temp_K"`
}

type UseCaseClimaCep struct {
	httpClient    *http.Client
	weatherApiKey string
}

func NewUseCaseClimaCep(weatherApiKey string) *UseCaseClimaCep {
	return &UseCaseClimaCep{
		httpClient:    http.DefaultClient,
		weatherApiKey: weatherApiKey,
	}
}

func (u *UseCaseClimaCep) Execute(ctx context.Context, dto CepInputDTO) (WeatherOutputDTO, error) {
	cepVO, cepErr := cep.New(dto.Cep)
	if cepErr != nil {
		return WeatherOutputDTO{}, cepErr
	}

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(VIACEPURL, cepVO.Value), nil)
	if reqErr != nil {
		return WeatherOutputDTO{}, reqErr
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	respCEP, doWeatherErr := u.httpClient.Do(req)
	if doWeatherErr != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("timeout calling server (300ms exceeded): %w", doWeatherErr)
			return WeatherOutputDTO{}, context.DeadlineExceeded
		}
		log.Println("calling server: %w", doWeatherErr)
		return WeatherOutputDTO{}, doWeatherErr
	}
	defer respCEP.Body.Close()

	bodyCEP, readErr := io.ReadAll(respCEP.Body)
	if readErr != nil {
		log.Println("reading response body: %w", readErr)
		return WeatherOutputDTO{}, readErr
	}

	var cepJson CepResponse
	if unmarshalErr := json.Unmarshal(bodyCEP, &cepJson); unmarshalErr != nil {
		log.Println("unmarshalling response body CEP: %w", unmarshalErr)
		return WeatherOutputDTO{}, unmarshalErr
	}

	url := WEATHERAPI + "?key=" + u.weatherApiKey + "&q=" + neturl.QueryEscape(cepJson.Localidade)
	reqWeather, reqWeatherErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if reqWeatherErr != nil {
		return WeatherOutputDTO{}, reqWeatherErr
	}
	reqWeather.Header.Set("Accept", "application/json")

	respWeather, doWeatherErr := u.httpClient.Do(reqWeather)
	if doWeatherErr != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("timeout calling server (300ms exceeded): %w", doWeatherErr)
			return WeatherOutputDTO{}, context.DeadlineExceeded
		}
		log.Println("calling server: %w", doWeatherErr)
		return WeatherOutputDTO{}, doWeatherErr
	}
	defer respWeather.Body.Close()

	bodyWeather, readWeatherErr := io.ReadAll(respWeather.Body)
	if readWeatherErr != nil {
		log.Println("reading response body: %w", readWeatherErr)
		return WeatherOutputDTO{}, readWeatherErr
	}

	var weatherResp WeatherAPIResponse
	if unmarshalErr := json.Unmarshal(bodyWeather, &weatherResp); unmarshalErr != nil {
		log.Println("unmarshalling response body Weather: %w", unmarshalErr)
		return WeatherOutputDTO{}, unmarshalErr
	}

	currentCelsius := weatherResp.Current.TempC

	return WeatherOutputDTO{
		TempCelsius:    fmt.Sprintf("%.2f", currentCelsius),
		TempFahrenheit: fmt.Sprintf("%.2f", convertCelsiusToFahrenheit(currentCelsius)),
		TempKelvin:     fmt.Sprintf("%.2f", convertCelsiusToKelvin(currentCelsius)),
	}, nil
}

func convertCelsiusToFahrenheit(celsius float64) float64 {
	return (celsius * 1.8) + 32.0
}

func convertCelsiusToKelvin(celsius float64) float64 {
	return celsius + 273
}
