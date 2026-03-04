package dto

type LoginResponse struct {
	Token   string          `json:"token"`
	Wedding WeddingSummary  `json:"wedding"`
}

type WeddingSummary struct {
	ID    string `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type PaginatedResponse struct {
	Data any            `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// RSVP

type RSVPResponse struct {
	Guest      GuestSummary      `json:"guest"`
	Invitation InvitationSummary `json:"invitation"`
	Message    string            `json:"message"`
}

type RSVPInvitationResponse struct {
	Invitation InvitationSummary `json:"invitation"`
	Guests     []GuestPublic     `json:"guests"`
}

type GuestPublic struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Invitations

type InvitationResponse struct {
	ID        string          `json:"id"`
	Code      string          `json:"code"`
	Label     string          `json:"label"`
	MaxGuests int             `json:"max_guests"`
	Notes     string          `json:"notes,omitempty"`
	Guests    []GuestResponse `json:"guests,omitempty"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

type InvitationSummary struct {
	Label     string `json:"label"`
	MaxGuests int    `json:"max_guests,omitempty"`
}

// Guests

type GuestResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Phone       string  `json:"phone,omitempty"`
	Email       string  `json:"email,omitempty"`
	Status      string  `json:"status"`
	ConfirmedAt *string `json:"confirmed_at"`
}

type GuestSummary struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	ConfirmedAt *string `json:"confirmed_at"`
}

// Gifts

type GiftResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url,omitempty"`
	Category    string  `json:"category"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// Payments

type PurchaseResponse struct {
	PaymentID    string  `json:"payment_id"`
	ProviderID   string  `json:"provider_id"`
	Status       string  `json:"status"`
	QRCode       string  `json:"qr_code,omitempty"`
	QRCodeBase64 string  `json:"qr_code_base64,omitempty"`
	ExpiresAt    *string `json:"expires_at,omitempty"`
}

type PaymentResponse struct {
	ID            string  `json:"id"`
	GiftID        string  `json:"gift_id"`
	ProviderID    string  `json:"provider_id,omitempty"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	PaymentMethod string  `json:"payment_method"`
	PayerName     string  `json:"payer_name"`
	PayerEmail    string  `json:"payer_email,omitempty"`
	Message       string  `json:"message,omitempty"`
	PaidAt        *string `json:"paid_at"`
	CreatedAt     string  `json:"created_at"`
}

type PaymentStatusResponse struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
	GiftName  string `json:"gift_name"`
}

// Dashboard

type DashboardResponse struct {
	RSVP  RSVPStats  `json:"rsvp"`
	Gifts *GiftStats `json:"gifts,omitempty"`
}

type RSVPStats struct {
	TotalInvitations int     `json:"total_invitations"`
	TotalGuests      int     `json:"total_guests"`
	Confirmed        int     `json:"confirmed"`
	Pending          int     `json:"pending"`
	Declined         int     `json:"declined"`
	ConfirmationRate float64 `json:"confirmation_rate"`
}

type GiftStats struct {
	TotalGifts    int     `json:"total_gifts"`
	Purchased     int     `json:"purchased"`
	Available     int     `json:"available"`
	TotalRevenue  float64 `json:"total_revenue"`
	TotalPayments int     `json:"total_payments"`
}
