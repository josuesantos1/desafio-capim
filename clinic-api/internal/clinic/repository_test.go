package clinic

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func newMockRepo(t *testing.T) (*postgresRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "postgres")
	return &postgresRepository{db: sqlxDB}, mock
}

func TestPostgresRepository_Create(t *testing.T) {
	now := time.Now().UTC()
	c := Clinic{ID: "id-1", Document: "12345678900", LegalName: "Legal", TradeName: "Trade", CreatedAt: now, UpdatedAt: now}

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO clinics")).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "unique violation maps to ErrDocumentExists",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO clinics")).
					WillReturnError(&pgconn.PgError{Code: uniqueViolationCode})
			},
			wantErr: ErrDocumentExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockRepo(t)
			tt.setupMock(mock)

			err := repo.Create(context.Background(), c)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresRepository_GetByID(t *testing.T) {
	now := time.Now().UTC()
	const query = "SELECT * FROM clinics WHERE id = $1 AND deleted_at IS NULL"

	tests := []struct {
		name      string
		id        string
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   error
		wantID    string
	}{
		{
			name: "success",
			id:   "id-1",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "document", "legal_name", "trade_name", "bank", "agency", "account", "created_at", "updated_at", "deleted_at"}).
					AddRow("id-1", "12345678900", "Legal", "Trade", nil, nil, nil, now, now, nil)
				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("id-1").WillReturnRows(rows)
			},
			wantID: "id-1",
		},
		{
			name: "not found",
			id:   "missing-id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("missing-id").WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockRepo(t)
			tt.setupMock(mock)

			c, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantID, c.ID)
		})
	}
}

func TestPostgresRepository_Update(t *testing.T) {
	now := time.Now().UTC()
	c := Clinic{ID: "id-1", Document: "12345678900", LegalName: "Legal", TradeName: "Trade", UpdatedAt: now}

	tests := []struct {
		name         string
		rowsAffected int64
		wantErr      error
	}{
		{name: "success", rowsAffected: 1},
		{name: "not found", rowsAffected: 0, wantErr: ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockRepo(t)
			mock.ExpectExec(regexp.QuoteMeta("UPDATE clinics")).
				WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))

			err := repo.Update(context.Background(), c)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPostgresRepository_SoftDelete(t *testing.T) {
	const query = "UPDATE clinics SET deleted_at = $2, updated_at = $2 WHERE id = $1 AND deleted_at IS NULL"

	tests := []struct {
		name         string
		id           string
		rowsAffected int64
		wantErr      error
	}{
		{name: "success", id: "id-1", rowsAffected: 1},
		{name: "not found", id: "missing-id", rowsAffected: 0, wantErr: ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockRepo(t)
			mock.ExpectExec(regexp.QuoteMeta(query)).
				WithArgs(tt.id, sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))

			err := repo.SoftDelete(context.Background(), tt.id, time.Now().UTC())

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
