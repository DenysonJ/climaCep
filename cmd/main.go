package main

import (
	"climaCEP/configs"
	"climaCEP/internal/infrastructure/web/handler"
	"climaCEP/internal/usecase"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	config, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}

	useCaseGetClima := usecase.NewUseCaseClimaCep(config.WeatherApiKey)
	cepHandler := handler.NewHandler(useCaseGetClima)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/cep", cepHandler.Get)
	serverErr := http.ListenAndServe(":"+config.WebServerPort, r)
	if serverErr != nil {
		log.Fatal(serverErr)
	}
}
