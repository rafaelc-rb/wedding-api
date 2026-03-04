package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rafaeljurkfitz/mr-wedding-api/internal/domain/entity"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/domain/repository"
)

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) repository.PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, p *entity.Payment) error {
	query := `
		INSERT INTO payments (id, gift_id, wedding_id, provider_id, amount, status, payment_method,
			payer_name, payer_email, message, pix_qr_code, pix_expiration, paid_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.GiftID, p.WeddingID, p.ProviderID, p.Amount, p.Status, p.PaymentMethod,
		p.PayerName, p.PayerEmail, p.Message, p.PixQRCode, p.PixExpiration,
		p.PaidAt, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("paymentRepository.Create: %w", err)
	}
	return nil
}

func (r *paymentRepository) FindByID(ctx context.Context, weddingID, id string) (*entity.Payment, error) {
	query := `
		SELECT id, gift_id, wedding_id, provider_id, amount, status, payment_method,
			payer_name, payer_email, message, pix_qr_code, pix_expiration, paid_at, created_at, updated_at
		FROM payments WHERE wedding_id = ? AND id = ?`

	return r.scanPayment(r.db.QueryRowContext(ctx, query, weddingID, id))
}

func (r *paymentRepository) FindByProviderID(ctx context.Context, providerID string) (*entity.Payment, error) {
	query := `
		SELECT id, gift_id, wedding_id, provider_id, amount, status, payment_method,
			payer_name, payer_email, message, pix_qr_code, pix_expiration, paid_at, created_at, updated_at
		FROM payments WHERE provider_id = ?`

	return r.scanPayment(r.db.QueryRowContext(ctx, query, providerID))
}

func (r *paymentRepository) List(ctx context.Context, weddingID string, page, perPage int, status, giftID string) ([]entity.Payment, int, error) {
	countQuery := `SELECT COUNT(*) FROM payments WHERE wedding_id = ?`
	listQuery := `
		SELECT id, gift_id, wedding_id, provider_id, amount, status, payment_method,
			payer_name, payer_email, message, pix_qr_code, pix_expiration, paid_at, created_at, updated_at
		FROM payments WHERE wedding_id = ?`

	args := []any{weddingID}

	if status != "" {
		f := ` AND status = ?`
		countQuery += f
		listQuery += f
		args = append(args, status)
	}
	if giftID != "" {
		f := ` AND gift_id = ?`
		countQuery += f
		listQuery += f
		args = append(args, giftID)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("paymentRepository.List: count: %w", err)
	}

	listQuery += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	offset := (page - 1) * perPage
	listArgs := append(args, perPage, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("paymentRepository.List: query: %w", err)
	}
	defer rows.Close()

	var payments []entity.Payment
	for rows.Next() {
		var p entity.Payment
		if err := rows.Scan(
			&p.ID, &p.GiftID, &p.WeddingID, &p.ProviderID, &p.Amount, &p.Status, &p.PaymentMethod,
			&p.PayerName, &p.PayerEmail, &p.Message, &p.PixQRCode, &p.PixExpiration,
			&p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("paymentRepository.List: scan: %w", err)
		}
		payments = append(payments, p)
	}
	return payments, total, nil
}

func (r *paymentRepository) Update(ctx context.Context, p *entity.Payment) error {
	query := `
		UPDATE payments SET provider_id = ?, status = ?, pix_qr_code = ?, pix_expiration = ?,
			paid_at = ?, updated_at = ?
		WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		p.ProviderID, p.Status, p.PixQRCode, p.PixExpiration, p.PaidAt, p.UpdatedAt,
		p.ID,
	)
	if err != nil {
		return fmt.Errorf("paymentRepository.Update: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *paymentRepository) SumByWedding(ctx context.Context, weddingID string) (totalRevenue float64, totalPayments int, err error) {
	query := `
		SELECT COALESCE(SUM(amount), 0), COUNT(*)
		FROM payments WHERE wedding_id = ? AND status = 'approved'`

	err = r.db.QueryRowContext(ctx, query, weddingID).Scan(&totalRevenue, &totalPayments)
	if err != nil {
		err = fmt.Errorf("paymentRepository.SumByWedding: %w", err)
	}
	return
}

func (r *paymentRepository) scanPayment(row *sql.Row) (*entity.Payment, error) {
	var p entity.Payment
	err := row.Scan(
		&p.ID, &p.GiftID, &p.WeddingID, &p.ProviderID, &p.Amount, &p.Status, &p.PaymentMethod,
		&p.PayerName, &p.PayerEmail, &p.Message, &p.PixQRCode, &p.PixExpiration,
		&p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("paymentRepository.scanPayment: %w", err)
	}
	return &p, nil
}
