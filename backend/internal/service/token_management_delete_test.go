//go:build managedtoken

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenManagementService_DeleteManagedToken_Success(t *testing.T) {
	repo := &managedTokenUserRepoStub{
		user: &User{
			ID:    7,
			Email: "a1b2c3d4e5f6@tokens.local",
			Notes: buildManagedTokenNotes("示例令牌", ""),
		},
	}
	subRepo := &managedTokenUserSubscriptionRepoStub{
		subs: []UserSubscription{
			{ID: 101, UserID: 7, GroupID: 11, Status: SubscriptionStatusActive},
			{ID: 102, UserID: 7, GroupID: 12, Status: SubscriptionStatusExpired},
		},
	}
	svc := &TokenManagementService{
		userRepo:            repo,
		subscriptionService: &SubscriptionService{userSubRepo: subRepo},
	}

	err := svc.Delete(context.Background(), 7)
	require.NoError(t, err)
	require.Equal(t, []int64{101, 102}, subRepo.deletedIDs)
	require.Equal(t, []int64{7}, repo.deletedIDs)
}

func TestTokenManagementService_DeleteRejectsNonManagedUser(t *testing.T) {
	repo := &managedTokenUserRepoStub{
		user: &User{
			ID:    8,
			Email: "normal@example.com",
			Notes: "plain note",
		},
	}
	svc := &TokenManagementService{userRepo: repo}

	err := svc.Delete(context.Background(), 8)
	require.Error(t, err)
	require.ErrorContains(t, err, "managed token not found")
	require.Empty(t, repo.deletedIDs)
}
