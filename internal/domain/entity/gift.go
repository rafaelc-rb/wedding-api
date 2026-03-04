package entity

import "time"

type GiftStatus string

const (
	GiftStatusAvailable GiftStatus = "available"
	GiftStatusPurchased GiftStatus = "purchased"
)

type Gift struct {
	ID          string
	WeddingID   string
	Name        string
	Description string
	Price       float64
	ImageURL    string
	Category    string
	Status      GiftStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
