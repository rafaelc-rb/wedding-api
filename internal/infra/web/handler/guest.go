package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/dto"
	"github.com/by-r2/weddo-api/internal/infra/web/middleware"
	"github.com/by-r2/weddo-api/internal/usecase/guest"
)

type GuestHandler struct {
	guestUC *guest.UseCase
}

func NewGuestHandler(uc *guest.UseCase) *GuestHandler {
	return &GuestHandler{guestUC: uc}
}

func (h *GuestHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	g, err := h.guestUC.FindByID(r.Context(), weddingID, id)
	if err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Convidado não encontrado.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro interno do servidor.")
		return
	}

	respondJSON(w, http.StatusOK, toGuestResponse(g))
}

func (h *GuestHandler) List(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	page, perPage := parsePagination(r)
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	guests, total, err := h.guestUC.List(r.Context(), weddingID, page, perPage, status, search)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Erro ao listar convidados.")
		return
	}

	items := make([]dto.GuestResponse, len(guests))
	for i, g := range guests {
		items[i] = toGuestResponse(&g)
	}

	respondJSON(w, http.StatusOK, dto.PaginatedResponse{
		Data: items,
		Meta: buildMeta(page, perPage, total),
	})
}

func (h *GuestHandler) Update(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	var req dto.UpdateGuestRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida.")
		return
	}

	g, err := h.guestUC.Update(r.Context(), guest.UpdateInput{
		WeddingID: weddingID,
		ID:        id,
		Name:      req.Name,
		Phone:     req.Phone,
		Email:     req.Email,
		Status:    req.Status,
	})
	if err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Convidado não encontrado.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro ao atualizar convidado.")
		return
	}

	respondJSON(w, http.StatusOK, toGuestResponse(g))
}

func (h *GuestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.guestUC.Delete(r.Context(), weddingID, id); err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Convidado não encontrado.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro ao remover convidado.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toGuestResponse(g *entity.Guest) dto.GuestResponse {
	resp := dto.GuestResponse{
		ID:     g.ID,
		Name:   g.Name,
		Phone:  g.Phone,
		Email:  g.Email,
		Status: string(g.Status),
	}
	if g.ConfirmedAt != nil {
		s := g.ConfirmedAt.Format("2006-01-02T15:04:05Z")
		resp.ConfirmedAt = &s
	}
	return resp
}

func toGuestSummary(g *entity.Guest) dto.GuestSummary {
	resp := dto.GuestSummary{
		ID:     g.ID,
		Name:   g.Name,
		Status: string(g.Status),
	}
	if g.ConfirmedAt != nil {
		s := g.ConfirmedAt.Format("2006-01-02T15:04:05Z")
		resp.ConfirmedAt = &s
	}
	return resp
}
