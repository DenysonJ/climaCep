package usecase

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func dispatchByHost(viaCep, weather roundTripFunc) roundTripFunc {
	return func(req *http.Request) (*http.Response, error) {
		switch req.URL.Host {
		case "viacep.com.br":
			return viaCep(req)
		case "api.weatherapi.com":
			return weather(req)
		default:
			return nil, errors.New("unexpected host: " + req.URL.Host)
		}
	}
}

func TestUseCaseClimaCep_Execute(t *testing.T) {
	cepOK := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return newResponse(http.StatusOK, `{"cep":"12345-678","localidade":"Sao Paulo"}`), nil
	})
	weatherOK := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return newResponse(http.StatusOK, `{"current":{"last_updated":"2026-10-06 10:00","temp_c":28.5}}`), nil
	})

	tests := []struct {
		name      string
		cep       string
		transport roundTripFunc
		want      WeatherOutputDTO
		wantErr   bool
	}{
		{
			name:      "success returns converted temperatures",
			cep:       "12345678",
			transport: dispatchByHost(cepOK, weatherOK),
			want: WeatherOutputDTO{
				TempCelsius:    "28.50",
				TempFahrenheit: "83.30",
				TempKelvin:     "301.50",
			},
			wantErr: false,
		},
		{
			name: "invalid cep does not call any service",
			cep:  "invalid",
			transport: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("should not be called")
			},
			wantErr: true,
		},
		{
			name: "viacep request error",
			cep:  "12345678",
			transport: dispatchByHost(
				func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("viacep down")
				},
				weatherOK,
			),
			wantErr: true,
		},
		{
			name: "viacep invalid json body",
			cep:  "12345678",
			transport: dispatchByHost(
				func(req *http.Request) (*http.Response, error) {
					return newResponse(http.StatusOK, `not a json`), nil
				},
				weatherOK,
			),
			wantErr: true,
		},
		{
			name: "weather request error",
			cep:  "12345678",
			transport: dispatchByHost(
				cepOK,
				func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("weather down")
				},
			),
			wantErr: true,
		},
		{
			name: "weather invalid json body",
			cep:  "12345678",
			transport: dispatchByHost(
				cepOK,
				func(req *http.Request) (*http.Response, error) {
					return newResponse(http.StatusOK, `not a json`), nil
				},
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUseCaseClimaCep(&http.Client{Transport: tt.transport}, "fake-key")

			got, err := uc.Execute(context.Background(), CepInputDTO{Cep: tt.cep})

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, WeatherOutputDTO{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertCelsiusToFahrenheit(t *testing.T) {
	tests := []struct {
		name    string
		celsius float64
		want    float64
	}{
		{name: "zero celsius", celsius: 0, want: 32},
		{name: "positive celsius", celsius: 28.5, want: 83.3},
		{name: "negative celsius", celsius: -40, want: -40},
		{name: "boiling point", celsius: 100, want: 212},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertCelsiusToFahrenheit(tt.celsius)

			assert.InDelta(t, tt.want, got, 0.0001)
		})
	}
}

func TestConvertCelsiusToKelvin(t *testing.T) {
	tests := []struct {
		name    string
		celsius float64
		want    float64
	}{
		{name: "zero celsius", celsius: 0, want: 273},
		{name: "positive celsius", celsius: 28.5, want: 301.5},
		{name: "negative celsius", celsius: -273, want: 0},
		{name: "boiling point", celsius: 100, want: 373},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertCelsiusToKelvin(tt.celsius)

			assert.InDelta(t, tt.want, got, 0.0001)
		})
	}
}
