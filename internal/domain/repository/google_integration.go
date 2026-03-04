package repository

import (
	"context"

	"github.com/by-r2/weddo-api/internal/domain/entity"
)

type GoogleIntegrationRepository interface {
	Upsert(ctx context.Context, integration *entity.GoogleIntegration) error
	FindByWeddingID(ctx context.Context, weddingID string) (*entity.GoogleIntegration, error)
	DeleteByWeddingID(ctx context.Context, weddingID string) error
}
