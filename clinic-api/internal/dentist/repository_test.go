package dentist

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

const testClinicIDRepo = "11111111-1111-1111-1111-111111111111"

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
	d := Dentist{ID: "id-1", ClinicID: testClinicIDRepo, Name: "Dr. A", Phone: "123", Email: "a@x.com", CreatedAt: now, UpdatedAt: now}

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO dentists")).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "unique violation maps to ErrEmailExists",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO dentists")).
					WillReturnError(&pgconn.PgError{Code: uniqueViolationCode})
			},
			wantErr: ErrEmailExists,
		},
		{
			name: "clinic inactive (EXISTS guard fails) maps to ErrNotFound",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO dentists")).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockRepo(t)
			tt.setupMock(mock)

			err := repo.Create(context.Background(), d)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresRepository_Create_QueryIncludesClinicActiveGuard(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now().UTC()
	d := Dentist{ID: "id-1", ClinicID: testClinicIDRepo, Name: "Dr. A", Phone: "123", Email: "a@x.com", CreatedAt: now, UpdatedAt: now}

	mock.ExpectExec(regexp.QuoteMeta("EXISTS (SELECT 1 FROM clinics")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Create(context.Background(), d))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_GetByID(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   error
		wantID    string
	}{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "clinic_id", "name", "phone", "email", "created_at", "updated_at", "deleted_at"}).
					AddRow("id-1", testClinicIDRepo, "Dr. A", "123", "a@x.com", now, now, nil)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT d.* FROM dentists d")).
					WithArgs(testClinicIDRepo, "id-1").
					WillReturnRows(rows)
			},
			wantID: "id-1",
		},
		{
			name: "not found (includes: id missing, soft-deleted, wrong clinic, or clinic inactive)",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT d.* FROM dentists d")).
					WithArgs(testClinicIDRepo, "missing-id").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockRepo(t)
			tt.setupMock(mock)

			id := "id-1"
			if tt.wantErr != nil {
				id = "missing-id"
			}

			d, err := repo.GetByID(context.Background(), testClinicIDRepo, id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantID, d.ID)
		})
	}
}

func TestPostgresRepository_GetByID_QueryIncludesClinicActiveGuard(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(regexp.QuoteMeta("EXISTS (SELECT 1 FROM clinics")).
		WithArgs(testClinicIDRepo, "id-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.GetByID(context.Background(), testClinicIDRepo, "id-1")
	require.ErrorIs(t, err, ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Update(t *testing.T) {
	now := time.Now().UTC()
	d := Dentist{ID: "id-1", ClinicID: testClinicIDRepo, Name: "Dr. A", Phone: "123", Email: "a@x.com", UpdatedAt: now}

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta("UPDATE dentists")).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "unique violation maps to ErrEmailExists",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta("UPDATE dentists")).
					WillReturnError(&pgconn.PgError{Code: uniqueViolationCode})
			},
			wantErr: ErrEmailExists,
		},
		{
			name: "not found (row missing, soft-deleted, or clinic inactive)",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta("UPDATE dentists")).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockRepo(t)
			tt.setupMock(mock)

			err := repo.Update(context.Background(), d)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPostgresRepository_Update_QueryIncludesClinicActiveGuard(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now().UTC()
	d := Dentist{ID: "id-1", ClinicID: testClinicIDRepo, Name: "Dr. A", Phone: "123", Email: "a@x.com", UpdatedAt: now}

	mock.ExpectExec(regexp.QuoteMeta("EXISTS (SELECT 1 FROM clinics")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Update(context.Background(), d))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_SoftDelete(t *testing.T) {
	tests := []struct {
		name         string
		rowsAffected int64
		wantErr      error
	}{
		{name: "success", rowsAffected: 1},
		{name: "not found (row missing, already deleted, or clinic inactive)", rowsAffected: 0, wantErr: ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockRepo(t)
			mock.ExpectExec(regexp.QuoteMeta("UPDATE dentists SET deleted_at")).
				WithArgs(testClinicIDRepo, "id-1", sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))

			err := repo.SoftDelete(context.Background(), testClinicIDRepo, "id-1", time.Now().UTC())

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPostgresRepository_SoftDelete_QueryIncludesClinicActiveGuard(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectExec(regexp.QuoteMeta("EXISTS (SELECT 1 FROM clinics")).
		WithArgs(testClinicIDRepo, "id-1", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.SoftDelete(context.Background(), testClinicIDRepo, "id-1", time.Now().UTC()))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_List(t *testing.T) {
	now := time.Now().UTC()
	repo, mock := newMockRepo(t)

	rows := sqlmock.NewRows([]string{"id", "clinic_id", "name", "phone", "email", "created_at", "updated_at", "deleted_at"}).
		AddRow("id-1", testClinicIDRepo, "Dr. A", "123", "a@x.com", now, now, nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM dentists")).
		WithArgs(testClinicIDRepo, 20, 0).
		WillReturnRows(rows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM dentists")).
		WithArgs(testClinicIDRepo).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	result, err := repo.List(context.Background(), testClinicIDRepo, ListParams{Limit: 20, Offset: 0})
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	require.Equal(t, 1, result.Total)
}

func TestPostgresRepository_List_QueriesIncludeClinicActiveGuard(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(regexp.QuoteMeta("EXISTS (SELECT 1 FROM clinics")).
		WithArgs(testClinicIDRepo, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(regexp.QuoteMeta("EXISTS (SELECT 1 FROM clinics")).
		WithArgs(testClinicIDRepo).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	result, err := repo.List(context.Background(), testClinicIDRepo, ListParams{Limit: 20, Offset: 0})
	require.NoError(t, err)
	require.Empty(t, result.Items)
	require.NoError(t, mock.ExpectationsWereMet())
}
