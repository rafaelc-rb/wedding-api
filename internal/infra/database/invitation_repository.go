package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

type invitationRepository struct {
	db *sql.DB
}

func NewInvitationRepository(db *sql.DB) repository.InvitationRepository {
	return &invitationRepository{db: db}
}

func (r *invitationRepository) Create(ctx context.Context, inv *entity.Invitation) error {
	query := `
		INSERT INTO invitations (id, wedding_id, code, label, max_guests, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		inv.ID, inv.WeddingID, inv.Code, inv.Label, inv.MaxGuests, inv.Notes,
		inv.CreatedAt, inv.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("invitationRepository.Create: %w", err)
	}
	return nil
}

func (r *invitationRepository) FindByID(ctx context.Context, weddingID, id string) (*entity.Invitation, error) {
	query := `
		SELECT id, wedding_id, code, label, max_guests, notes, created_at, updated_at
		FROM invitations WHERE wedding_id = $1 AND id = $2`

	var inv entity.Invitation
	err := r.db.QueryRowContext(ctx, query, weddingID, id).Scan(
		&inv.ID, &inv.WeddingID, &inv.Code, &inv.Label, &inv.MaxGuests, &inv.Notes,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("invitationRepository.FindByID: %w", err)
	}
	return &inv, nil
}

func (r *invitationRepository) FindByCode(ctx context.Context, weddingID, code string) (*entity.Invitation, error) {
	query := `
		SELECT id, wedding_id, code, label, max_guests, notes, created_at, updated_at
		FROM invitations WHERE wedding_id = $1 AND code = $2`

	var inv entity.Invitation
	err := r.db.QueryRowContext(ctx, query, weddingID, code).Scan(
		&inv.ID, &inv.WeddingID, &inv.Code, &inv.Label, &inv.MaxGuests, &inv.Notes,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("invitationRepository.FindByCode: %w", err)
	}
	return &inv, nil
}

func (r *invitationRepository) List(ctx context.Context, weddingID string, page, perPage int, search string) ([]entity.Invitation, int, error) {
	countQuery := `SELECT COUNT(*) FROM invitations WHERE wedding_id = $1`
	listQuery := `
		SELECT id, wedding_id, code, label, max_guests, notes, created_at, updated_at
		FROM invitations WHERE wedding_id = $1`

	args := []any{weddingID}
	paramIdx := 2

	if search != "" {
		f := fmt.Sprintf(` AND (label ILIKE $%d OR code ILIKE $%d)`, paramIdx, paramIdx)
		countQuery += f
		listQuery += f
		args = append(args, "%"+search+"%")
		paramIdx++
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("invitationRepository.List: count: %w", err)
	}

	listQuery += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, paramIdx, paramIdx+1)
	offset := (page - 1) * perPage
	listArgs := append(args, perPage, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("invitationRepository.List: query: %w", err)
	}
	defer rows.Close()

	var invitations []entity.Invitation
	for rows.Next() {
		var inv entity.Invitation
		if err := rows.Scan(&inv.ID, &inv.WeddingID, &inv.Code, &inv.Label, &inv.MaxGuests, &inv.Notes, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("invitationRepository.List: scan: %w", err)
		}
		invitations = append(invitations, inv)
	}
	return invitations, total, nil
}

func (r *invitationRepository) Update(ctx context.Context, inv *entity.Invitation) error {
	query := `
		UPDATE invitations SET code = $1, label = $2, max_guests = $3, notes = $4, updated_at = $5
		WHERE wedding_id = $6 AND id = $7`

	res, err := r.db.ExecContext(ctx, query,
		inv.Code, inv.Label, inv.MaxGuests, inv.Notes, inv.UpdatedAt,
		inv.WeddingID, inv.ID,
	)
	if err != nil {
		return fmt.Errorf("invitationRepository.Update: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *invitationRepository) Delete(ctx context.Context, weddingID, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM invitations WHERE wedding_id = $1 AND id = $2`, weddingID, id)
	if err != nil {
		return fmt.Errorf("invitationRepository.Delete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *invitationRepository) CountByWedding(ctx context.Context, weddingID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM invitations WHERE wedding_id = $1`, weddingID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("invitationRepository.CountByWedding: %w", err)
	}
	return count, nil
}
