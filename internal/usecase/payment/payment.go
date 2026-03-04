package payment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/domain/entity"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/domain/repository"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/infra/gateway"
)

var ErrGiftUnavailable = errors.New("presente não está disponível")

type UseCase struct {
	paymentRepo repository.PaymentRepository
	giftRepo    repository.GiftRepository
	mpGateway   *gateway.MercadoPagoGateway
}

func NewUseCase(pr repository.PaymentRepository, gr repository.GiftRepository, mp *gateway.MercadoPagoGateway) *UseCase {
	return &UseCase{paymentRepo: pr, giftRepo: gr, mpGateway: mp}
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
}

type PurchaseResult struct {
	Payment      *entity.Payment
	QRCode       string
	QRCodeBase64 string
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

	description := fmt.Sprintf("Presente: %s", gift.Name)
	externalRef := p.ID

	var mpResult *gateway.PaymentResult

	if input.PaymentMethod == string(entity.PaymentMethodPix) {
		mpResult, err = uc.mpGateway.CreatePixPayment(ctx, gateway.CreatePixInput{
			Amount:            gift.Price,
			Description:       description,
			PayerEmail:        input.PayerEmail,
			ExternalReference: externalRef,
		})
	} else {
		mpResult, err = uc.mpGateway.CreateCardPayment(ctx, gateway.CreateCardInput{
			Amount:            gift.Price,
			Description:       description,
			PayerEmail:        input.PayerEmail,
			CardToken:         input.CardToken,
			PaymentMethodID:   input.PaymentMethodID,
			Installments:      input.Installments,
			ExternalReference: externalRef,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("payment.Purchase: gateway: %w", err)
	}

	p.ProviderID = mpResult.ProviderID
	p.Status = mapProviderStatus(mpResult.Status)
	p.PixQRCode = mpResult.QRCode
	p.PixExpiration = mpResult.ExpiresAt

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
		QRCode:       mpResult.QRCode,
		QRCodeBase64: mpResult.QRCodeBase64,
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

// HandleWebhook processa notificação do Mercado Pago.
// Recebe o provider ID, consulta o status atual no MP e atualiza localmente.
func (uc *UseCase) HandleWebhook(ctx context.Context, providerID string) error {
	mpResult, err := uc.mpGateway.GetPayment(ctx, providerID)
	if err != nil {
		return fmt.Errorf("payment.HandleWebhook: get from provider: %w", err)
	}

	p, err := uc.paymentRepo.FindByProviderID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("payment.HandleWebhook: find payment: %w", err)
	}

	newStatus := mapProviderStatus(mpResult.Status)
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
