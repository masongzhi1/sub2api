//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type accountRepoBatchDeleteStub struct {
	*accountRepoStub
	batchDeletedIDs []int64
	batchDeleteErr  error
}

func (s *accountRepoBatchDeleteStub) BatchDelete(ctx context.Context, ids []int64) ([]int64, error) {
	s.batchDeletedIDs = append([]int64(nil), ids...)
	if s.batchDeleteErr != nil {
		return nil, s.batchDeleteErr
	}
	return append([]int64(nil), ids...), nil
}

func TestAdminService_BatchDeleteAccounts_UsesBatchRepo(t *testing.T) {
	repo := &accountRepoBatchDeleteStub{accountRepoStub: &accountRepoStub{}}
	svc := &adminServiceImpl{accountRepo: repo}

	result, err := svc.BatchDeleteAccounts(context.Background(), []int64{3, 1, 3, 0, -1, 2})
	require.NoError(t, err)
	require.Equal(t, []int64{3, 1, 2}, result.DeletedIDs)
	require.Equal(t, []int64{3, 1, 2}, repo.batchDeletedIDs)
	require.Empty(t, repo.deletedIDs)
}

func TestAdminService_BatchDeleteAccounts_BatchRepoError(t *testing.T) {
	repo := &accountRepoBatchDeleteStub{
		accountRepoStub: &accountRepoStub{},
		batchDeleteErr:  errors.New("batch delete failed"),
	}
	svc := &adminServiceImpl{accountRepo: repo}

	result, err := svc.BatchDeleteAccounts(context.Background(), []int64{1, 2})
	require.Nil(t, result)
	require.EqualError(t, err, "batch delete failed")
	require.Empty(t, repo.deletedIDs)
}

func TestAdminService_BatchDeleteAccounts_FallbackToSingleDelete(t *testing.T) {
	repo := &accountRepoStub{}
	svc := &adminServiceImpl{accountRepo: repo}

	result, err := svc.BatchDeleteAccounts(context.Background(), []int64{5, 0, 5, 4})
	require.NoError(t, err)
	require.Equal(t, []int64{5, 4}, result.DeletedIDs)
	require.Equal(t, []int64{5, 4}, repo.deletedIDs)
}
