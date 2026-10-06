package handler

import (
	"climaCEP/internal/usecase"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const DEADLINE = time.Second * 5

type Handler struct {
	useCase *usecase.UseCaseClimaCep
}

func NewHandler(useCase *usecase.UseCaseClimaCep) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	var dto usecase.CepInputDTO
	dto.Cep = r.URL.Query().Get("cep")

	ctx, cancel := context.WithTimeout(r.Context(), DEADLINE)
	defer cancel()

	weather, useErr := h.useCase.Execute(ctx, dto)
	if useErr != nil {
		http.Error(w, useErr.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(weather)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
