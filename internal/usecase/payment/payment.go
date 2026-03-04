package payment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/by-r2/weddo-api/internal/domain/entity"
	gw "github.com/by-r2/weddo-api/internal/domain/gateway"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

var ErrGiftUnavailable = errors.New("presente não está disponível")

type UseCase struct {
	paymentRepo repository.PaymentRepository
	giftRepo    repository.GiftRepository
	gateway     gw.PaymentGateway
}

func NewUseCase(pr repository.PaymentRepository, gr repository.GiftRepository, gateway gw.PaymentGateway) *UseCase {
	return &UseCase{paymentRepo: pr, giftRepo: gr, gateway: gateway}
}

// ProviderName retorna o nome do provedor de pagamento ativo.
func (uc *UseCase) ProviderName() string {
	return uc.gateway.Name()
}

type PurchaseInput struct {
	WeddingID       string
	GiftID          string
	PayerName       string
	PayerEmail      string
	Message         string
	PaymentMethod   string
	CardToken       string
	PaymentMethodID string
	Installments    int
	RedirectURL     string
}

type PurchaseResult struct {
	Payment      *entity.Payment
	QRCode       string
	QRCodeBase64 string
	CheckoutURL  string
}

func (uc *UseCase) Purchase(ctx context.Context, input PurchaseInput) (*PurchaseResult, error) {
	gift, err := uc.giftRepo.FindByID(ctx, input.WeddingID, input.GiftID)
	if err != nil {
		return nil, err
	}

	if gift.Status != entity.GiftStatusAvailable {
		return nil, ErrGiftUnavailable
	}

	now := time.Now()
	p := &entity.Payment{
		ID:            uuid.New().String(),
		GiftID:        gift.ID,
		WeddingID:     input.WeddingID,
		Amount:        gift.Price,
		Status:        entity.PaymentStatusPending,
		PaymentMethod: entity.PaymentMethod(input.PaymentMethod),
		PayerName:     input.PayerName,
		PayerEmail:    input.PayerEmail,
		Message:       input.Message,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	gwInput := gw.CreatePaymentInput{
		Amount:            gift.Price,
		Description:       fmt.Sprintf("Presente: %s", gift.Name),
		PayerName:         input.PayerName,
		PayerEmail:        input.PayerEmail,
		PaymentMethod:     input.PaymentMethod,
		ExternalReference: p.ID,
		CardToken:         input.CardToken,
		PaymentMethodID:   input.PaymentMethodID,
		Installments:      input.Installments,
		RedirectURL:       input.RedirectURL,
	}

	gwResult, err := uc.gateway.CreatePayment(ctx, gwInput)
	if err != nil {
		return nil, fmt.Errorf("payment.Purchase: gateway: %w", err)
	}

	p.ProviderID = gwResult.ProviderID
	p.Status = mapProviderStatus(gwResult.Status)
	p.PixQRCode = gwResult.QRCode
	p.PixExpiration = gwResult.ExpiresAt

	if p.Status == entity.PaymentStatusApproved {
		p.PaidAt = &now
		gift.Status = entity.GiftStatusPurchased
		gift.UpdatedAt = now
		if err := uc.giftRepo.Update(ctx, gift); err != nil {
			slog.Error("payment.Purchase: failed to mark gift as purchased", "gift_id", gift.ID, "error", err)
		}
	}

	if err := uc.paymentRepo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("payment.Purchase: save: %w", err)
	}

	return &PurchaseResult{
		Payment:      p,
		QRCode:       gwResult.QRCode,
		QRCodeBase64: gwResult.QRCodeBase64,
		CheckoutURL:  gwResult.CheckoutURL,
	}, nil
}

func (uc *UseCase) GetStatus(ctx context.Context, weddingID, paymentID string) (*entity.Payment, string, error) {
	p, err := uc.paymentRepo.FindByID(ctx, weddingID, paymentID)
	if err != nil {
		return nil, "", err
	}

	gift, err := uc.giftRepo.FindByID(ctx, weddingID, p.GiftID)
	if err != nil {
		return nil, "", fmt.Errorf("payment.GetStatus: find gift: %w", err)
	}

	return p, gift.Name, nil
}

