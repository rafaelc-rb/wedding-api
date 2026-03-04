package repository

import (
	"context"

	"github.com/rafaeljurkfitz/mr-wedding-api/internal/domain/entity"
)

type GiftRepository interface {
	Create(ctx context.Context, gift *entity.Gift) error
	FindByID(ctx context.Context, weddingID, id string) (*entity.Gift, error)
	List(ctx context.Context, weddingID string, page, perPage int, category, status, search string) ([]entity.Gift, int, error)
	Update(ctx context.Context, gift *entity.Gift) error
	Delete(ctx context.Context, weddingID, id string) error
	CountByWedding(ctx context.Context, weddingID string) (total, available, purchased int, err error)
}
