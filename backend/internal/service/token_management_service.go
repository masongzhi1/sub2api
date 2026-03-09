package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	managedTokenEmailDomain    = "tokens.local"
	managedTokenNoteMarker     = "[token-management]"
	managedTokenKeyNamePrefix  = "Token | "
	maxTokenManagedListFetch   = 500
	managedTokenUsernameMaxLen = 100
)

type ManagedToken struct {
	Label        string
	User         *User
	APIKey       *APIKey
	Subscription *UserSubscription
}

type ListManagedTokensFilters struct {
	Search string
}

type CreateManagedTokenInput struct {
	Name         string
	GroupID      int64
	ValidityDays int
	CustomKey    *string
	Notes        string
	AssignedBy   int64
}

type TokenManagementService struct {
	userRepo            UserRepository
	subscriptionService *SubscriptionService
	apiKeyService       *APIKeyService
	settingService      *SettingService
	cfg                 *config.Config
}

func NewTokenManagementService(
	userRepo UserRepository,
	subscriptionService *SubscriptionService,
	apiKeyService *APIKeyService,
	settingService *SettingService,
	cfg *config.Config,
) *TokenManagementService {
	return &TokenManagementService{
		userRepo:            userRepo,
		subscriptionService: subscriptionService,
		apiKeyService:       apiKeyService,
		settingService:      settingService,
		cfg:                 cfg,
	}
}

func (s *TokenManagementService) List(ctx context.Context, page, pageSize int, filters ListManagedTokensFilters) ([]ManagedToken, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	includeSubscriptions := false
	users, _, err := s.userRepo.ListWithFilters(ctx, pagination.PaginationParams{
		Page:     1,
		PageSize: maxTokenManagedListFetch,
	}, UserListFilters{
		Role:                 RoleUser,
		Search:               managedTokenEmailDomain,
		IncludeSubscriptions: &includeSubscriptions,
	})
	if err != nil {
		return nil, 0, err
	}

	needle := strings.ToLower(strings.TrimSpace(filters.Search))
	records := make([]ManagedToken, 0, len(users))
	for i := range users {
		if !IsManagedTokenUser(&users[i]) {
			continue
		}

		userCopy := users[i]
		apiKeys, _, err := s.apiKeyService.List(ctx, userCopy.ID, pagination.PaginationParams{Page: 1, PageSize: 50}, APIKeyListFilters{})
		if err != nil {
			return nil, 0, err
		}
		managedKey := findManagedTokenAPIKey(apiKeys)
		if managedKey == nil {
			continue
		}

		subscriptions, err := s.subscriptionService.ListUserSubscriptions(ctx, userCopy.ID)
		if err != nil {
			return nil, 0, err
		}
		managedSubscription := findManagedTokenSubscription(subscriptions, managedKey.GroupID)

		record := ManagedToken{
			Label:        managedTokenLabel(managedKey.Name),
			User:         &userCopy,
			APIKey:       managedKey,
			Subscription: managedSubscription,
		}
		if !matchesManagedTokenSearch(record, needle) {
			continue
		}
		records = append(records, record)
	}

	sort.Slice(records, func(i, j int) bool {
		leftTime := records[i].User.CreatedAt
		rightTime := records[j].User.CreatedAt
		if records[i].APIKey != nil {
			leftTime = records[i].APIKey.CreatedAt
		}
		if records[j].APIKey != nil {
			rightTime = records[j].APIKey.CreatedAt
		}
		return leftTime.After(rightTime)
	})

	total := int64(len(records))
	start := (page - 1) * pageSize
	if start >= len(records) {
		return []ManagedToken{}, total, nil
	}
	end := start + pageSize
	if end > len(records) {
		end = len(records)
	}

	return records[start:end], total, nil
}

func (s *TokenManagementService) Create(ctx context.Context, input *CreateManagedTokenInput) (*ManagedToken, error) {
	if input == nil {
		return nil, infraerrors.BadRequest("TOKEN_INPUT_REQUIRED", "token input is required")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, infraerrors.BadRequest("TOKEN_NAME_REQUIRED", "token name is required")
	}
	if input.GroupID <= 0 {
		return nil, infraerrors.BadRequest("TOKEN_GROUP_REQUIRED", "group_id is required")
	}

	keyValue, err := s.resolveTokenValue(input.CustomKey)
	if err != nil {
		return nil, err
	}

	username := managedTokenUsername(name)
	email := username
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("check managed token email exists: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	user := &User{
		Email:       email,
		Username:    username,
		Notes:       buildManagedTokenNotes(name, input.Notes),
		Role:        RoleUser,
		Balance:     s.defaultUserBalance(ctx),
		Concurrency: s.defaultUserConcurrency(ctx),
		Status:      StatusActive,
	}
	if err := user.SetPassword(keyValue); err != nil {
		return nil, err
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	var createdSubscriptionID int64
	cleanupUser := func(stage string, cause error) {
		if createdSubscriptionID > 0 {
			if revokeErr := s.subscriptionService.RevokeSubscription(ctx, createdSubscriptionID); revokeErr != nil {
				slog.Error("token management cleanup failed", "stage", stage, "cause", cause, "cleanup_error", revokeErr, "subscription_id", createdSubscriptionID)
			}
		}
		if user.ID <= 0 {
			return
		}
		if deleteErr := s.userRepo.Delete(ctx, user.ID); deleteErr != nil {
			slog.Error("token management cleanup failed", "stage", stage, "cause", cause, "cleanup_error", deleteErr, "user_id", user.ID)
		}
	}

	subscription, err := s.subscriptionService.AssignSubscription(ctx, &AssignSubscriptionInput{
		UserID:       user.ID,
		GroupID:      input.GroupID,
		ValidityDays: input.ValidityDays,
		AssignedBy:   input.AssignedBy,
		Notes:        buildManagedTokenNotes(name, input.Notes),
	})
	if err != nil {
		cleanupUser("assign_subscription", err)
		return nil, err
	}
	if subscription != nil {
		createdSubscriptionID = subscription.ID
	}

	groupID := input.GroupID
	managedKeyName := managedTokenKeyNamePrefix + name
	apiKey, err := s.apiKeyService.Create(ctx, user.ID, CreateAPIKeyRequest{
		Name:      managedKeyName,
		GroupID:   &groupID,
		CustomKey: &keyValue,
	})
	if err != nil {
		cleanupUser("create_api_key", err)
		return nil, err
	}

	return &ManagedToken{
		Label:        name,
		User:         user,
		APIKey:       apiKey,
		Subscription: subscription,
	}, nil
}

func (s *TokenManagementService) Delete(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return infraerrors.BadRequest("TOKEN_ID_REQUIRED", "token id is required")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if !IsManagedTokenUser(user) {
		return infraerrors.NotFound("TOKEN_NOT_FOUND", "managed token not found")
	}

	if s.subscriptionService != nil {
		subscriptions, err := s.subscriptionService.ListUserSubscriptions(ctx, userID)
		if err != nil {
			return err
		}
		for i := range subscriptions {
			if err := s.subscriptionService.RevokeSubscription(ctx, subscriptions[i].ID); err != nil {
				return err
			}
		}
	}

	if s.apiKeyService != nil {
		s.apiKeyService.InvalidateAuthCacheByUserID(ctx, userID)
	}

	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return err
	}
	return nil
}

