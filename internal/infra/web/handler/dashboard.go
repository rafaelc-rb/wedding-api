package handler

import (
	"net/http"

	"github.com/by-r2/weddo-api/internal/dto"
	"github.com/by-r2/weddo-api/internal/infra/web/middleware"
	giftuc "github.com/by-r2/weddo-api/internal/usecase/gift"
	"github.com/by-r2/weddo-api/internal/usecase/guest"
)

type DashboardHandler struct {
	guestUC *guest.UseCase
	giftUC  *giftuc.UseCase
}

func NewDashboardHandler(guestUC *guest.UseCase, giftUC *giftuc.UseCase) *DashboardHandler {
	return &DashboardHandler{guestUC: guestUC, giftUC: giftUC}
}

func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	weddingID := middleware.GetWeddingID(r.Context())

	rsvpStats, err := h.guestUC.Dashboard(r.Context(), weddingID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Erro ao buscar estatísticas.")
		return
	}

	var rsvpRate float64
	if rsvpStats.TotalGuests > 0 {
		rsvpRate = float64(rsvpStats.Confirmed) / float64(rsvpStats.TotalGuests) * 100
	}

	resp := dto.DashboardResponse{
		RSVP: dto.RSVPStats{
			TotalInvitations: rsvpStats.TotalInvitations,
			TotalGuests:      rsvpStats.TotalGuests,
			Confirmed:        rsvpStats.Confirmed,
			Pending:          rsvpStats.Pending,
			Declined:         rsvpStats.Declined,
			ConfirmationRate: rsvpRate,
		},
	}

	giftStats, err := h.giftUC.Dashboard(r.Context(), weddingID)
	if err == nil && giftStats.TotalGifts > 0 {
		resp.Gifts = &dto.GiftStats{
			TotalGifts:    giftStats.TotalGifts,
			Purchased:     giftStats.Purchased,
			Available:     giftStats.Available,
			TotalRevenue:  giftStats.TotalRevenue,
			TotalPayments: giftStats.TotalPayments,
		}
	}

	respondJSON(w, http.StatusOK, resp)
}
