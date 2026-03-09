//go:build managedtoken

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type orderedDeleteUserRepoStub struct {
	*managedTokenUserRepoStub
	events *[]string
}

func (s *orderedDeleteUserRepoStub) Delete(ctx context.Context, id int64) error {
	if s.events != nil {
		*s.events = append(*s.events, "delete")
	}
	return s.managedTokenUserRepoStub.Delete(ctx, id)
}

func TestAdminService_DeleteUser_InvalidatesAuthCacheBeforeDelete(t *testing.T) {
	events := make([]string, 0, 2)
	repo := &orderedDeleteUserRepoStub{
		managedTokenUserRepoStub: &managedTokenUserRepoStub{
			user: &User{ID: 9, Email: "managed@example.com", Role: RoleUser},
		},
		events: &events,
	}
	invalidator := &managedTokenAuthCacheInvalidatorStub{events: &events}
	svc := &adminServiceImpl{
		userRepo:             repo,
		authCacheInvalidator: invalidator,
	}

	err := svc.DeleteUser(context.Background(), 9)
	require.NoError(t, err)
	require.Equal(t, []string{"invalidate", "delete"}, events)
	require.Equal(t, []int64{9}, invalidator.userIDs)
}