// HandleWebhook processa notificação do provedor de pagamento.
// Recebe o provider ID, consulta o status atual no provedor e atualiza localmente.
func (uc *UseCase) HandleWebhook(ctx context.Context, providerID string) error {
	gwResult, err := uc.gateway.GetPaymentStatus(ctx, providerID)
	if err != nil {
		return fmt.Errorf("payment.HandleWebhook: get from provider: %w", err)
	}

	p, err := uc.paymentRepo.FindByProviderID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("payment.HandleWebhook: find payment: %w", err)
	}

	newStatus := mapProviderStatus(gwResult.Status)
	if p.Status == newStatus {
		return nil
	}

	now := time.Now()
	p.Status = newStatus
	p.UpdatedAt = now

	if newStatus == entity.PaymentStatusApproved && p.PaidAt == nil {
		p.PaidAt = &now

		gift, err := uc.giftRepo.FindByID(ctx, p.WeddingID, p.GiftID)
		if err == nil && gift.Status == entity.GiftStatusAvailable {
			gift.Status = entity.GiftStatusPurchased
			gift.UpdatedAt = now
			if err := uc.giftRepo.Update(ctx, gift); err != nil {
				slog.Error("payment.HandleWebhook: failed to mark gift", "gift_id", gift.ID, "error", err)
			}
		}
	}

	if newStatus == entity.PaymentStatusExpired || newStatus == entity.PaymentStatusRejected {
		gift, err := uc.giftRepo.FindByID(ctx, p.WeddingID, p.GiftID)
		if err == nil && gift.Status == entity.GiftStatusPurchased {
			gift.Status = entity.GiftStatusAvailable
			gift.UpdatedAt = now
			if err := uc.giftRepo.Update(ctx, gift); err != nil {
				slog.Error("payment.HandleWebhook: failed to revert gift", "gift_id", gift.ID, "error", err)
			}
		}
	}

	if err := uc.paymentRepo.Update(ctx, p); err != nil {
		return fmt.Errorf("payment.HandleWebhook: update: %w", err)
	}

	slog.Info("payment.HandleWebhook: updated", "provider_id", providerID, "status", newStatus)
	return nil
}

// HandleInfinitePayWebhook processa a notificação direta da InfinitePay.
// A InfinitePay envia order_nsu (nosso payment ID) diretamente no payload.
func (uc *UseCase) HandleInfinitePayWebhook(ctx context.Context, orderNSU, invoiceSlug string, paid bool) error {
	p, err := uc.paymentRepo.FindByID(ctx, "", orderNSU)
	if err != nil {
		p, err = uc.paymentRepo.FindByProviderID(ctx, invoiceSlug)
		if err != nil {
			return fmt.Errorf("payment.HandleInfinitePayWebhook: find payment: %w", err)
		}
	}

	var newStatus entity.PaymentStatus
	if paid {
		newStatus = entity.PaymentStatusApproved
	} else {
		newStatus = entity.PaymentStatusPending
	}

	if p.Status == newStatus {
		return nil
	}

	now := time.Now()
	p.Status = newStatus
	p.UpdatedAt = now

	if newStatus == entity.PaymentStatusApproved && p.PaidAt == nil {
		p.PaidAt = &now

		gift, err := uc.giftRepo.FindByID(ctx, p.WeddingID, p.GiftID)
		if err == nil && gift.Status == entity.GiftStatusAvailable {
			gift.Status = entity.GiftStatusPurchased
			gift.UpdatedAt = now
			if err := uc.giftRepo.Update(ctx, gift); err != nil {
				slog.Error("payment.HandleIPWebhook: failed to mark gift", "gift_id", gift.ID, "error", err)
			}
		}
	}

	if err := uc.paymentRepo.Update(ctx, p); err != nil {
		return fmt.Errorf("payment.HandleInfinitePayWebhook: update: %w", err)
	}

	slog.Info("payment.HandleInfinitePayWebhook: updated", "order_nsu", orderNSU, "status", newStatus)
	return nil
}

func (uc *UseCase) List(ctx context.Context, weddingID string, page, perPage int, status, giftID string) ([]entity.Payment, int, error) {
	return uc.paymentRepo.List(ctx, weddingID, page, perPage, status, giftID)
}

func (uc *UseCase) FindByID(ctx context.Context, weddingID, id string) (*entity.Payment, error) {
	return uc.paymentRepo.FindByID(ctx, weddingID, id)
}

func mapProviderStatus(s string) entity.PaymentStatus {
	switch s {
	case "approved":
		return entity.PaymentStatusApproved
	case "rejected", "cancelled", "refunded", "charged_back":
		return entity.PaymentStatusRejected
	case "expired":
		return entity.PaymentStatusExpired
	default:
		return entity.PaymentStatusPending
	}
}
