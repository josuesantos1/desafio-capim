package dentist

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
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
//
// UpdateRoles and SoftDelete additionally protect an active clinic's
// last administrator/legal representative. That guard needs more than
// a single statement's snapshot to be race-free against two concurrent
// operations on different dentists of the same clinic, so both methods
// run inside an explicit sqlx.Tx that locks the clinic row first (see
// lockClinic) — the first use of an explicit transaction in this
// codebase; see committee-dentist-roles.md for why.
type Repository interface {
	Create(ctx context.Context, d Dentist) error
	GetByID(ctx context.Context, clinicID, id string) (Dentist, error)
	Update(ctx context.Context, d Dentist) error
	UpdateRoles(ctx context.Context, clinicID, id string, in RolesInput) (Dentist, error)
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
	ID                    string     `db:"id"`
	ClinicID              string     `db:"clinic_id"`
	Name                  string     `db:"name"`
	Phone                 string     `db:"phone"`
	Email                 string     `db:"email"`
	IsAdministrator       bool       `db:"is_administrator"`
	IsLegalRepresentative bool       `db:"is_legal_representative"`
	CreatedAt             time.Time  `db:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at"`
	DeletedAt             *time.Time `db:"deleted_at"`
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

// lockClinic locks the clinic row for the duration of the enclosing
// transaction, serializing concurrent role/delete operations on the
// same clinic, and reports whether it's currently active. Every
// statement run afterward, in the same tx, takes its own fresh
// snapshot (READ COMMITTED) — so it already reflects any commit that
// was waiting behind this lock.
func lockClinic(ctx context.Context, tx *sqlx.Tx, clinicID string) (active bool, err error) {
	const query = `SELECT status FROM clinics WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`

	var status string
	err = tx.GetContext(ctx, &status, query, clinicID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, ErrClinicNotFound
	}
	if err != nil {
		return false, err
	}
	return status == "active", nil
}

func (r *postgresRepository) UpdateRoles(ctx context.Context, clinicID, id string, in RolesInput) (Dentist, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return Dentist{}, err
	}
	defer tx.Rollback()

	clinicActive, err := lockClinic(ctx, tx, clinicID)
	if err != nil {
		return Dentist{}, err
	}

	now := time.Now().UTC()
	const query = `
		UPDATE dentists
		SET is_administrator = COALESCE(CAST($3 AS boolean), is_administrator),
			is_legal_representative = COALESCE(CAST($4 AS boolean), is_legal_representative),
			updated_at = $6
		WHERE clinic_id = $1 AND id = $2 AND deleted_at IS NULL
		AND (NOT CAST($5 AS boolean) OR CAST($3 AS boolean) IS NULL OR CAST($3 AS boolean) = true OR is_administrator = false
			OR EXISTS (SELECT 1 FROM dentists o WHERE o.clinic_id = $1 AND o.id <> $2
				AND o.deleted_at IS NULL AND o.is_administrator = true))
		AND (NOT CAST($5 AS boolean) OR CAST($4 AS boolean) IS NULL OR CAST($4 AS boolean) = true OR is_legal_representative = false
			OR EXISTS (SELECT 1 FROM dentists o WHERE o.clinic_id = $1 AND o.id <> $2
				AND o.deleted_at IS NULL AND o.is_legal_representative = true))
		RETURNING *`

	var row dentistRow
	err = tx.GetContext(ctx, &row, query, clinicID, id, in.IsAdministrator, in.IsLegalRepresentative, clinicActive, now)
	if errors.Is(err, sql.ErrNoRows) {
		return Dentist{}, diagnoseRolesFailure(ctx, tx, clinicID, id, in, clinicActive)
	}
	if err != nil {
		return Dentist{}, err
	}

	const recompute = `
		UPDATE clinics SET status = 'active', updated_at = $2
		WHERE id = $1 AND status = 'pending'
		AND EXISTS (SELECT 1 FROM dentists WHERE clinic_id = $1 AND deleted_at IS NULL AND is_administrator = true)
		AND EXISTS (SELECT 1 FROM dentists WHERE clinic_id = $1 AND deleted_at IS NULL AND is_legal_representative = true)`

	if _, err := tx.ExecContext(ctx, recompute, clinicID, now); err != nil {
		return Dentist{}, err
	}

	if err := tx.Commit(); err != nil {
		return Dentist{}, err
	}
	return row.toDentist(), nil
}

