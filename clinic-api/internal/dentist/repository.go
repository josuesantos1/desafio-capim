package dentist

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

const uniqueViolationCode = "23505"

var (
	ErrNotFound    = errors.New("dentist: not found")
	ErrEmailExists = errors.New("dentist: email already exists")
)

// Every method below that scopes an operation to a clinic_id embeds,
// in the same SQL statement, an "EXISTS (SELECT 1 FROM clinics ...
// WHERE deleted_at IS NULL)" guard. This makes the "clinic is active"
// check atomic with the operation itself (single MVCC snapshot),
// closing the TOCTOU window a separate pre-check-then-write would
// leave open.
type Repository interface {
	Create(ctx context.Context, d Dentist) error
	GetByID(ctx context.Context, clinicID, id string) (Dentist, error)
	Update(ctx context.Context, d Dentist) error
	SoftDelete(ctx context.Context, clinicID, id string, deletedAt time.Time) error
	List(ctx context.Context, clinicID string, params ListParams) (ListResult, error)
}

type postgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) Repository {
	return &postgresRepository{db: db}
}

type dentistRow struct {
	ID        string     `db:"id"`
	ClinicID  string     `db:"clinic_id"`
	Name      string     `db:"name"`
	Phone     string     `db:"phone"`
	Email     string     `db:"email"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (r *postgresRepository) Create(ctx context.Context, d Dentist) error {
	const query = `
		INSERT INTO dentists (id, clinic_id, name, phone, email, created_at, updated_at)
		SELECT :id, :clinic_id, :name, :phone, :email, :created_at, :updated_at
		WHERE EXISTS (SELECT 1 FROM clinics c WHERE c.id = :clinic_id AND c.deleted_at IS NULL)`

	res, err := r.db.NamedExecContext(ctx, query, toRow(d))
	if isUniqueViolation(err) {
		return ErrEmailExists
	}
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *postgresRepository) GetByID(ctx context.Context, clinicID, id string) (Dentist, error) {
	const query = `
		SELECT d.* FROM dentists d
		WHERE d.clinic_id = $1 AND d.id = $2 AND d.deleted_at IS NULL
		AND EXISTS (SELECT 1 FROM clinics c WHERE c.id = $1 AND c.deleted_at IS NULL)`

	var row dentistRow
	err := r.db.GetContext(ctx, &row, query, clinicID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Dentist{}, ErrNotFound
	}
	if err != nil {
		return Dentist{}, err
	}
	return row.toDentist(), nil
}

func (r *postgresRepository) Update(ctx context.Context, d Dentist) error {
	const query = `
		UPDATE dentists
		SET name = :name, phone = :phone, email = :email, updated_at = :updated_at
		WHERE clinic_id = :clinic_id AND id = :id AND deleted_at IS NULL
		AND EXISTS (SELECT 1 FROM clinics c WHERE c.id = :clinic_id AND c.deleted_at IS NULL)`

	res, err := r.db.NamedExecContext(ctx, query, toRow(d))
	if isUniqueViolation(err) {
		return ErrEmailExists
	}
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *postgresRepository) SoftDelete(ctx context.Context, clinicID, id string, deletedAt time.Time) error {
	const query = `
		UPDATE dentists SET deleted_at = $3, updated_at = $3
		WHERE clinic_id = $1 AND id = $2 AND deleted_at IS NULL
		AND EXISTS (SELECT 1 FROM clinics c WHERE c.id = $1 AND c.deleted_at IS NULL)`

	res, err := r.db.ExecContext(ctx, query, clinicID, id, deletedAt)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *postgresRepository) List(ctx context.Context, clinicID string, params ListParams) (ListResult, error) {
	const listQuery = `
		SELECT * FROM dentists
		WHERE clinic_id = $1 AND deleted_at IS NULL
		AND EXISTS (SELECT 1 FROM clinics c WHERE c.id = $1 AND c.deleted_at IS NULL)
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3`

	var rows []dentistRow
	if err := r.db.SelectContext(ctx, &rows, listQuery, clinicID, params.Limit, params.Offset); err != nil {
		return ListResult{}, err
	}

	const countQuery = `
		SELECT COUNT(*) FROM dentists
		WHERE clinic_id = $1 AND deleted_at IS NULL
		AND EXISTS (SELECT 1 FROM clinics c WHERE c.id = $1 AND c.deleted_at IS NULL)`

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, clinicID); err != nil {
		return ListResult{}, err
	}

	items := make([]Dentist, len(rows))
	for i, row := range rows {
		items[i] = row.toDentist()
	}
	return ListResult{Items: items, Total: total}, nil
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

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

func toRow(d Dentist) dentistRow {
	return dentistRow{
		ID:        d.ID,
		ClinicID:  d.ClinicID,
		Name:      d.Name,
		Phone:     d.Phone,
		Email:     d.Email,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		DeletedAt: d.DeletedAt,
	}
}

func (row dentistRow) toDentist() Dentist {
	return Dentist{
		ID:        row.ID,
		ClinicID:  row.ClinicID,
		Name:      row.Name,
		Phone:     row.Phone,
		Email:     row.Email,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: row.DeletedAt,
	}
}
