package repository

import (
	"context"

	"github.com/rafaeljurkfitz/mr-wedding-api/internal/domain/entity"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *entity.Payment) error
	FindByID(ctx context.Context, weddingID, id string) (*entity.Payment, error)
	FindByProviderID(ctx context.Context, providerID string) (*entity.Payment, error)
	List(ctx context.Context, weddingID string, page, perPage int, status, giftID string) ([]entity.Payment, int, error)
	Update(ctx context.Context, payment *entity.Payment) error
	SumByWedding(ctx context.Context, weddingID string) (totalRevenue float64, totalPayments int, err error)
}
