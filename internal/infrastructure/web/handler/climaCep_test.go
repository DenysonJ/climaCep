package handler

import (
	"climaCEP/internal/usecase"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
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

func withTransport(t *testing.T, rt roundTripFunc) {
	t.Helper()
	original := http.DefaultClient.Transport
	http.DefaultClient.Transport = rt
	t.Cleanup(func() {
		http.DefaultClient.Transport = original
	})
}

func TestHandler_Get(t *testing.T) {
	success := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Host {
		case "viacep.com.br":
			return newResponse(http.StatusOK, `{"cep":"12345-678","localidade":"Sao Paulo"}`), nil
		case "api.weatherapi.com":
			return newResponse(http.StatusOK, `{"current":{"temp_c":28.5}}`), nil
		default:
			return nil, errors.New("unexpected host: " + req.URL.Host)
		}
	})

	tests := []struct {
		name       string
		query      string
		transport  roundTripFunc
		wantStatus int
		wantBody   *usecase.WeatherOutputDTO
	}{
		{
			name:       "success returns weather json",
			query:      "cep=12345678",
			transport:  success,
			wantStatus: http.StatusOK,
			wantBody: &usecase.WeatherOutputDTO{
				TempCelsius:    "28.50",
				TempFahrenheit: "83.30",
				TempKelvin:     "301.50",
			},
		},
		{
			name:  "invalid cep returns internal server error",
			query: "cep=invalid",
			transport: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("should not be called")
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:  "missing cep returns internal server error",
			query: "",
			transport: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("should not be called")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withTransport(t, tt.transport)
			h := NewHandler(usecase.NewUseCaseClimaCep("fake-key"))
			req := httptest.NewRequest(http.MethodGet, "/cep?"+tt.query, nil)
			rec := httptest.NewRecorder()

			h.Get(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)

			if tt.wantBody == nil {
				return
			}

			assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
			var got usecase.WeatherOutputDTO
			require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
			assert.Equal(t, *tt.wantBody, got)
		})
	}
}
