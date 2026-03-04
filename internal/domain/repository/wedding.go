package repository

import (
	"context"

	"github.com/by-r2/weddo-api/internal/domain/entity"
)

type WeddingRepository interface {
	Create(ctx context.Context, wedding *entity.Wedding) error
	FindByID(ctx context.Context, id string) (*entity.Wedding, error)
	FindByEmail(ctx context.Context, email string) (*entity.Wedding, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Wedding, error)
}
