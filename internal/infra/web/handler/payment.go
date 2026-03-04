package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/domain/entity"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/dto"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/infra/web/middleware"
	paymentuc "github.com/rafaeljurkfitz/mr-wedding-api/internal/usecase/payment"
)

type PaymentHandler struct {
	paymentUC *paymentuc.UseCase
}

func NewPaymentHandler(uc *paymentuc.UseCase) *PaymentHandler {
	return &PaymentHandler{paymentUC: uc}
}

func (h *PaymentHandler) checkAvailable(w http.ResponseWriter) bool {
	if h.paymentUC == nil {
		respondError(w, http.StatusServiceUnavailable, "Pagamentos não configurados. Configure MP_ACCESS_TOKEN.")
		return false
	}
	return true
}

func (h *PaymentHandler) Purchase(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}
	weddingID := middleware.GetWeddingID(r.Context())
	giftID := chi.URLParam(r, "id")

	var req dto.PurchaseGiftRequest
	if err := decodeAndValidate(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Requisição inválida. Verifique os campos obrigatórios.")
		return
	}

	if req.PaymentMethod == "credit_card" && req.CardToken == "" {
		respondError(w, http.StatusBadRequest, "O campo card_token é obrigatório para pagamento com cartão.")
		return
	}

	result, err := h.paymentUC.Purchase(r.Context(), paymentuc.PurchaseInput{
		WeddingID:       weddingID,
		GiftID:          giftID,
		PayerName:       req.PayerName,
		PayerEmail:      req.PayerEmail,
		Message:         req.Message,
		PaymentMethod:   req.PaymentMethod,
		CardToken:       req.CardToken,
		PaymentMethodID: req.PaymentMethodID,
		Installments:    req.Installments,
	})
	if err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Presente não encontrado.")
			return
		}
		if err == paymentuc.ErrGiftUnavailable {
			respondError(w, http.StatusConflict, "Este presente já foi comprado.")
			return
		}
		slog.Error("payment.Purchase", "error", err)
		respondError(w, http.StatusInternalServerError, "Erro ao processar pagamento.")
		return
	}

	resp := dto.PurchaseResponse{
		PaymentID:  result.Payment.ID,
		ProviderID: result.Payment.ProviderID,
		Status:     string(result.Payment.Status),
	}

	if result.QRCode != "" {
		resp.QRCode = result.QRCode
		resp.QRCodeBase64 = result.QRCodeBase64
		if result.Payment.PixExpiration != nil {
			s := result.Payment.PixExpiration.Format("2006-01-02T15:04:05Z")
			resp.ExpiresAt = &s
		}
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *PaymentHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	p, giftName, err := h.paymentUC.GetStatus(r.Context(), weddingID, id)
	if err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Pagamento não encontrado.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro interno do servidor.")
		return
	}

	respondJSON(w, http.StatusOK, dto.PaymentStatusResponse{
		PaymentID: p.ID,
		Status:    string(p.Status),
		GiftName:  giftName,
	})
}

type webhookPayload struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type webhookData struct {
	ID string `json:"id"`
}

func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}
	var payload webhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if payload.Action != "payment.updated" && payload.Action != "payment.created" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var data webhookData
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		slog.Error("webhook: failed to parse data", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if data.ID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.paymentUC.HandleWebhook(r.Context(), data.ID); err != nil {
		slog.Error("webhook: failed to handle", "provider_id", data.ID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PaymentHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}
	weddingID := middleware.GetWeddingID(r.Context())
	page, perPage := parsePagination(r)
	status := r.URL.Query().Get("status")
	giftID := r.URL.Query().Get("gift_id")

	payments, total, err := h.paymentUC.List(r.Context(), weddingID, page, perPage, status, giftID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Erro ao listar pagamentos.")
		return
	}

	items := make([]dto.PaymentResponse, len(payments))
	for i, p := range payments {
		items[i] = toPaymentResponse(&p)
	}

	respondJSON(w, http.StatusOK, dto.PaginatedResponse{
		Data: items,
		Meta: buildMeta(page, perPage, total),
	})
}

func (h *PaymentHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	if !h.checkAvailable(w) {
		return
	}
	weddingID := middleware.GetWeddingID(r.Context())
	id := chi.URLParam(r, "id")

	p, err := h.paymentUC.FindByID(r.Context(), weddingID, id)
	if err != nil {
		if err == entity.ErrNotFound {
			respondError(w, http.StatusNotFound, "Pagamento não encontrado.")
			return
		}
		respondError(w, http.StatusInternalServerError, "Erro interno do servidor.")
		return
	}

	respondJSON(w, http.StatusOK, toPaymentResponse(p))
}

func toPaymentResponse(p *entity.Payment) dto.PaymentResponse {
	resp := dto.PaymentResponse{
		ID:            p.ID,
		GiftID:        p.GiftID,
		ProviderID:    p.ProviderID,
		Amount:        p.Amount,
		Status:        string(p.Status),
		PaymentMethod: string(p.PaymentMethod),
		PayerName:     p.PayerName,
		PayerEmail:    p.PayerEmail,
		Message:       p.Message,
		CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if p.PaidAt != nil {
		s := p.PaidAt.Format("2006-01-02T15:04:05Z")
		resp.PaidAt = &s
	}
	return resp
}