func (s *TokenManagementService) resolveTokenValue(customKey *string) (string, error) {
	if customKey != nil {
		trimmed := strings.TrimSpace(*customKey)
		if trimmed != "" {
			return trimmed, nil
		}
	}
	return s.apiKeyService.GenerateKey()
}

func (s *TokenManagementService) defaultUserBalance(ctx context.Context) float64 {
	if s.settingService != nil {
		return s.settingService.GetDefaultBalance(ctx)
	}
	if s.cfg != nil {
		return s.cfg.Default.UserBalance
	}
	return 0
}

func (s *TokenManagementService) defaultUserConcurrency(ctx context.Context) int {
	if s.settingService != nil {
		if concurrency := s.settingService.GetDefaultConcurrency(ctx); concurrency > 0 {
			return concurrency
		}
	}
	if s.cfg != nil && s.cfg.Default.UserConcurrency > 0 {
		return s.cfg.Default.UserConcurrency
	}
	return 1
}

func buildManagedTokenNotes(name, notes string) string {
	parts := []string{managedTokenNoteMarker, "name=" + strings.TrimSpace(name)}
	if trimmedNotes := strings.TrimSpace(notes); trimmedNotes != "" {
		parts = append(parts, trimmedNotes)
	}
	return strings.Join(parts, "\n")
}

func managedTokenUsername(name string) string {
	trimmed := strings.TrimSpace(name)
	suffix := "@" + managedTokenEmailDomain
	maxLocalPartLen := managedTokenUsernameMaxLen - len([]rune(suffix))
	if maxLocalPartLen > 0 {
		runes := []rune(trimmed)
		if len(runes) > maxLocalPartLen {
			trimmed = string(runes[:maxLocalPartLen])
		}
	}
	return trimmed + suffix
}

func managedTokenEmail(name string) string {
	return managedTokenUsername(name)
}

func IsManagedTokenUser(user *User) bool {
	if user == nil {
		return false
	}
	if !strings.HasSuffix(strings.ToLower(user.Email), "@"+managedTokenEmailDomain) {
		return false
	}
	return strings.Contains(user.Notes, managedTokenNoteMarker)
}

func findManagedTokenAPIKey(keys []APIKey) *APIKey {
	if len(keys) == 0 {
		return nil
	}

	var managedKeys []APIKey
	for i := range keys {
		if strings.HasPrefix(keys[i].Name, managedTokenKeyNamePrefix) {
			managedKeys = append(managedKeys, keys[i])
		}
	}
	if len(managedKeys) == 0 {
		return nil
	}

	sort.Slice(managedKeys, func(i, j int) bool {
		return managedKeys[i].CreatedAt.After(managedKeys[j].CreatedAt)
	})

	key := managedKeys[0]
	return &key
}

func findManagedTokenSubscription(subscriptions []UserSubscription, groupID *int64) *UserSubscription {
	if len(subscriptions) == 0 {
		return nil
	}
	if groupID != nil {
		for i := range subscriptions {
			if subscriptions[i].GroupID == *groupID {
				sub := subscriptions[i]
				return &sub
			}
		}
	}

	subscription := subscriptions[0]
	return &subscription
}

func managedTokenLabel(keyName string) string {
	if strings.HasPrefix(keyName, managedTokenKeyNamePrefix) {
		return strings.TrimSpace(strings.TrimPrefix(keyName, managedTokenKeyNamePrefix))
	}
	return keyName
}

func matchesManagedTokenSearch(record ManagedToken, needle string) bool {
	if needle == "" {
		return true
	}

	fields := []string{record.Label}
	if record.User != nil {
		fields = append(fields, record.User.Email, record.User.Username)
	}
	if record.APIKey != nil {
		fields = append(fields, record.APIKey.Key, record.APIKey.Name)
	}
	if record.Subscription != nil && record.Subscription.Group != nil {
		fields = append(fields, record.Subscription.Group.Name)
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), needle) {
			return true
		}
	}
	return false
}
