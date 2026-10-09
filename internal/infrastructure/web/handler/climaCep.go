package handler

import (
	cep "climaCEP/internal/domain/vo"
	"climaCEP/internal/usecase"
	"context"
	"encoding/json"
	"errors"
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
		errorHandler(useErr, w)
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

func errorHandler(err error, w http.ResponseWriter) {
	if errors.Is(err, context.DeadlineExceeded) {
		http.Error(w, err.Error(), http.StatusRequestTimeout)
		return
	}
	if errors.Is(err, cep.ErrorInvalidCep) {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if errors.Is(err, usecase.ErrorCepNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
}
