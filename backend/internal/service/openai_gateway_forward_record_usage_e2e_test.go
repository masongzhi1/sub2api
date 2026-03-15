package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpenAIGatewayService_ForwardThenRecordUsage_PreservesClientModelOnOAuthSSE(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "curl/8.0")
	SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)

	upstreamSSE := strings.Join([]string{
		`event: response.completed`,
		`data: {"type":"response.completed","response":{"id":"resp_test","model":"gpt-5.2-codex","usage":{"input_tokens":21,"output_tokens":18,"input_tokens_details":{"cached_tokens":0}}}}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")

	upstream := &httpUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"text/event-stream"},
				"x-request-id":  []string{"rid-sse-e2e"},
			},
			Body: io.NopCloser(strings.NewReader(upstreamSSE)),
		},
	}

	cfg := &config.Config{}
	cfg.Security.URLAllowlist.Enabled = false
	cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	cfg.Gateway.OpenAIWS.Enabled = false

	usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
	userRepo := &openAIRecordUsageUserRepoStub{}
	subRepo := &openAIRecordUsageSubRepoStub{}

	svc := &OpenAIGatewayService{
		cfg:                 cfg,
		httpUpstream:        upstream,
		openaiWSResolver:    NewOpenAIWSProtocolResolver(cfg),
		usageLogRepo:        usageRepo,
		userRepo:            userRepo,
		userSubRepo:         subRepo,
		billingService:      NewBillingService(cfg, nil),
		billingCacheService: &BillingCacheService{},
		deferredService:     &DeferredService{},
	}

	account := &Account{
		ID:             54,
		Name:           "oauth-normal",
		Platform:       PlatformOpenAI,
		Type:           AccountTypeOAuth,
		Concurrency:    1,
		Credentials:    map[string]any{"access_token": "oauth-token", "chatgpt_account_id": "chatgpt-acc"},
		Status:         StatusActive,
		Schedulable:    true,
		RateMultiplier: f64p(1),
	}
	apiKey := &APIKey{ID: 1001, User: &User{ID: 2001}}

	body := []byte(`{"model":"gpt-5.4","input":"reply with OK only","stream":false}`)
	result, err := svc.Forward(context.Background(), c, account, body)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "gpt-5.4", result.Model)
	require.Equal(t, 21, result.Usage.InputTokens)
	require.Equal(t, 18, result.Usage.OutputTokens)

	err = svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
		Result:       result,
		APIKey:       apiKey,
		User:         apiKey.User,
		Account:      account,
		Subscription: nil,
		UserAgent:    "curl/8.0",
		IPAddress:    "127.0.0.1",
	})
	require.NoError(t, err)
	require.NotNil(t, usageRepo.lastLog)
	require.Equal(t, "gpt-5.4", usageRepo.lastLog.Model)
	require.WithinDuration(t, time.Now(), usageRepo.lastLog.CreatedAt, 5*time.Second)
}
