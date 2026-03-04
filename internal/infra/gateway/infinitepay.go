package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	gw "github.com/by-r2/weddo-api/internal/domain/gateway"
)

const infinitePayBaseURL = "https://api.infinitepay.io/invoices/public/checkout"

type InfinitePayGateway struct {
	handle      string
	redirectURL string
	webhookURL  string
	httpClient  *http.Client
}

func NewInfinitePayGateway(handle, redirectURL, webhookURL string) *InfinitePayGateway {
	return &InfinitePayGateway{
		handle:      handle,
		redirectURL: redirectURL,
		webhookURL:  webhookURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (g *InfinitePayGateway) Name() string {
	return "infinitepay"
}

// --- request/response types para a API InfinitePay ---

type ipCreateLinkRequest struct {
	Handle      string       `json:"handle"`
	Items       []ipItem     `json:"items"`
	OrderNSU    string       `json:"order_nsu,omitempty"`
	RedirectURL string       `json:"redirect_url,omitempty"`
	WebhookURL  string       `json:"webhook_url,omitempty"`
	Customer    *ipCustomer  `json:"customer,omitempty"`
}

type ipItem struct {
	Quantity    int    `json:"quantity"`
	Price       int    `json:"price"`
	Description string `json:"description"`
}

type ipCustomer struct {
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

type ipCreateLinkResponse struct {
	URL  string `json:"url"`
	Slug string `json:"slug"`
}

type ipPaymentCheckRequest struct {
	Handle         string `json:"handle"`
	OrderNSU       string `json:"order_nsu"`
	TransactionNSU string `json:"transaction_nsu,omitempty"`
	Slug           string `json:"slug,omitempty"`
}

type ipPaymentCheckResponse struct {
	Success       bool   `json:"success"`
	Paid          bool   `json:"paid"`
	Amount        int    `json:"amount"`
	PaidAmount    int    `json:"paid_amount"`
	Installments  int    `json:"installments"`
	CaptureMethod string `json:"capture_method"`
}

func (g *InfinitePayGateway) CreatePayment(ctx context.Context, input gw.CreatePaymentInput) (*gw.PaymentResult, error) {
	priceInCents := int(input.Amount * 100)

	reqBody := ipCreateLinkRequest{
		Handle:   g.handle,
		OrderNSU: input.ExternalReference,
		Items: []ipItem{
			{
				Quantity:    1,
				Price:       priceInCents,
				Description: input.Description,
			},
		},
	}

	if g.redirectURL != "" {
		reqBody.RedirectURL = g.redirectURL
	}
	if input.RedirectURL != "" {
		reqBody.RedirectURL = input.RedirectURL
	}
	if g.webhookURL != "" {
		reqBody.WebhookURL = g.webhookURL
	}

	if input.PayerName != "" || input.PayerEmail != "" {
		reqBody.Customer = &ipCustomer{
			Name:  input.PayerName,
			Email: input.PayerEmail,
		}
	}

	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("infinitepay.CreatePayment: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, infinitePayBaseURL+"/links", bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("infinitepay.CreatePayment: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("infinitepay.CreatePayment: request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("infinitepay.CreatePayment: read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("infinitepay.CreatePayment: status %d: %s", resp.StatusCode, string(respBody))
	}

	var ipResp ipCreateLinkResponse
	if err := json.Unmarshal(respBody, &ipResp); err != nil {
		return nil, fmt.Errorf("infinitepay.CreatePayment: unmarshal: %w", err)
	}

	return &gw.PaymentResult{
		ProviderID:  ipResp.Slug,
		Status:      "pending",
		CheckoutURL: ipResp.URL,
	}, nil
}

func (g *InfinitePayGateway) GetPaymentStatus(ctx context.Context, providerID string) (*gw.WebhookResult, error) {
	reqBody := ipPaymentCheckRequest{
		Handle: g.handle,
		Slug:   providerID,
	}

	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("infinitepay.GetPaymentStatus: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, infinitePayBaseURL+"/payment_check", bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("infinitepay.GetPaymentStatus: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("infinitepay.GetPaymentStatus: request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("infinitepay.GetPaymentStatus: read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("infinitepay.GetPaymentStatus: status %d: %s", resp.StatusCode, string(respBody))
	}

	var ipResp ipPaymentCheckResponse
	if err := json.Unmarshal(respBody, &ipResp); err != nil {
		return nil, fmt.Errorf("infinitepay.GetPaymentStatus: unmarshal: %w", err)
	}

	status := "pending"
	if ipResp.Paid {
		status = "approved"
	}

	return &gw.WebhookResult{
		ProviderID: providerID,
		Status:     status,
	}, nil
}
