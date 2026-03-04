package handler

import (
	"net/http"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/dto"
	"github.com/by-r2/weddo-api/internal/infra/web/middleware"
	"github.com/by-r2/weddo-api/internal/usecase/rsvp"
)

type RSVPHandler struct {
	rsvpUC *rsvp.UseCase
}

func NewRSVPHandler(uc *rsvp.UseCase) *RSVPHandler {
	return &RSVPHandler{rsvpUC: uc}
}

func (h *RSVPHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())

	var req dto.RSVPRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida. Informe o nome como está no convite.")
		return
	}

	guest, inv, alreadyConfirmed, err := h.rsvpUC.Confirm(r.Context(), weddingID, req.Name)
	if err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Convidado não encontrado. Verifique se o nome está exatamente como no convite.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro interno do servidor.")
		return
	}

	msg := "Presença confirmada com sucesso!"
	status := http.StatusOK
	if alreadyConfirmed {
		msg = "Presença já estava confirmada."
		status = http.StatusConflict
	}

	respondJSON(w, status, dto.RSVPResponse{
		Guest:      toGuestSummary(guest),
		Invitation: dto.InvitationSummary{Label: inv.Label},
		Message:    msg,
	})
}

func (h *RSVPHandler) LookupInvitation(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())
	name := r.URL.Query().Get("name")

	if name == "" {
		respondError(w, http.StatusBadRequest, "Parâmetro 'name' é obrigatório.")
		return
	}

	inv, guests, err := h.rsvpUC.LookupInvitation(r.Context(), weddingID, name)
	if err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Convidado não encontrado. Verifique se o nome está exatamente como no convite.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro interno do servidor.")
		return
	}

	guestsPublic := make([]dto.GuestPublic, len(guests))
	for i, g := range guests {
		guestsPublic[i] = dto.GuestPublic{Name: g.Name, Status: string(g.Status)}
	}

	respondJSON(w, http.StatusOK, dto.RSVPInvitationResponse{
		Invitation: dto.InvitationSummary{Label: inv.Label, MaxGuests: inv.MaxGuests},
		Guests:     guestsPublic,
	})
}
