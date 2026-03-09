//go:build managedtoken

package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type managedTokenUserRepoStub struct {
	user              *User
	getErr            error
	createErr         error
	deleteErr         error
	existsByEmail     bool
	existsByEmailErr  error
	existsByEmailHits []string
	nextCreateID      int64
	deletedIDs        []int64
}

func (s *managedTokenUserRepoStub) Create(ctx context.Context, user *User) error {
	if s.createErr != nil {
		return s.createErr
	}
	if user != nil {
		if user.ID == 0 && s.nextCreateID > 0 {
			user.ID = s.nextCreateID
		}
		clone := *user
		s.user = &clone
	}
	return nil
}

func (s *managedTokenUserRepoStub) GetByID(ctx context.Context, id int64) (*User, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.user == nil {
		return nil, ErrUserNotFound
	}
	return s.user, nil
}

func (s *managedTokenUserRepoStub) GetByEmail(ctx context.Context, email string) (*User, error) {
	panic("unexpected GetByEmail call")
}

func (s *managedTokenUserRepoStub) GetFirstAdmin(ctx context.Context) (*User, error) {
	panic("unexpected GetFirstAdmin call")
}

func (s *managedTokenUserRepoStub) Update(ctx context.Context, user *User) error {
	panic("unexpected Update call")
}

func (s *managedTokenUserRepoStub) Delete(ctx context.Context, id int64) error {
	s.deletedIDs = append(s.deletedIDs, id)
	return s.deleteErr
}

func (s *managedTokenUserRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *managedTokenUserRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, filters UserListFilters) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *managedTokenUserRepoStub) UpdateBalance(ctx context.Context, id int64, amount float64) error {
	panic("unexpected UpdateBalance call")
}

func (s *managedTokenUserRepoStub) DeductBalance(ctx context.Context, id int64, amount float64) error {
	panic("unexpected DeductBalance call")
}

func (s *managedTokenUserRepoStub) UpdateConcurrency(ctx context.Context, id int64, amount int) error {
	panic("unexpected UpdateConcurrency call")
}

func (s *managedTokenUserRepoStub) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	s.existsByEmailHits = append(s.existsByEmailHits, email)
	if s.existsByEmailErr != nil {
		return false, s.existsByEmailErr
	}
	return s.existsByEmail, nil
}

func (s *managedTokenUserRepoStub) RemoveGroupFromAllowedGroups(ctx context.Context, groupID int64) (int64, error) {
	panic("unexpected RemoveGroupFromAllowedGroups call")
}

func (s *managedTokenUserRepoStub) AddGroupToAllowedGroups(ctx context.Context, userID int64, groupID int64) error {
	panic("unexpected AddGroupToAllowedGroups call")
}

func (s *managedTokenUserRepoStub) UpdateTotpSecret(ctx context.Context, userID int64, encryptedSecret *string) error {
	panic("unexpected UpdateTotpSecret call")
}

func (s *managedTokenUserRepoStub) EnableTotp(ctx context.Context, userID int64) error {
	panic("unexpected EnableTotp call")
}

func (s *managedTokenUserRepoStub) DisableTotp(ctx context.Context, userID int64) error {
	panic("unexpected DisableTotp call")
}

type managedTokenAuthCacheInvalidatorStub struct {
	userIDs []int64
	events  *[]string
}

func (s *managedTokenAuthCacheInvalidatorStub) InvalidateAuthCacheByKey(ctx context.Context, key string) {
}

func (s *managedTokenAuthCacheInvalidatorStub) InvalidateAuthCacheByUserID(ctx context.Context, userID int64) {
	s.userIDs = append(s.userIDs, userID)
	if s.events != nil {
		*s.events = append(*s.events, "invalidate")
	}
}

func (s *managedTokenAuthCacheInvalidatorStub) InvalidateAuthCacheByGroupID(ctx context.Context, groupID int64) {
}

type managedTokenUserSubscriptionRepoStub struct {
	subs       []UserSubscription
	getByIDErr error
	deleteErr  error
	deletedIDs []int64
}

func (s *managedTokenUserSubscriptionRepoStub) Create(ctx context.Context, sub *UserSubscription) error {
	panic("unexpected Create call")
}

