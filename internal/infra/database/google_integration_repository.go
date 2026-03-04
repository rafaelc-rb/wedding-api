package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

type googleIntegrationRepository struct {
	db *sql.DB
}

func NewGoogleIntegrationRepository(db *sql.DB) repository.GoogleIntegrationRepository {
	return &googleIntegrationRepository{db: db}
}

func (r *googleIntegrationRepository) Upsert(ctx context.Context, gi *entity.GoogleIntegration) error {
	query := `
		INSERT INTO google_integrations (
			wedding_id, spreadsheet_id, spreadsheet_url,
			encrypted_access_token, encrypted_refresh_token, token_expiry,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (wedding_id) DO UPDATE SET
			spreadsheet_id = EXCLUDED.spreadsheet_id,
			spreadsheet_url = EXCLUDED.spreadsheet_url,
			encrypted_access_token = EXCLUDED.encrypted_access_token,
			encrypted_refresh_token = EXCLUDED.encrypted_refresh_token,
			token_expiry = EXCLUDED.token_expiry,
			updated_at = EXCLUDED.updated_at`

	_, err := r.db.ExecContext(ctx, query,
		gi.WeddingID,
		gi.SpreadsheetID,
		gi.SpreadsheetURL,
		gi.EncryptedAccessToken,
		gi.EncryptedRefreshToken,
		gi.TokenExpiry,
		gi.CreatedAt,
		gi.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("googleIntegrationRepository.Upsert: %w", err)
	}
	return nil
}

func (r *googleIntegrationRepository) FindByWeddingID(ctx context.Context, weddingID string) (*entity.GoogleIntegration, error) {
	query := `
		SELECT wedding_id, spreadsheet_id, spreadsheet_url, encrypted_access_token,
		       encrypted_refresh_token, token_expiry, created_at, updated_at
		FROM google_integrations
		WHERE wedding_id = $1`

	var gi entity.GoogleIntegration
	err := r.db.QueryRowContext(ctx, query, weddingID).Scan(
		&gi.WeddingID,
		&gi.SpreadsheetID,
		&gi.SpreadsheetURL,
		&gi.EncryptedAccessToken,
		&gi.EncryptedRefreshToken,
		&gi.TokenExpiry,
		&gi.CreatedAt,
		&gi.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("googleIntegrationRepository.FindByWeddingID: %w", err)
	}
	return &gi, nil
}

func (r *googleIntegrationRepository) DeleteByWeddingID(ctx context.Context, weddingID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM google_integrations WHERE wedding_id = $1`, weddingID)
	if err != nil {
		return fmt.Errorf("googleIntegrationRepository.DeleteByWeddingID: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrNotFound
	}
	return nil
}