// diagnoseRolesFailure re-reads, inside the same tx/snapshot as the
// failed guarded UPDATE, to classify a zero-rows-affected result as
// not-found vs. which guard tripped. Priority: administrator checked
// before legal representative (committee decision — errorResponse
// carries a single error code per response).
func diagnoseRolesFailure(ctx context.Context, tx *sqlx.Tx, clinicID, id string, in RolesInput, clinicActive bool) error {
	var current dentistRow
	const getQuery = `SELECT * FROM dentists WHERE clinic_id = $1 AND id = $2 AND deleted_at IS NULL`
	if err := tx.GetContext(ctx, &current, getQuery, clinicID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if !clinicActive {
		return ErrNotFound
	}

	if in.IsAdministrator != nil && !*in.IsAdministrator && current.IsAdministrator {
		hasOther, err := hasOtherActiveWithFlag(ctx, tx, clinicID, id, "is_administrator")
		if err != nil {
			return err
		}
		if !hasOther {
			return ErrLastAdminRequired
		}
	}
	if in.IsLegalRepresentative != nil && !*in.IsLegalRepresentative && current.IsLegalRepresentative {
		hasOther, err := hasOtherActiveWithFlag(ctx, tx, clinicID, id, "is_legal_representative")
		if err != nil {
			return err
		}
		if !hasOther {
			return ErrLastLegalRepresentativeRequired
		}
	}
	return ErrNotFound
}

func hasOtherActiveWithFlag(ctx context.Context, tx *sqlx.Tx, clinicID, id, flagColumn string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM dentists WHERE clinic_id = $1 AND id <> $2 AND deleted_at IS NULL AND ` + flagColumn + ` = true)`
	var exists bool
	err := tx.GetContext(ctx, &exists, query, clinicID, id)
	return exists, err
}

func (r *postgresRepository) SoftDelete(ctx context.Context, clinicID, id string, deletedAt time.Time) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	clinicActive, err := lockClinic(ctx, tx, clinicID)
	if err != nil {
		return err
	}

	const query = `
		UPDATE dentists SET deleted_at = $3, updated_at = $3
		WHERE clinic_id = $1 AND id = $2 AND deleted_at IS NULL
		AND (NOT $4 OR is_administrator = false
			OR EXISTS (SELECT 1 FROM dentists o WHERE o.clinic_id = $1 AND o.id <> $2
				AND o.deleted_at IS NULL AND o.is_administrator = true))
		AND (NOT $4 OR is_legal_representative = false
			OR EXISTS (SELECT 1 FROM dentists o WHERE o.clinic_id = $1 AND o.id <> $2
				AND o.deleted_at IS NULL AND o.is_legal_representative = true))`

	res, err := tx.ExecContext(ctx, query, clinicID, id, deletedAt, clinicActive)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return diagnoseDeleteFailure(ctx, tx, clinicID, id, clinicActive)
	}
	return tx.Commit()
}

func diagnoseDeleteFailure(ctx context.Context, tx *sqlx.Tx, clinicID, id string, clinicActive bool) error {
	var current dentistRow
	const getQuery = `SELECT * FROM dentists WHERE clinic_id = $1 AND id = $2 AND deleted_at IS NULL`
	if err := tx.GetContext(ctx, &current, getQuery, clinicID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if !clinicActive {
		return ErrNotFound
	}

	if current.IsAdministrator {
		hasOther, err := hasOtherActiveWithFlag(ctx, tx, clinicID, id, "is_administrator")
		if err != nil {
			return err
		}
		if !hasOther {
			return ErrLastAdminRequired
		}
	}
	if current.IsLegalRepresentative {
		hasOther, err := hasOtherActiveWithFlag(ctx, tx, clinicID, id, "is_legal_representative")
		if err != nil {
			return err
		}
		if !hasOther {
			return ErrLastLegalRepresentativeRequired
		}
	}
	return ErrNotFound
}

func (r *postgresRepository) List(ctx context.Context, clinicID string, params ListParams) (ListResult, error) {
	var whereExtra strings.Builder
	args := []any{clinicID}

	if params.IsAdministrator != nil {
		args = append(args, *params.IsAdministrator)
		whereExtra.WriteString(" AND is_administrator = $" + strconv.Itoa(len(args)))
	}
	if params.IsLegalRepresentative != nil {
		args = append(args, *params.IsLegalRepresentative)
		whereExtra.WriteString(" AND is_legal_representative = $" + strconv.Itoa(len(args)))
	}

	baseWhere := `clinic_id = $1 AND deleted_at IS NULL
		AND EXISTS (SELECT 1 FROM clinics c WHERE c.id = $1 AND c.deleted_at IS NULL)` + whereExtra.String()

	listArgs := append(append([]any{}, args...), params.Limit, params.Offset)
	listQuery := `SELECT * FROM dentists WHERE ` + baseWhere +
		` ORDER BY created_at ASC LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)

	var rows []dentistRow
	if err := r.db.SelectContext(ctx, &rows, listQuery, listArgs...); err != nil {
		return ListResult{}, err
	}

	countQuery := `SELECT COUNT(*) FROM dentists WHERE ` + baseWhere

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
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
		ID:                    d.ID,
		ClinicID:              d.ClinicID,
		Name:                  d.Name,
		Phone:                 d.Phone,
		Email:                 d.Email,
		IsAdministrator:       d.IsAdministrator,
		IsLegalRepresentative: d.IsLegalRepresentative,
		CreatedAt:             d.CreatedAt,
		UpdatedAt:             d.UpdatedAt,
		DeletedAt:             d.DeletedAt,
	}
}

func (row dentistRow) toDentist() Dentist {
	return Dentist{
		ID:                    row.ID,
		ClinicID:              row.ClinicID,
		Name:                  row.Name,
		Phone:                 row.Phone,
		Email:                 row.Email,
		IsAdministrator:       row.IsAdministrator,
		IsLegalRepresentative: row.IsLegalRepresentative,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
		DeletedAt:             row.DeletedAt,
	}
}
