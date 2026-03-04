package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/dto"
	"github.com/by-r2/weddo-api/internal/infra/web/middleware"
	giftuc "github.com/by-r2/weddo-api/internal/usecase/gift"
)

type GiftHandler struct {
	giftUC *giftuc.UseCase
}

func NewGiftHandler(uc *giftuc.UseCase) *GiftHandler {
	return &GiftHandler{giftUC: uc}
}

func (h *GiftHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	page, perPage := parsePagination(r)
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")

	gifts, total, err := h.giftUC.List(r.Context(), weddingID, page, perPage, category, string(entity.GiftStatusAvailable), search)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Erro ao listar presentes.")
		return
	}

	items := make([]dto.GiftResponse, len(gifts))
	for i, g := range gifts {
		items[i] = toGiftResponse(&g)
	}

	respondJSON(w, http.StatusOK, dto.PaginatedResponse{
		Data: items,
		Meta: buildMeta(page, perPage, total),
	})
}

func (h *GiftHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	g, err := h.giftUC.FindByID(r.Context(), weddingID, id)
	if err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Presente não encontrado.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro interno do servidor.")
		return
	}

	respondJSON(w, http.StatusOK, toGiftResponse(g))
}

func (h *GiftHandler) Create(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())

	var req dto.CreateGiftRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida. Verifique os campos obrigatórios.")
		return
	}

	g, err := h.giftUC.Create(r.Context(), giftuc.CreateInput{
		WeddingID:   weddingID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		Category:    req.Category,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Erro ao criar presente.")
		return
	}

	respondJSON(w, http.StatusCreated, toGiftResponse(g))
}

func (h *GiftHandler) List(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	page, perPage := parsePagination(r)
	category := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	gifts, total, err := h.giftUC.List(r.Context(), weddingID, page, perPage, category, status, search)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Erro ao listar presentes.")
		return
	}

	items := make([]dto.GiftResponse, len(gifts))
	for i, g := range gifts {
		items[i] = toGiftResponse(&g)
	}

	respondJSON(w, http.StatusOK, dto.PaginatedResponse{
		Data: items,
		Meta: buildMeta(page, perPage, total),
	})
}

func (h *GiftHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	g, err := h.giftUC.FindByID(r.Context(), weddingID, id)
	if err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Presente não encontrado.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro interno do servidor.")
		return
	}

	respondJSON(w, http.StatusOK, toGiftResponse(g))
}

func (h *GiftHandler) Update(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	var req dto.UpdateGiftRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida.")
		return
	}

	g, err := h.giftUC.Update(r.Context(), giftuc.UpdateInput{
		WeddingID:   weddingID,
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		Category:    req.Category,
		Status:      req.Status,
	})
	if err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Presente não encontrado.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro ao atualizar presente.")
		return
	}

	respondJSON(w, http.StatusOK, toGiftResponse(g))
}

func (h *GiftHandler) Delete(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.giftUC.Delete(r.Context(), weddingID, id); err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Presente não encontrado.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro ao remover presente.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toGiftResponse(g *entity.Gift) dto.GiftResponse {
	return dto.GiftResponse{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		Price:       g.Price,
		ImageURL:    g.ImageURL,
		Category:    g.Category,
		Status:      string(g.Status),
		CreatedAt:   g.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   g.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
