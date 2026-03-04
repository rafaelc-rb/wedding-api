package repository

import (
	"context"

	"github.com/by-r2/weddo-api/internal/domain/entity"
)

type InvitationRepository interface {
	Create(ctx context.Context, inv *entity.Invitation) error
	FindByID(ctx context.Context, weddingID, id string) (*entity.Invitation, error)
	FindByCode(ctx context.Context, weddingID, code string) (*entity.Invitation, error)
	List(ctx context.Context, weddingID string, page, perPage int, search string) ([]entity.Invitation, int, error)
	Update(ctx context.Context, inv *entity.Invitation) error
	Delete(ctx context.Context, weddingID, id string) error
	CountByWedding(ctx context.Context, weddingID string) (int, error)
}
