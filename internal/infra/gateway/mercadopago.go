package gateway

import (
	"context"
	"fmt"
	"strconv"
	"time"

	gw "github.com/by-r2/weddo-api/internal/domain/gateway"

	mpconfig "github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
)

type MercadoPagoGateway struct {
	client           payment.Client
	notificationURL  string
	pixExpirationMin int
}

func NewMercadoPagoGateway(accessToken, notificationURL string, pixExpirationMin int) (*MercadoPagoGateway, error) {
	cfg, err := mpconfig.New(accessToken)
	if err != nil {
		return nil, fmt.Errorf("mercadopago.New: %w", err)
	}

	return &MercadoPagoGateway{
		client:           payment.NewClient(cfg),
		notificationURL:  notificationURL,
		pixExpirationMin: pixExpirationMin,
	}, nil
}

func (g *MercadoPagoGateway) Name() string {
	return "mercadopago"
}

func (g *MercadoPagoGateway) CreatePayment(ctx context.Context, input gw.CreatePaymentInput) (*gw.PaymentResult, error) {
	if input.PaymentMethod == "pix" {
		return g.createPixPayment(ctx, input)
	}
	return g.createCardPayment(ctx, input)
}

func (g *MercadoPagoGateway) createPixPayment(ctx context.Context, input gw.CreatePaymentInput) (*gw.PaymentResult, error) {
	expiration := time.Now().Add(time.Duration(g.pixExpirationMin) * time.Minute)

	req := payment.Request{
		TransactionAmount: input.Amount,
		PaymentMethodID:   "pix",
		Description:       input.Description,
		ExternalReference: input.ExternalReference,
		NotificationURL:   g.notificationURL,
		DateOfExpiration:  &expiration,
		Payer: &payment.PayerRequest{
			Email: input.PayerEmail,
		},
	}

	resp, err := g.client.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("mercadopago.CreatePixPayment: %w", err)
	}

	return &gw.PaymentResult{
		ProviderID:   strconv.Itoa(resp.ID),
		Status:       resp.Status,
		QRCode:       resp.PointOfInteraction.TransactionData.QRCode,
		QRCodeBase64: resp.PointOfInteraction.TransactionData.QRCodeBase64,
		ExpiresAt:    &expiration,
	}, nil
}

func (g *MercadoPagoGateway) createCardPayment(ctx context.Context, input gw.CreatePaymentInput) (*gw.PaymentResult, error) {
	installments := input.Installments
	if installments < 1 {
		installments = 1
	}

	req := payment.Request{
		TransactionAmount: input.Amount,
		Token:             input.CardToken,
		PaymentMethodID:   input.PaymentMethodID,
		Installments:      installments,
		Description:       input.Description,
		ExternalReference: input.ExternalReference,
		NotificationURL:   g.notificationURL,
		Payer: &payment.PayerRequest{
			Email: input.PayerEmail,
		},
	}

	resp, err := g.client.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("mercadopago.CreateCardPayment: %w", err)
	}

	return &gw.PaymentResult{
		ProviderID: strconv.Itoa(resp.ID),
		Status:     resp.Status,
	}, nil
}

func (g *MercadoPagoGateway) GetPaymentStatus(ctx context.Context, providerID string) (*gw.WebhookResult, error) {
	id, err := strconv.Atoi(providerID)
	if err != nil {
		return nil, fmt.Errorf("mercadopago.GetPaymentStatus: invalid provider ID: %w", err)
	}

	resp, err := g.client.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("mercadopago.GetPaymentStatus: %w", err)
	}

	return &gw.WebhookResult{
		ProviderID: strconv.Itoa(resp.ID),
		Status:     resp.Status,
	}, nil
}