func (s *managedTokenUserSubscriptionRepoStub) GetByID(ctx context.Context, id int64) (*UserSubscription, error) {
	if s.getByIDErr != nil {
		return nil, s.getByIDErr
	}
	for i := range s.subs {
		if s.subs[i].ID == id {
			clone := s.subs[i]
			return &clone, nil
		}
	}
	return nil, ErrSubscriptionNotFound
}

func (s *managedTokenUserSubscriptionRepoStub) GetByUserIDAndGroupID(ctx context.Context, userID, groupID int64) (*UserSubscription, error) {
	panic("unexpected GetByUserIDAndGroupID call")
}

func (s *managedTokenUserSubscriptionRepoStub) GetActiveByUserIDAndGroupID(ctx context.Context, userID, groupID int64) (*UserSubscription, error) {
	panic("unexpected GetActiveByUserIDAndGroupID call")
}

func (s *managedTokenUserSubscriptionRepoStub) Update(ctx context.Context, sub *UserSubscription) error {
	panic("unexpected Update call")
}

func (s *managedTokenUserSubscriptionRepoStub) Delete(ctx context.Context, id int64) error {
	s.deletedIDs = append(s.deletedIDs, id)
	return s.deleteErr
}

func (s *managedTokenUserSubscriptionRepoStub) ListByUserID(ctx context.Context, userID int64) ([]UserSubscription, error) {
	out := make([]UserSubscription, 0)
	for i := range s.subs {
		if s.subs[i].UserID == userID {
			out = append(out, s.subs[i])
		}
	}
	return out, nil
}

func (s *managedTokenUserSubscriptionRepoStub) ListActiveByUserID(ctx context.Context, userID int64) ([]UserSubscription, error) {
	panic("unexpected ListActiveByUserID call")
}

func (s *managedTokenUserSubscriptionRepoStub) ListByGroupID(ctx context.Context, groupID int64, params pagination.PaginationParams) ([]UserSubscription, *pagination.PaginationResult, error) {
	panic("unexpected ListByGroupID call")
}

func (s *managedTokenUserSubscriptionRepoStub) List(ctx context.Context, params pagination.PaginationParams, userID, groupID *int64, status, sortBy, sortOrder string) ([]UserSubscription, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *managedTokenUserSubscriptionRepoStub) ExistsByUserIDAndGroupID(ctx context.Context, userID, groupID int64) (bool, error) {
	panic("unexpected ExistsByUserIDAndGroupID call")
}

func (s *managedTokenUserSubscriptionRepoStub) ExtendExpiry(ctx context.Context, subscriptionID int64, newExpiresAt time.Time) error {
	panic("unexpected ExtendExpiry call")
}

func (s *managedTokenUserSubscriptionRepoStub) UpdateStatus(ctx context.Context, subscriptionID int64, status string) error {
	panic("unexpected UpdateStatus call")
}

func (s *managedTokenUserSubscriptionRepoStub) UpdateNotes(ctx context.Context, subscriptionID int64, notes string) error {
	panic("unexpected UpdateNotes call")
}

func (s *managedTokenUserSubscriptionRepoStub) ActivateWindows(ctx context.Context, id int64, start time.Time) error {
	panic("unexpected ActivateWindows call")
}

func (s *managedTokenUserSubscriptionRepoStub) ResetDailyUsage(ctx context.Context, id int64, newWindowStart time.Time) error {
	panic("unexpected ResetDailyUsage call")
}

func (s *managedTokenUserSubscriptionRepoStub) ResetWeeklyUsage(ctx context.Context, id int64, newWindowStart time.Time) error {
	panic("unexpected ResetWeeklyUsage call")
}

func (s *managedTokenUserSubscriptionRepoStub) ResetMonthlyUsage(ctx context.Context, id int64, newWindowStart time.Time) error {
	panic("unexpected ResetMonthlyUsage call")
}

func (s *managedTokenUserSubscriptionRepoStub) IncrementUsage(ctx context.Context, id int64, costUSD float64) error {
	panic("unexpected IncrementUsage call")
}

func (s *managedTokenUserSubscriptionRepoStub) BatchUpdateExpiredStatus(ctx context.Context) (int64, error) {
	panic("unexpected BatchUpdateExpiredStatus call")
}
