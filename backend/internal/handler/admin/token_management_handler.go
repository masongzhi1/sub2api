package admin

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type TokenManagementHandler struct {
	tokenManagementService *service.TokenManagementService
}

type CreateManagedTokenRequest struct {
	Name         string  `json:"name" binding:"required"`
	GroupID      int64   `json:"group_id" binding:"required"`
	ValidityDays int     `json:"validity_days" binding:"omitempty,min=1,max=36500"`
	CustomKey    *string `json:"custom_key"`
	Notes        string  `json:"notes"`
}

type ManagedTokenResponse struct {
	Label        string                     `json:"label"`
	User         *dto.AdminUser             `json:"user"`
	APIKey       *dto.APIKey                `json:"api_key"`
	Subscription *dto.AdminUserSubscription `json:"subscription,omitempty"`
}

func NewTokenManagementHandler(tokenManagementService *service.TokenManagementService) *TokenManagementHandler {
	return &TokenManagementHandler{tokenManagementService: tokenManagementService}
}

func (h *TokenManagementHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)

	items, total, err := h.tokenManagementService.List(c.Request.Context(), page, pageSize, service.ListManagedTokensFilters{
		Search: strings.TrimSpace(c.Query("search")),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]ManagedTokenResponse, 0, len(items))
	for i := range items {
		out = append(out, managedTokenResponse(&items[i]))
	}
	response.Paginated(c, out, total, page, pageSize)
}

func (h *TokenManagementHandler) Create(c *gin.Context) {
	var req CreateManagedTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	item, err := h.tokenManagementService.Create(c.Request.Context(), &service.CreateManagedTokenInput{
		Name:         req.Name,
		GroupID:      req.GroupID,
		ValidityDays: req.ValidityDays,
		CustomKey:    req.CustomKey,
		Notes:        req.Notes,
		AssignedBy:   getAdminIDFromContext(c),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, managedTokenResponse(item))
}

func managedTokenResponse(item *service.ManagedToken) ManagedTokenResponse {
	if item == nil {
		return ManagedTokenResponse{}
	}
	return ManagedTokenResponse{
		Label:        item.Label,
		User:         dto.UserFromServiceAdmin(item.User),
		APIKey:       dto.APIKeyFromService(item.APIKey),
		Subscription: dto.UserSubscriptionFromServiceAdmin(item.Subscription),
	}
}
