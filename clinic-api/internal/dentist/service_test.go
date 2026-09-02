package dentist_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/josuesantos1/desafio/internal/clinic"
	"github.com/josuesantos1/desafio/internal/dentist"
	"github.com/josuesantos1/desafio/mocks"
)

const testClinicID = "11111111-1111-1111-1111-111111111111"

func expectClinicActive(clinicRepo *mocks.Repository) {
	clinicRepo.EXPECT().GetByID(mock.Anything, testClinicID).Return(clinic.Clinic{ID: testClinicID}, nil).Once()
}

func expectClinicNotFound(clinicRepo *mocks.Repository) {
	clinicRepo.EXPECT().GetByID(mock.Anything, testClinicID).Return(clinic.Clinic{}, clinic.ErrNotFound).Once()
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name              string
		in                dentist.CreateInput
		setupRepo         func(repo *mocks.DentistRepository)
		setupClinic       func(clinicRepo *mocks.Repository)
		wantErr           error
		wantValidationErr bool
	}{
		{
			name: "success",
			in:   dentist.CreateInput{Name: "Dr. A", Phone: "123", Email: "a@x.com"},
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
			},
			setupClinic: expectClinicActive,
		},
		{
			name:        "clinic not found",
			in:          dentist.CreateInput{Name: "Dr. A", Phone: "123", Email: "a@x.com"},
			setupRepo:   func(repo *mocks.DentistRepository) {},
			setupClinic: expectClinicNotFound,
			wantErr:     dentist.ErrClinicNotFound,
		},
		{
			name: "email exists",
			in:   dentist.CreateInput{Name: "Dr. A", Phone: "123", Email: "a@x.com"},
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(dentist.ErrEmailExists).Once()
			},
			setupClinic: expectClinicActive,
			wantErr:     dentist.ErrEmailExists,
		},
		{
			name:              "validation error",
			in:                dentist.CreateInput{Name: ""},
			setupRepo:         func(repo *mocks.DentistRepository) {},
			setupClinic:       expectClinicActive,
			wantValidationErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewDentistRepository(t)
			clinicRepo := mocks.NewRepository(t)
			tt.setupRepo(repo)
			tt.setupClinic(clinicRepo)
			svc := dentist.NewService(repo, clinicRepo)

			d, err := svc.Create(context.Background(), testClinicID, tt.in)

			switch {
			case tt.wantValidationErr:
				var ve *dentist.ValidationError
				require.True(t, errors.As(err, &ve))
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			default:
				require.NoError(t, err)
				assert.NotEmpty(t, d.ID)
				assert.Equal(t, testClinicID, d.ClinicID)
			}
		})
	}
}

func TestService_Get(t *testing.T) {
	found := dentist.Dentist{ID: "dentist-1", ClinicID: testClinicID}

	tests := []struct {
		name        string
		id          string
		setupRepo   func(repo *mocks.DentistRepository)
		setupClinic func(clinicRepo *mocks.Repository)
		wantErr     error
	}{
		{
			name: "success",
			id:   "dentist-1",
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().GetByID(mock.Anything, testClinicID, "dentist-1").Return(found, nil).Once()
			},
			setupClinic: expectClinicActive,
		},
		{
			name:        "clinic not found",
			id:          "dentist-1",
			setupRepo:   func(repo *mocks.DentistRepository) {},
			setupClinic: expectClinicNotFound,
			wantErr:     dentist.ErrClinicNotFound,
		},
		{
			name: "dentist not found",
			id:   "missing-id",
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().GetByID(mock.Anything, testClinicID, "missing-id").Return(dentist.Dentist{}, dentist.ErrNotFound).Once()
			},
			setupClinic: expectClinicActive,
			wantErr:     dentist.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewDentistRepository(t)
			clinicRepo := mocks.NewRepository(t)
			tt.setupRepo(repo)
			tt.setupClinic(clinicRepo)
			svc := dentist.NewService(repo, clinicRepo)

			d, err := svc.Get(context.Background(), testClinicID, tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, found.ID, d.ID)
		})
	}
}

