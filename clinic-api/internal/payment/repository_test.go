package payment

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
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

	tests := []struct {
		name         string
		payment      Payment
		rowsAffected int64
		wantErr      error
	}{
		{
			name:         "success without dentist",
			payment:      Payment{ID: "id-1", ClinicID: testClinicIDRepo, AmountCents: 15000, Status: StatusPending, PixCode: "code", CreatedAt: now, UpdatedAt: now},
			rowsAffected: 1,
		},
		{
			name:         "clinic inactive (EXISTS guard fails)",
			payment:      Payment{ID: "id-1", ClinicID: testClinicIDRepo, AmountCents: 15000, Status: StatusPending, PixCode: "code", CreatedAt: now, UpdatedAt: now},
			rowsAffected: 0,
			wantErr:      ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockRepo(t)
			mock.ExpectExec(regexp.QuoteMeta("INSERT INTO payments")).
				WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))

			err := repo.Create(context.Background(), tt.payment)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresRepository_Create_QueryIncludesClinicAndDentistGuards(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now().UTC()
	dentistID := "22222222-2222-2222-2222-222222222222"
	p := Payment{ID: "id-1", ClinicID: testClinicIDRepo, DentistID: &dentistID, AmountCents: 15000, Status: StatusPending, PixCode: "code", CreatedAt: now, UpdatedAt: now}

	mock.ExpectExec(regexp.QuoteMeta("EXISTS (SELECT 1 FROM clinics")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Create(context.Background(), p))
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
				rows := sqlmock.NewRows([]string{"id", "clinic_id", "dentist_id", "amount_cents", "status", "pix_code", "created_at", "updated_at", "approved_at"}).
					AddRow("id-1", testClinicIDRepo, nil, 15000, StatusPending, "code", now, now, nil)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM payments")).WithArgs("id-1").WillReturnRows(rows)
			},
			wantID: "id-1",
		},
		{
			name: "not found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM payments")).WithArgs("missing-id").WillReturnRows(sqlmock.NewRows([]string{"id"}))
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

			p, err := repo.GetByID(context.Background(), id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantID, p.ID)
		})
	}
}

func TestPostgresRepository_Approve(t *testing.T) {
	tests := []struct {
		name         string
		rowsAffected int64
		wantErr      error
	}{
		{name: "success", rowsAffected: 1},
		{name: "already approved or missing", rowsAffected: 0, wantErr: ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockRepo(t)
			mock.ExpectExec(regexp.QuoteMeta("UPDATE payments SET status = 'approved'")).
				WithArgs("id-1", sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))

			err := repo.Approve(context.Background(), "id-1", time.Now().UTC())

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
