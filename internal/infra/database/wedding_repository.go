package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

type weddingRepository struct {
	db *sql.DB
}

func NewWeddingRepository(db *sql.DB) repository.WeddingRepository {
	return &weddingRepository{db: db}
}

func (r *weddingRepository) Create(ctx context.Context, w *entity.Wedding) error {
	query := `
		INSERT INTO weddings (id, slug, title, date, partner1_name, partner2_name, admin_email, admin_pass_hash, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.ExecContext(ctx, query,
		w.ID, w.Slug, w.Title, w.Date, w.Partner1Name, w.Partner2Name,
		w.AdminEmail, w.AdminPassHash, w.Active, w.CreatedAt, w.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("weddingRepository.Create: %w", err)
	}
	return nil
}

func (r *weddingRepository) FindByID(ctx context.Context, id string) (*entity.Wedding, error) {
	return r.findByColumn(ctx, "id", id)
}

func (r *weddingRepository) FindByEmail(ctx context.Context, email string) (*entity.Wedding, error) {
	return r.findByColumn(ctx, "admin_email", email)
}

func (r *weddingRepository) FindBySlug(ctx context.Context, slug string) (*entity.Wedding, error) {
	return r.findByColumn(ctx, "slug", slug)
}

func (r *weddingRepository) findByColumn(ctx context.Context, column, value string) (*entity.Wedding, error) {
	query := fmt.Sprintf(`
		SELECT id, slug, title, date, partner1_name, partner2_name, admin_email, admin_pass_hash, active, created_at, updated_at
		FROM weddings WHERE %s = $1`, column)

	var w entity.Wedding
	var dateStr sql.NullString

	err := r.db.QueryRowContext(ctx, query, value).Scan(
		&w.ID, &w.Slug, &w.Title, &dateStr, &w.Partner1Name, &w.Partner2Name,
		&w.AdminEmail, &w.AdminPassHash, &w.Active, &w.CreatedAt, &w.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("weddingRepository.findByColumn(%s): %w", column, err)
	}

	if dateStr.Valid {
		w.Date = dateStr.String
	}

	return &w, nil
}
