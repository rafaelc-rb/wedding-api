package dto

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RSVP

type RSVPRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

// Invitations

type CreateInvitationRequest struct {
	Code      string               `json:"code" validate:"required,max=50"`
	Label     string               `json:"label" validate:"required,max=100"`
	MaxGuests int                  `json:"max_guests" validate:"required,min=1"`
	Notes     string               `json:"notes"`
	Guests    []CreateGuestInline  `json:"guests"`
}

type CreateGuestInline struct {
	Name string `json:"name" validate:"required,max=100"`
}

type UpdateInvitationRequest struct {
	Code      string `json:"code" validate:"required,max=50"`
	Label     string `json:"label" validate:"required,max=100"`
	MaxGuests int    `json:"max_guests" validate:"required,min=1"`
	Notes     string `json:"notes"`
}

// Guests

type AddGuestRequest struct {
	Name  string `json:"name" validate:"required,max=100"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type UpdateGuestRequest struct {
	Name   string `json:"name" validate:"required,max=100"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	Status string `json:"status" validate:"omitempty,oneof=pending confirmed declined"`
}

// Gifts

type CreateGiftRequest struct {
	Name        string  `json:"name" validate:"required,max=200"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category" validate:"required,max=50"`
}

type UpdateGiftRequest struct {
	Name        string  `json:"name" validate:"required,max=200"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category" validate:"required,max=50"`
	Status      string  `json:"status" validate:"omitempty,oneof=available purchased"`
}

// Payments

type PurchaseGiftRequest struct {
	PayerName       string `json:"payer_name" validate:"required,max=100"`
	PayerEmail      string `json:"payer_email" validate:"required,email"`
	Message         string `json:"message"`
	PaymentMethod   string `json:"payment_method" validate:"required,oneof=pix credit_card"`
	CardToken       string `json:"card_token"`
	PaymentMethodID string `json:"payment_method_id"`
	Installments    int    `json:"installments"`
}
