package entity

import "time"

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusApproved PaymentStatus = "approved"
	PaymentStatusRejected PaymentStatus = "rejected"
	PaymentStatusExpired  PaymentStatus = "expired"
)

type PaymentMethod string

const (
	PaymentMethodPix        PaymentMethod = "pix"
	PaymentMethodCreditCard PaymentMethod = "credit_card"
)

type Payment struct {
	ID             string
	GiftID         string
	WeddingID      string
	ProviderID     string
	Amount         float64
	Status         PaymentStatus
	PaymentMethod  PaymentMethod
	PayerName      string
	PayerEmail     string
	Message        string
	PixQRCode      string
	PixExpiration  *time.Time
	PaidAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
