package payment

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("payment: not found")

// Create's query embeds, atomically, the same "clinic is active" and
// (when DentistID is set) "dentist is active and belongs to this
// clinic" guards used by internal/dentist — see spec Business Rule 1.
type Repository interface {
	Create(ctx context.Context, p Payment) error
	GetByID(ctx context.Context, id string) (Payment, error)
	// Approve transitions status "pending" -> "approved". Returns
	// ErrNotFound if the payment does not exist or is not "pending"
	// (already approved — idempotency guard against duplicate runs).
	Approve(ctx context.Context, id string, approvedAt time.Time) error
}

type postgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) Repository {
	return &postgresRepository{db: db}
}

type paymentRow struct {
	ID          string     `db:"id"`
	ClinicID    string     `db:"clinic_id"`
	DentistID   *string    `db:"dentist_id"`
	AmountCents int64      `db:"amount_cents"`
	Status      string     `db:"status"`
	PixCode     string     `db:"pix_code"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	ApprovedAt  *time.Time `db:"approved_at"`
}

func (r *postgresRepository) Create(ctx context.Context, p Payment) error {
	const query = `
		INSERT INTO payments (id, clinic_id, dentist_id, amount_cents, status, pix_code, created_at, updated_at)
		SELECT :id, :clinic_id, :dentist_id, :amount_cents, :status, :pix_code, :created_at, :updated_at
		WHERE EXISTS (SELECT 1 FROM clinics c WHERE c.id = :clinic_id AND c.deleted_at IS NULL)
		AND (CAST(:dentist_id AS uuid) IS NULL OR EXISTS (
			SELECT 1 FROM dentists d WHERE d.id = :dentist_id AND d.clinic_id = :clinic_id AND d.deleted_at IS NULL
		))`

	res, err := r.db.NamedExecContext(ctx, query, toRow(p))
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *postgresRepository) GetByID(ctx context.Context, id string) (Payment, error) {
	const query = `SELECT * FROM payments WHERE id = $1`

	var row paymentRow
	err := r.db.GetContext(ctx, &row, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Payment{}, ErrNotFound
	}
	if err != nil {
		return Payment{}, err
	}
	return row.toPayment(), nil
}

func (r *postgresRepository) Approve(ctx context.Context, id string, approvedAt time.Time) error {
	const query = `UPDATE payments SET status = 'approved', approved_at = $2, updated_at = $2 WHERE id = $1 AND status = 'pending'`

	res, err := r.db.ExecContext(ctx, query, id, approvedAt)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func checkRowsAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func toRow(p Payment) paymentRow {
	return paymentRow{
		ID:          p.ID,
		ClinicID:    p.ClinicID,
		DentistID:   p.DentistID,
		AmountCents: p.AmountCents,
		Status:      p.Status,
		PixCode:     p.PixCode,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		ApprovedAt:  p.ApprovedAt,
	}
}

func (row paymentRow) toPayment() Payment {
	return Payment{
		ID:          row.ID,
		ClinicID:    row.ClinicID,
		DentistID:   row.DentistID,
		AmountCents: row.AmountCents,
		Status:      row.Status,
		PixCode:     row.PixCode,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		ApprovedAt:  row.ApprovedAt,
	}
}
