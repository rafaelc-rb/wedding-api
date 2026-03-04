package handler

import (
	"errors"
	"net/http"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/infra/web/middleware"
	sheetsuc "github.com/by-r2/weddo-api/internal/usecase/sheets"
)

type SheetsHandler struct {
	uc *sheetsuc.UseCase
}

func NewSheetsHandler(uc *sheetsuc.UseCase) *SheetsHandler {
	return &SheetsHandler{uc: uc}
}

func (h *SheetsHandler) checkAvailable(w http.ResponseWriter) bool {
	if h.uc == nil {
		respondError(w, http.StatusServiceUnavailable, "Google Sheets OAuth não configurado. Configure GOOGLE_OAUTH_CLIENT_ID, GOOGLE_OAUTH_CLIENT_SECRET, GOOGLE_OAUTH_REDIRECT_URL e GOOGLE_OAUTH_TOKEN_CIPHER_KEY.")
		return false
	}
	return true
}

func (h *SheetsHandler) ConnectStart(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}
	weddingID := middleware.GetWeddingID(r.Context())
	result, err := h.uc.StartConnect(r.Context(), weddingID)
	if err != nil {
		if errors.Is(err, sheetsuc.ErrNotConfigured) {
			respondError(w, http.StatusServiceUnavailable, "Integração Google Sheets não configurada.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro ao iniciar conexão com Google.")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (h *SheetsHandler) ConnectCallback(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		respondError(w, http.StatusBadRequest, "Parâmetros code/state são obrigatórios.")
		return
	}

	result, err := h.uc.HandleOAuthCallback(r.Context(), code, state)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Falha no callback OAuth do Google.")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (h *SheetsHandler) Push(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}
	weddingID := middleware.GetWeddingID(r.Context())
	result, err := h.uc.Push(r.Context(), weddingID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Integração Google não encontrada para este wedding.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro ao sincronizar dados para o Google Sheets.")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (h *SheetsHandler) Pull(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}
	weddingID := middleware.GetWeddingID(r.Context())
	result, err := h.uc.Pull(r.Context(), weddingID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Integração Google não encontrada para este wedding.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro ao importar dados do Google Sheets.")
		return
	}
	respondJSON(w, http.StatusOK, result)
}
