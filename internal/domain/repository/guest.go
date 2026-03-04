package repository

import (
	"context"

	"github.com/by-r2/weddo-api/internal/domain/entity"
)

type GuestRepository interface {
	Create(ctx context.Context, guest *entity.Guest) error
	FindByID(ctx context.Context, weddingID, id string) (*entity.Guest, error)
	FindByName(ctx context.Context, weddingID, name string) (*entity.Guest, error)
	ListByInvitation(ctx context.Context, weddingID, invitationID string) ([]entity.Guest, error)
	List(ctx context.Context, weddingID string, page, perPage int, status, search string) ([]entity.Guest, int, error)
	Update(ctx context.Context, guest *entity.Guest) error
	Delete(ctx context.Context, weddingID, id string) error
	CountByWedding(ctx context.Context, weddingID string) (total, confirmed, pending, declined int, err error)
}