func TestService_Update(t *testing.T) {
	current := dentist.Dentist{ID: "dentist-1", ClinicID: testClinicID, Name: "Old", Phone: "111", Email: "old@x.com"}
	newName := "New"
	empty := ""

	tests := []struct {
		name              string
		id                string
		in                dentist.UpdateInput
		setupRepo         func(repo *mocks.DentistRepository)
		setupClinic       func(clinicRepo *mocks.Repository)
		wantErr           error
		wantValidationErr bool
		wantName          string
	}{
		{
			name: "partial update",
			id:   "dentist-1",
			in:   dentist.UpdateInput{Name: &newName},
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().GetByID(mock.Anything, testClinicID, "dentist-1").Return(current, nil).Once()
				repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()
			},
			setupClinic: expectClinicActive,
			wantName:    "New",
		},
		{
			name: "empty body is no-op",
			id:   "dentist-1",
			in:   dentist.UpdateInput{},
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().GetByID(mock.Anything, testClinicID, "dentist-1").Return(current, nil).Once()
				repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()
			},
			setupClinic: expectClinicActive,
			wantName:    "Old",
		},
		{
			name:        "clinic not found",
			id:          "dentist-1",
			in:          dentist.UpdateInput{},
			setupRepo:   func(repo *mocks.DentistRepository) {},
			setupClinic: expectClinicNotFound,
			wantErr:     dentist.ErrClinicNotFound,
		},
		{
			name: "email exists",
			id:   "dentist-1",
			in:   dentist.UpdateInput{Name: &newName},
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().GetByID(mock.Anything, testClinicID, "dentist-1").Return(current, nil).Once()
				repo.EXPECT().Update(mock.Anything, mock.Anything).Return(dentist.ErrEmailExists).Once()
			},
			setupClinic: expectClinicActive,
			wantErr:     dentist.ErrEmailExists,
		},
		{
			name: "not found",
			id:   "missing-id",
			in:   dentist.UpdateInput{},
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().GetByID(mock.Anything, testClinicID, "missing-id").Return(dentist.Dentist{}, dentist.ErrNotFound).Once()
			},
			setupClinic: expectClinicActive,
			wantErr:     dentist.ErrNotFound,
		},
		{
			name: "validation error",
			id:   "dentist-1",
			in:   dentist.UpdateInput{Name: &empty},
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().GetByID(mock.Anything, testClinicID, "dentist-1").Return(current, nil).Once()
			},
			setupClinic:       expectClinicActive,
			wantValidationErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewDentistRepository(t)
			clinicRepo := mocks.NewRepository(t)
			tt.setupRepo(repo)
			tt.setupClinic(clinicRepo)
			svc := dentist.NewService(repo, clinicRepo)

			d, err := svc.Update(context.Background(), testClinicID, tt.id, tt.in)

			switch {
			case tt.wantValidationErr:
				var ve *dentist.ValidationError
				require.True(t, errors.As(err, &ve))
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.wantName, d.Name)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		setupRepo   func(repo *mocks.DentistRepository)
		setupClinic func(clinicRepo *mocks.Repository)
		wantErr     error
	}{
		{
			name: "success",
			id:   "dentist-1",
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().SoftDelete(mock.Anything, testClinicID, "dentist-1", mock.AnythingOfType("time.Time")).Return(nil).Once()
			},
			setupClinic: expectClinicActive,
		},
		{
			name:        "clinic not found",
			id:          "dentist-1",
			setupRepo:   func(repo *mocks.DentistRepository) {},
			setupClinic: expectClinicNotFound,
			wantErr:     dentist.ErrClinicNotFound,
		},
		{
			name: "not found",
			id:   "missing-id",
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().SoftDelete(mock.Anything, testClinicID, "missing-id", mock.AnythingOfType("time.Time")).Return(dentist.ErrNotFound).Once()
			},
			setupClinic: expectClinicActive,
			wantErr:     dentist.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewDentistRepository(t)
			clinicRepo := mocks.NewRepository(t)
			tt.setupRepo(repo)
			tt.setupClinic(clinicRepo)
			svc := dentist.NewService(repo, clinicRepo)

			err := svc.Delete(context.Background(), testClinicID, tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestService_List(t *testing.T) {
	tests := []struct {
		name        string
		params      dentist.ListParams
		setupRepo   func(repo *mocks.DentistRepository)
		setupClinic func(clinicRepo *mocks.Repository)
		wantErr     error
		wantTotal   int
	}{
		{
			name:   "success",
			params: dentist.ListParams{Limit: 20, Offset: 0},
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().List(mock.Anything, testClinicID, mock.MatchedBy(func(p dentist.ListParams) bool {
					return p.Limit == 20 && p.Offset == 0
				})).Return(dentist.ListResult{Items: []dentist.Dentist{{ID: "d1"}}, Total: 1}, nil).Once()
			},
			setupClinic: expectClinicActive,
			wantTotal:   1,
		},
		{
			name:        "clinic not found",
			params:      dentist.ListParams{Limit: 20, Offset: 0},
			setupRepo:   func(repo *mocks.DentistRepository) {},
			setupClinic: expectClinicNotFound,
			wantErr:     dentist.ErrClinicNotFound,
		},
		{
			name:   "clamps out-of-range limit and offset",
			params: dentist.ListParams{Limit: 1000, Offset: -5},
			setupRepo: func(repo *mocks.DentistRepository) {
				repo.EXPECT().List(mock.Anything, testClinicID, mock.MatchedBy(func(p dentist.ListParams) bool {
					return p.Limit == 100 && p.Offset == 0
				})).Return(dentist.ListResult{}, nil).Once()
			},
			setupClinic: expectClinicActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewDentistRepository(t)
			clinicRepo := mocks.NewRepository(t)
			tt.setupRepo(repo)
			tt.setupClinic(clinicRepo)
			svc := dentist.NewService(repo, clinicRepo)

			result, err := svc.List(context.Background(), testClinicID, tt.params)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, result.Total)
		})
	}
}
