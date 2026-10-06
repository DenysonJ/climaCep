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
	Cep string `json:"cep"`
	//Logradouro  string `json:"logradouro"`
	//Complemento string `json:"complemento"`
	//Unidade     string `json:"unidade"`
	//Bairro      string `json:"bairro"`
	Localidade string `json:"localidade"`
	//Uf          string `json:"uf"`
	//Estado      string `json:"estado"`
	//Regiao      string `json:"regiao"`
	//Ibge        string `json:"ibge"`
	//Gia         string `json:"gia"`
	//Ddd         string `json:"ddd"`
	//Siafi       string `json:"siafi"`
}

type WeatherAPIResponse struct {
	Current struct {
		LastUpdatedEpoch int     `json:"last_updated_epoch"`
		LastUpdated      string  `json:"last_updated"`
		TempC            float64 `json:"temp_c"`
		TempF            float64 `json:"temp_f"`
		IsDay            int     `json:"is_day"`
		Condition        struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
			Code int    `json:"code"`
		} `json:"condition"`
		WindMph    float64 `json:"wind_mph"`
		WindKph    float64 `json:"wind_kph"`
		WindDegree int     `json:"wind_degree"`
		WindDir    string  `json:"wind_dir"`
		PressureMb float64 `json:"pressure_mb"`
		PressureIn float64 `json:"pressure_in"`
		PrecipMm   float64 `json:"precip_mm"`
		PrecipIn   float64 `json:"precip_in"`
		Humidity   int     `json:"humidity"`
		Cloud      int     `json:"cloud"`
		FeelslikeC float64 `json:"feelslike_c"`
		FeelslikeF float64 `json:"feelslike_f"`
		VisKm      float64 `json:"vis_km"`
		VisMiles   float64 `json:"vis_miles"`
		Uv         float64 `json:"uv"`
		GustMph    float64 `json:"gust_mph"`
		GustKph    float64 `json:"gust_kph"`
	} `json:"current"`
}

type WeatherOutputDTO struct {
	TempCelsius    string
	TempFahrenheit string
	TempKelvin     string
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

	respCEP, doWeatherErr := http.DefaultClient.Do(req)
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
	req.Header.Set("Accept", "application/json")

	respWeather, doWeatherErr := http.DefaultClient.Do(reqWeather)
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
