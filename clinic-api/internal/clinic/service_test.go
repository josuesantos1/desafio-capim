package clinic_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/josuesantos1/desafio/internal/clinic"
	"github.com/josuesantos1/desafio/mocks"
)

func TestService_Create(t *testing.T) {
	tests := []struct {
		name              string
		in                clinic.CreateInput
		setupRepo         func(repo *mocks.Repository)
		wantErr           error // sentinel checked with errors.Is; nil means no error expected
		wantValidationErr bool
	}{
		{
			name: "success",
			in:   clinic.CreateInput{Document: "12345678900", LegalName: "Legal", TradeName: "Trade"},
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
		{
			name: "duplicate document",
			in:   clinic.CreateInput{Document: "12345678900", LegalName: "Legal", TradeName: "Trade"},
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(clinic.ErrDocumentExists).Once()
			},
			wantErr: clinic.ErrDocumentExists,
		},
		{
			name:              "validation error",
			in:                clinic.CreateInput{Document: "123"},
			setupRepo:         func(repo *mocks.Repository) {},
			wantValidationErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewRepository(t)
			tt.setupRepo(repo)
			svc := clinic.NewService(repo)

			c, err := svc.Create(context.Background(), tt.in)

			switch {
			case tt.wantValidationErr:
				var ve *clinic.ValidationError
				require.True(t, errors.As(err, &ve))
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			default:
				require.NoError(t, err)
				assert.NotEmpty(t, c.ID)
			}
		})
	}
}

func TestService_Get(t *testing.T) {
	found := clinic.Clinic{ID: "id-1", Document: "12345678900"}

	tests := []struct {
		name      string
		id        string
		setupRepo func(repo *mocks.Repository)
		wantErr   error
	}{
		{
			name: "success",
			id:   "id-1",
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().GetByID(mock.Anything, "id-1").Return(found, nil).Once()
			},
		},
		{
			name: "not found",
			id:   "missing-id",
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().GetByID(mock.Anything, "missing-id").Return(clinic.Clinic{}, clinic.ErrNotFound).Once()
			},
			wantErr: clinic.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewRepository(t)
			tt.setupRepo(repo)
			svc := clinic.NewService(repo)

			c, err := svc.Get(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, found.ID, c.ID)
		})
	}
}

func TestService_Update(t *testing.T) {
	current := clinic.Clinic{ID: "id-1", Document: "12345678900", LegalName: "Old Legal", TradeName: "Old Trade"}
	newTrade := "New Trade"
	sameDoc := "12345678900"
	otherDoc := "98765432100"
	emptyStr := ""

	tests := []struct {
		name              string
		id                string
		in                clinic.UpdateInput
		setupRepo         func(repo *mocks.Repository)
		wantErr           error
		wantValidationErr bool
		wantTradeName     string
	}{
		{
			name: "partial update",
			id:   "id-1",
			in:   clinic.UpdateInput{TradeName: &newTrade},
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().GetByID(mock.Anything, "id-1").Return(current, nil).Once()
				repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantTradeName: "New Trade",
		},
		{
			name: "empty body is no-op",
			id:   "id-1",
			in:   clinic.UpdateInput{},
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().GetByID(mock.Anything, "id-1").Return(current, nil).Once()
				repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantTradeName: "Old Trade",
		},
		{
			name: "same document is no-op",
			id:   "id-1",
			in:   clinic.UpdateInput{Document: &sameDoc},
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().GetByID(mock.Anything, "id-1").Return(current, nil).Once()
				repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantTradeName: "Old Trade",
		},
		{
			name: "different document is immutable",
			id:   "id-1",
			in:   clinic.UpdateInput{Document: &otherDoc},
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().GetByID(mock.Anything, "id-1").Return(current, nil).Once()
			},
			wantErr: clinic.ErrDocumentImmutable,
		},
		{
			name: "not found",
			id:   "missing-id",
			in:   clinic.UpdateInput{},
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().GetByID(mock.Anything, "missing-id").Return(clinic.Clinic{}, clinic.ErrNotFound).Once()
			},
			wantErr: clinic.ErrNotFound,
		},
		{
			name: "validation error",
			id:   "id-1",
			in:   clinic.UpdateInput{LegalName: &emptyStr},
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().GetByID(mock.Anything, "id-1").Return(current, nil).Once()
			},
			wantValidationErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewRepository(t)
			tt.setupRepo(repo)
			svc := clinic.NewService(repo)

			c, err := svc.Update(context.Background(), tt.id, tt.in)

			switch {
			case tt.wantValidationErr:
				var ve *clinic.ValidationError
				require.True(t, errors.As(err, &ve))
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.wantTradeName, c.TradeName)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupRepo func(repo *mocks.Repository)
		wantErr   error
	}{
		{
			name: "success",
			id:   "id-1",
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().SoftDelete(mock.Anything, "id-1", mock.AnythingOfType("time.Time")).Return(nil).Once()
			},
		},
		{
			name: "not found",
			id:   "missing-id",
			setupRepo: func(repo *mocks.Repository) {
				repo.EXPECT().SoftDelete(mock.Anything, "missing-id", mock.AnythingOfType("time.Time")).Return(clinic.ErrNotFound).Once()
			},
			wantErr: clinic.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewRepository(t)
			tt.setupRepo(repo)
			svc := clinic.NewService(repo)

			err := svc.Delete(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestService_DocumentReusableAfterSoftDelete(t *testing.T) {
	repo := mocks.NewRepository(t)
	svc := clinic.NewService(repo)
	ctx := context.Background()

	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
	first, err := svc.Create(ctx, clinic.CreateInput{Document: "12345678900", LegalName: "First", TradeName: "First"})
	require.NoError(t, err)

	repo.EXPECT().SoftDelete(mock.Anything, first.ID, mock.AnythingOfType("time.Time")).Return(nil).Once()
	require.NoError(t, svc.Delete(ctx, first.ID))

	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
	second, err := svc.Create(ctx, clinic.CreateInput{Document: "12345678900", LegalName: "Second", TradeName: "Second"})
	require.NoError(t, err)
	assert.NotEqual(t, first.ID, second.ID)
}
