//go:build managedtoken

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type managedTokenCreateGroupRepoStub struct {
	groupRepoNoop
	group *Group
}

func (s *managedTokenCreateGroupRepoStub) GetByID(context.Context, int64) (*Group, error) {
	return s.group, nil
}

type managedTokenCreateUserSubRepoStub struct {
	*subscriptionUserSubRepoStub
}

func newManagedTokenCreateUserSubRepoStub() *managedTokenCreateUserSubRepoStub {
	return &managedTokenCreateUserSubRepoStub{subscriptionUserSubRepoStub: newSubscriptionUserSubRepoStub()}
}

func (s *managedTokenCreateUserSubRepoStub) GetActiveByUserIDAndGroupID(ctx context.Context, userID, groupID int64) (*UserSubscription, error) {
	sub, err := s.GetByUserIDAndGroupID(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if sub.Status != SubscriptionStatusActive || time.Now().After(sub.ExpiresAt) {
		return nil, ErrSubscriptionNotFound
	}
	return sub, nil
}

type managedTokenCreateAPIKeyRepoStub struct {
	created *APIKey
	exists  bool
}

func (s *managedTokenCreateAPIKeyRepoStub) Create(_ context.Context, key *APIKey) error {
	if key != nil {
		clone := *key
		if clone.ID == 0 {
			clone.ID = 1
			key.ID = 1
		}
		s.created = &clone
	}
	return nil
}

func (s *managedTokenCreateAPIKeyRepoStub) ExistsByKey(context.Context, string) (bool, error) {
	return s.exists, nil
}

func (s *managedTokenCreateAPIKeyRepoStub) GetByID(context.Context, int64) (*APIKey, error) {
	panic("unexpected GetByID call")
}
func (s *managedTokenCreateAPIKeyRepoStub) GetKeyAndOwnerID(context.Context, int64) (string, int64, error) {
	panic("unexpected GetKeyAndOwnerID call")
}
func (s *managedTokenCreateAPIKeyRepoStub) GetByKey(context.Context, string) (*APIKey, error) {
	panic("unexpected GetByKey call")
}
func (s *managedTokenCreateAPIKeyRepoStub) GetByKeyForAuth(context.Context, string) (*APIKey, error) {
	panic("unexpected GetByKeyForAuth call")
}
func (s *managedTokenCreateAPIKeyRepoStub) Update(context.Context, *APIKey) error {
	panic("unexpected Update call")
}
func (s *managedTokenCreateAPIKeyRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}
func (s *managedTokenCreateAPIKeyRepoStub) ListByUserID(context.Context, int64, pagination.PaginationParams, APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByUserID call")
}
func (s *managedTokenCreateAPIKeyRepoStub) VerifyOwnership(context.Context, int64, []int64) ([]int64, error) {
	panic("unexpected VerifyOwnership call")
}
func (s *managedTokenCreateAPIKeyRepoStub) CountByUserID(context.Context, int64) (int64, error) {
	panic("unexpected CountByUserID call")
}
func (s *managedTokenCreateAPIKeyRepoStub) ListByGroupID(context.Context, int64, pagination.PaginationParams) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByGroupID call")
}
func (s *managedTokenCreateAPIKeyRepoStub) SearchAPIKeys(context.Context, int64, string, int) ([]APIKey, error) {
	panic("unexpected SearchAPIKeys call")
}
func (s *managedTokenCreateAPIKeyRepoStub) ClearGroupIDByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected ClearGroupIDByGroupID call")
}
func (s *managedTokenCreateAPIKeyRepoStub) CountByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected CountByGroupID call")
}
func (s *managedTokenCreateAPIKeyRepoStub) ListKeysByUserID(context.Context, int64) ([]string, error) {
	return nil, nil
}
func (s *managedTokenCreateAPIKeyRepoStub) ListKeysByGroupID(context.Context, int64) ([]string, error) {
	return nil, nil
}
func (s *managedTokenCreateAPIKeyRepoStub) IncrementQuotaUsed(context.Context, int64, float64) (float64, error) {
	panic("unexpected IncrementQuotaUsed call")
}
func (s *managedTokenCreateAPIKeyRepoStub) UpdateLastUsed(context.Context, int64, time.Time) error {
	panic("unexpected UpdateLastUsed call")
}
func (s *managedTokenCreateAPIKeyRepoStub) IncrementRateLimitUsage(context.Context, int64, float64) error {
	panic("unexpected IncrementRateLimitUsage call")
}
func (s *managedTokenCreateAPIKeyRepoStub) ResetRateLimitWindows(context.Context, int64) error {
	panic("unexpected ResetRateLimitWindows call")
}
func (s *managedTokenCreateAPIKeyRepoStub) GetRateLimitData(context.Context, int64) (*APIKeyRateLimitData, error) {
	panic("unexpected GetRateLimitData call")
}

func TestTokenManagementService_CreateUsesUsernameAsEmail(t *testing.T) {
	repo := &managedTokenUserRepoStub{nextCreateID: 42}
	groupRepo := &managedTokenCreateGroupRepoStub{group: &Group{ID: 11, SubscriptionType: SubscriptionTypeSubscription}}
	subRepo := newManagedTokenCreateUserSubRepoStub()
	apiKeyRepo := &managedTokenCreateAPIKeyRepoStub{}
	apiKeyService := &APIKeyService{
		apiKeyRepo:  apiKeyRepo,
		userRepo:    repo,
		groupRepo:   groupRepo,
		userSubRepo: subRepo,
	}
	svc := &TokenManagementService{
		userRepo:            repo,
		subscriptionService: &SubscriptionService{groupRepo: groupRepo, userSubRepo: subRepo},
		apiKeyService:       apiKeyService,
	}
	customKey := "custom_key_123456"

	token, err := svc.Create(context.Background(), &CreateManagedTokenInput{
		Name:         " 示例令牌 ",
		GroupID:      11,
		ValidityDays: 30,
		CustomKey:    &customKey,
		Notes:        "note",
		AssignedBy:   9,
	})
	require.NoError(t, err)
	require.NotNil(t, token)
	require.Equal(t, "示例令牌@tokens.local", token.User.Username)
	require.Equal(t, token.User.Username, token.User.Email)
	require.Equal(t, []string{token.User.Username}, repo.existsByEmailHits)
	require.NotNil(t, token.Subscription)
	require.Equal(t, int64(11), token.Subscription.GroupID)
	require.NotNil(t, token.APIKey)
	require.Equal(t, "Token | 示例令牌", token.APIKey.Name)
	require.Equal(t, customKey, token.APIKey.Key)
}

func TestTokenManagementService_CreateRejectsDuplicateUsernameEmail(t *testing.T) {
	repo := &managedTokenUserRepoStub{existsByEmail: true}
	svc := &TokenManagementService{userRepo: repo}
	customKey := "custom_key_123456"

	token, err := svc.Create(context.Background(), &CreateManagedTokenInput{
		Name:      "示例令牌",
		GroupID:   11,
		CustomKey: &customKey,
	})
	require.Nil(t, token)
	require.Error(t, err)
	require.ErrorContains(t, err, "email already exists")
	require.Equal(t, []string{"示例令牌@tokens.local"}, repo.existsByEmailHits)
	require.Nil(t, repo.user)
}
