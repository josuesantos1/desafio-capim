package clinic

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
	ErrNotFound       = errors.New("clinic: not found")
	ErrDocumentExists = errors.New("clinic: document already exists")
)

type Repository interface {
	// Create returns ErrDocumentExists if the insert violates the
	// partial unique constraint on document (Postgres error 23505) —
	// the database is the source of truth for uniqueness, not a
	// prior lookup.
	Create(ctx context.Context, c Clinic) error
	GetByID(ctx context.Context, id string) (Clinic, error)
	GetByDocument(ctx context.Context, document string) (Clinic, error)
	Update(ctx context.Context, c Clinic) error
	SoftDelete(ctx context.Context, id string, deletedAt time.Time) error
}

type postgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) Repository {
	return &postgresRepository{db: db}
}

type clinicRow struct {
	ID        string     `db:"id"`
	Document  string     `db:"document"`
	LegalName string     `db:"legal_name"`
	TradeName string     `db:"trade_name"`
	Bank      *string    `db:"bank"`
	Agency    *string    `db:"agency"`
	Account   *string    `db:"account"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (r *postgresRepository) Create(ctx context.Context, c Clinic) error {
	const query = `
		INSERT INTO clinics (id, document, legal_name, trade_name, bank, agency, account, created_at, updated_at)
		VALUES (:id, :document, :legal_name, :trade_name, :bank, :agency, :account, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, toRow(c))
	if isUniqueViolation(err) {
		return ErrDocumentExists
	}
	return err
}

func (r *postgresRepository) GetByID(ctx context.Context, id string) (Clinic, error) {
	const query = `SELECT * FROM clinics WHERE id = $1 AND deleted_at IS NULL`
	return r.getOne(ctx, query, id)
}

func (r *postgresRepository) GetByDocument(ctx context.Context, document string) (Clinic, error) {
	const query = `SELECT * FROM clinics WHERE document = $1 AND deleted_at IS NULL`
	return r.getOne(ctx, query, document)
}

func (r *postgresRepository) getOne(ctx context.Context, query string, arg any) (Clinic, error) {
	var row clinicRow
	err := r.db.GetContext(ctx, &row, query, arg)
	if errors.Is(err, sql.ErrNoRows) {
		return Clinic{}, ErrNotFound
	}
	if err != nil {
		return Clinic{}, err
	}
	return row.toClinic(), nil
}

func (r *postgresRepository) Update(ctx context.Context, c Clinic) error {
	const query = `
		UPDATE clinics
		SET legal_name = :legal_name, trade_name = :trade_name, bank = :bank, agency = :agency, account = :account, updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	res, err := r.db.NamedExecContext(ctx, query, toRow(c))
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *postgresRepository) SoftDelete(ctx context.Context, id string, deletedAt time.Time) error {
	const query = `UPDATE clinics SET deleted_at = $2, updated_at = $2 WHERE id = $1 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, query, id, deletedAt)
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

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

func toRow(c Clinic) clinicRow {
	return clinicRow{
		ID:        c.ID,
		Document:  c.Document,
		LegalName: c.LegalName,
		TradeName: c.TradeName,
		Bank:      c.Bank,
		Agency:    c.Agency,
		Account:   c.Account,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		DeletedAt: c.DeletedAt,
	}
}

func (row clinicRow) toClinic() Clinic {
	return Clinic{
		ID:        row.ID,
		Document:  row.Document,
		LegalName: row.LegalName,
		TradeName: row.TradeName,
		Bank:      row.Bank,
		Agency:    row.Agency,
		Account:   row.Account,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: row.DeletedAt,
	}
}
