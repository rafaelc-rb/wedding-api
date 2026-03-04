package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/dto"
	"github.com/by-r2/weddo-api/internal/infra/web/middleware"
	"github.com/by-r2/weddo-api/internal/usecase/invitation"
)

type InvitationHandler struct {
	invUC *invitation.UseCase
}

func NewInvitationHandler(uc *invitation.UseCase) *InvitationHandler {
	return &InvitationHandler{invUC: uc}
}

func (h *InvitationHandler) Create(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())

	var req dto.CreateInvitationRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida. Verifique os campos obrigatórios.")
		return
	}

	guestNames := make([]string, len(req.Guests))
	for i, g := range req.Guests {
		guestNames[i] = g.Name
	}

	inv, err := h.invUC.Create(r.Context(), invitation.CreateInput{
		WeddingID: weddingID,
		Code:      req.Code,
		Label:     req.Label,
		MaxGuests: req.MaxGuests,
		Notes:     req.Notes,
		Guests:    guestNames,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Erro ao criar convite.")
		return
	}

	respondJSON(w, http.StatusCreated, toInvitationResponse(inv))
}

func (h *InvitationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	inv, err := h.invUC.FindByID(r.Context(), weddingID, id)
	if err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Convite não encontrado.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro interno do servidor.")
		return
	}

	respondJSON(w, http.StatusOK, toInvitationResponse(inv))
}

func (h *InvitationHandler) List(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	page, perPage := parsePagination(r)
	search := r.URL.Query().Get("search")

	invitations, total, err := h.invUC.List(r.Context(), weddingID, page, perPage, search)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Erro ao listar convites.")
		return
	}

	items := make([]dto.InvitationResponse, len(invitations))
	for i, inv := range invitations {
		items[i] = toInvitationResponse(&inv)
	}

	respondJSON(w, http.StatusOK, dto.PaginatedResponse{
		Data: items,
		Meta: buildMeta(page, perPage, total),
	})
}

func (h *InvitationHandler) Update(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	var req dto.UpdateInvitationRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida.")
		return
	}

	inv, err := h.invUC.Update(r.Context(), invitation.UpdateInput{
		WeddingID: weddingID,
		ID:        id,
		Code:      req.Code,
		Label:     req.Label,
		MaxGuests: req.MaxGuests,
		Notes:     req.Notes,
	})
	if err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Convite não encontrado.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro ao atualizar convite.")
		return
	}

	respondJSON(w, http.StatusOK, toInvitationResponse(inv))
}

func (h *InvitationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.invUC.Delete(r.Context(), weddingID, id); err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Convite não encontrado.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro ao remover convite.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *InvitationHandler) AddGuest(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	invID := chi.URLParam(r, "id")

	var req dto.AddGuestRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida.")
		return
	}

	guest, err := h.invUC.AddGuest(r.Context(), weddingID, invID, req.Name, req.Phone, req.Email)
	if err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Convite não encontrado.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro ao adicionar convidado.")
		return
	}

	respondJSON(w, http.StatusCreated, toGuestResponse(guest))
}

func toInvitationResponse(inv *entity.Invitation) dto.InvitationResponse {
	resp := dto.InvitationResponse{
		ID:        inv.ID,
		Code:      inv.Code,
		Label:     inv.Label,
		MaxGuests: inv.MaxGuests,
		Notes:     inv.Notes,
		CreatedAt: inv.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: inv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if len(inv.Guests) > 0 {
		resp.Guests = make([]dto.GuestResponse, len(inv.Guests))
		for i, g := range inv.Guests {
			resp.Guests[i] = toGuestResponse(&g)
		}
	}

	return resp
}

func parsePagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return page, perPage
}

func buildMeta(page, perPage, total int) dto.PaginationMeta {
	totalPages := total / perPage
	if total%perPage > 0 {
		totalPages++
	}
	return dto.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
