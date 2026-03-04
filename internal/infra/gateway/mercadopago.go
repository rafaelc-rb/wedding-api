package gateway

import (
	"context"
	"fmt"
	"strconv"
	"time"

	mpconfig "github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
)

type MercadoPagoGateway struct {
	client          payment.Client
	notificationURL string
	pixExpirationMin int
}

func NewMercadoPagoGateway(accessToken, notificationURL string, pixExpirationMin int) (*MercadoPagoGateway, error) {
	cfg, err := mpconfig.New(accessToken)
	if err != nil {
		return nil, fmt.Errorf("mercadopago.New: %w", err)
	}

	return &MercadoPagoGateway{
		client:          payment.NewClient(cfg),
		notificationURL: notificationURL,
		pixExpirationMin: pixExpirationMin,
	}, nil
}

type CreatePixInput struct {
	Amount            float64
	Description       string
	PayerEmail        string
	ExternalReference string
}

type CreateCardInput struct {
	Amount            float64
	Description       string
	PayerEmail        string
	CardToken         string
	PaymentMethodID   string
	Installments      int
	ExternalReference string
}

type PaymentResult struct {
	ProviderID   string
	Status       string
	QRCode       string
	QRCodeBase64 string
	ExpiresAt    *time.Time
}

func (g *MercadoPagoGateway) CreatePixPayment(ctx context.Context, input CreatePixInput) (*PaymentResult, error) {
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

	return &PaymentResult{
		ProviderID:   strconv.Itoa(resp.ID),
		Status:       resp.Status,
		QRCode:       resp.PointOfInteraction.TransactionData.QRCode,
		QRCodeBase64: resp.PointOfInteraction.TransactionData.QRCodeBase64,
		ExpiresAt:    &expiration,
	}, nil
}

func (g *MercadoPagoGateway) CreateCardPayment(ctx context.Context, input CreateCardInput) (*PaymentResult, error) {
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

	return &PaymentResult{
		ProviderID: strconv.Itoa(resp.ID),
		Status:     resp.Status,
	}, nil
}

func (g *MercadoPagoGateway) GetPayment(ctx context.Context, providerID string) (*PaymentResult, error) {
	id, err := strconv.Atoi(providerID)
	if err != nil {
		return nil, fmt.Errorf("mercadopago.GetPayment: invalid provider ID: %w", err)
	}

	resp, err := g.client.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("mercadopago.GetPayment: %w", err)
	}

	return &PaymentResult{
		ProviderID: strconv.Itoa(resp.ID),
		Status:     resp.Status,
	}, nil
}
