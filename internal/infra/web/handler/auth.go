package handler

import (
	"net/http"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/dto"
	"github.com/by-r2/weddo-api/internal/usecase/wedding"
)

type AuthHandler struct {
	weddingUC *wedding.UseCase
}

func NewAuthHandler(weddingUC *wedding.UseCase) *AuthHandler {
	return &AuthHandler{weddingUC: weddingUC}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida. Verifique email e senha.")
		return
	}

	token, wed, err := h.weddingUC.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == entity.ErrUnauthorized {
			respondError(w, http.StatusUnauthorized, "Email ou senha incorretos.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro interno do servidor.")
		return
	}

	respondJSON(w, http.StatusOK, dto.LoginResponse{
		Token: token,
		Wedding: dto.WeddingSummary{
			ID:    wed.ID,
			Slug:  wed.Slug,
			Title: wed.Title,
		},
	})
}
