package storage_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/josuesantos1/desafio/pkg/storage"
)

func TestStore_InsertRead(t *testing.T) {
	s := storage.New[string]()

	require.NoError(t, s.Insert("id-1", "value-1"))

	v, err := s.Read("id-1")
	require.NoError(t, err)
	require.Equal(t, "value-1", v)
}

func TestStore_Insert_AlreadyExists(t *testing.T) {
	s := storage.New[string]()
	require.NoError(t, s.Insert("id-1", "value-1"))

	err := s.Insert("id-1", "value-2")
	require.ErrorIs(t, err, storage.ErrAlreadyExists)

	v, _ := s.Read("id-1")
	require.Equal(t, "value-1", v, "original value must be unchanged")
}

func TestStore_Read_NotFound(t *testing.T) {
	s := storage.New[string]()

	v, err := s.Read("missing")
	require.ErrorIs(t, err, storage.ErrNotFound)
	require.Equal(t, "", v)
}

func TestStore_Update(t *testing.T) {
	s := storage.New[string]()
	require.NoError(t, s.Insert("id-1", "value-1"))

	require.NoError(t, s.Update("id-1", "value-2"))

	v, err := s.Read("id-1")
	require.NoError(t, err)
	require.Equal(t, "value-2", v)
}

func TestStore_Update_NotFound(t *testing.T) {
	s := storage.New[string]()

	err := s.Update("missing", "value")
	require.ErrorIs(t, err, storage.ErrNotFound)
}

func TestStore_Delete(t *testing.T) {
	s := storage.New[string]()
	require.NoError(t, s.Insert("id-1", "value-1"))

	require.NoError(t, s.Delete("id-1"))

	_, err := s.Read("id-1")
	require.ErrorIs(t, err, storage.ErrNotFound)
}

func TestStore_Delete_NotFound(t *testing.T) {
	s := storage.New[string]()

	err := s.Delete("missing")
	require.ErrorIs(t, err, storage.ErrNotFound)
}

func TestStore_All(t *testing.T) {
	s := storage.New[int]()
	require.NoError(t, s.Insert("a", 1))
	require.NoError(t, s.Insert("b", 2))
	require.NoError(t, s.Insert("c", 3))

	all := s.All()
	require.ElementsMatch(t, []int{1, 2, 3}, all)
}

func TestStore_All_Empty(t *testing.T) {
	s := storage.New[int]()

	require.Empty(t, s.All())
}

func TestStore_ConcurrentAccess(t *testing.T) {
	s := storage.New[int]()

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := string(rune('a' + i%26))
			_ = s.Insert(id, i)
			_, _ = s.Read(id)
			_ = s.Update(id, i+1)
			_ = s.All()
		}(i)
	}
	wg.Wait()
}
