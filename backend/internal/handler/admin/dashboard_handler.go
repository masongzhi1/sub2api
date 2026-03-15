package admin

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shirou/gopsutil/v4/cpu"
	gopsutilmem "github.com/shirou/gopsutil/v4/mem"
	gopsutilnet "github.com/shirou/gopsutil/v4/net"

	"github.com/gin-gonic/gin"
)

// DashboardHandler handles admin dashboard statistics
type DashboardHandler struct {
	dashboardService   *service.DashboardService
	aggregationService *service.DashboardAggregationService
	startTime          time.Time // Server start time for uptime calculation
}

// NewDashboardHandler creates a new admin dashboard handler
func NewDashboardHandler(dashboardService *service.DashboardService, aggregationService *service.DashboardAggregationService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService:   dashboardService,
		aggregationService: aggregationService,
		startTime:          time.Now(),
	}
}

// parseTimeRange parses start_date, end_date query parameters
// Uses user's timezone if provided, otherwise falls back to server timezone
func parseTimeRange(c *gin.Context) (time.Time, time.Time) {
	userTZ := c.Query("timezone") // Get user's timezone from request
	now := timezone.NowInUserLocation(userTZ)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var startTime, endTime time.Time

	if startDate != "" {
		if t, err := timezone.ParseInUserLocation("2006-01-02", startDate, userTZ); err == nil {
			startTime = t
		} else {
			startTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -7), userTZ)
		}
	} else {
		startTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -7), userTZ)
	}

	if endDate != "" {
		if t, err := timezone.ParseInUserLocation("2006-01-02", endDate, userTZ); err == nil {
			endTime = t.Add(24 * time.Hour) // Include the end date
		} else {
			endTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)
		}
	} else {
		endTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)
	}

	return startTime, endTime
}

// GetStats handles getting dashboard statistics
// GET /api/v1/admin/dashboard/stats
func (h *DashboardHandler) GetStats(c *gin.Context) {
	stats, err := h.dashboardService.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.Error(c, 500, "Failed to get dashboard statistics")
		return
	}

	// Calculate uptime in seconds
	uptime := int64(time.Since(h.startTime).Seconds())

	response.Success(c, gin.H{
		// 用户统计
		"total_users":     stats.TotalUsers,
		"today_new_users": stats.TodayNewUsers,
		"active_users":    stats.ActiveUsers,

		// API Key 统计
		"total_api_keys":  stats.TotalAPIKeys,
		"active_api_keys": stats.ActiveAPIKeys,

		// 账户统计
		"total_accounts":     stats.TotalAccounts,
		"normal_accounts":    stats.NormalAccounts,
		"error_accounts":     stats.ErrorAccounts,
		"ratelimit_accounts": stats.RateLimitAccounts,
		"overload_accounts":  stats.OverloadAccounts,

		// 累计 Token 使用统计
		"total_requests":              stats.TotalRequests,
		"total_input_tokens":          stats.TotalInputTokens,
		"total_output_tokens":         stats.TotalOutputTokens,
		"total_cache_creation_tokens": stats.TotalCacheCreationTokens,
		"total_cache_read_tokens":     stats.TotalCacheReadTokens,
		"total_tokens":                stats.TotalTokens,
		"total_cost":                  stats.TotalCost,       // 标准计费
		"total_actual_cost":           stats.TotalActualCost, // 实际扣除

		// 今日 Token 使用统计
		"today_requests":              stats.TodayRequests,
		"today_input_tokens":          stats.TodayInputTokens,
		"today_output_tokens":         stats.TodayOutputTokens,
		"today_cache_creation_tokens": stats.TodayCacheCreationTokens,
		"today_cache_read_tokens":     stats.TodayCacheReadTokens,
		"today_tokens":                stats.TodayTokens,
		"today_cost":                  stats.TodayCost,       // 今日标准计费
		"today_actual_cost":           stats.TodayActualCost, // 今日实际扣除

		// 系统运行统计
		"average_duration_ms": stats.AverageDurationMs,
		"uptime":              uptime,

		// 性能指标
		"rpm": stats.Rpm,
		"tpm": stats.Tpm,

		// 预聚合新鲜度
		"hourly_active_users": stats.HourlyActiveUsers,
		"stats_updated_at":    stats.StatsUpdatedAt,
		"stats_stale":         stats.StatsStale,
	})
}

type DashboardAggregationBackfillRequest struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// BackfillAggregation handles triggering aggregation backfill
// POST /api/v1/admin/dashboard/aggregation/backfill
func (h *DashboardHandler) BackfillAggregation(c *gin.Context) {
	if h.aggregationService == nil {
		response.InternalError(c, "Aggregation service not available")
		return
	}

	var req DashboardAggregationBackfillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	start, err := time.Parse(time.RFC3339, req.Start)
	if err != nil {
		response.BadRequest(c, "Invalid start time")
		return
	}
	end, err := time.Parse(time.RFC3339, req.End)
	if err != nil {
		response.BadRequest(c, "Invalid end time")
		return
	}

	if err := h.aggregationService.TriggerBackfill(start, end); err != nil {
		if errors.Is(err, service.ErrDashboardBackfillDisabled) {
			response.Forbidden(c, "Backfill is disabled")
			return
		}
		if errors.Is(err, service.ErrDashboardBackfillTooLarge) {
			response.BadRequest(c, "Backfill range too large")
			return
		}
		response.InternalError(c, "Failed to trigger backfill")
		return
	}

	response.Success(c, gin.H{
		"status": "accepted",
	})
}

// GetRealtimeMetrics handles getting real-time system metrics
// GET /api/v1/admin/dashboard/realtime
func (h *DashboardHandler) GetRealtimeMetrics(c *gin.Context) {
	// Return mock data for now
	response.Success(c, gin.H{
		"active_requests":       0,
		"requests_per_minute":   0,
		"average_response_time": 0,
		"error_rate":            0.0,
	})
}

type dashboardRuntimeNodeMetric struct {
	Node              string  `json:"node"`
	NodeName          string  `json:"node_name"`
	Timestamp         int64   `json:"timestamp"`
	ActiveConnections int64   `json:"active_connections"`
	ActiveAPIKeys     int64   `json:"active_api_keys"`
	CPUPercent        float64 `json:"cpu_percent"`
	MemoryPercent     float64 `json:"memory_percent"`
	MemoryUsedBytes   uint64  `json:"memory_used_bytes"`
	MemoryTotalBytes  uint64  `json:"memory_total_bytes"`
	NetworkRxBytes    uint64  `json:"network_rx_bytes"`
	NetworkTxBytes    uint64  `json:"network_tx_bytes"`
	OK                bool    `json:"ok"`
	Error             string  `json:"error,omitempty"`
}

type dashboardRuntimeResponseEnvelope struct {
	Code    int                        `json:"code"`
	Message string                     `json:"message"`
	Data    dashboardRuntimeNodeMetric `json:"data"`
}

// GetRuntimeLocal returns local node runtime metrics for admin dashboard cluster aggregation.
// GET /api/v1/admin/dashboard/runtime-local
func (h *DashboardHandler) GetRuntimeLocal(c *gin.Context) {
	ctx := c.Request.Context()

	hostname := os.Getenv("HOSTNAME")
	if strings.TrimSpace(hostname) == "" {
		if hn, err := os.Hostname(); err == nil {
			hostname = hn
		}
	}

	var (
		cpuPercent       float64
		memoryPercent    float64
		memoryUsedBytes  uint64
		memoryTotalBytes uint64
		networkRxBytes   uint64
		networkTxBytes   uint64
		activeConns      int64
		activeAPIKeys    int64
	)

	if percents, err := cpu.PercentWithContext(ctx, 0, false); err == nil && len(percents) > 0 {
		cpuPercent = roundTo2DP(percents[0])
	}

	if vm, err := gopsutilmem.VirtualMemoryWithContext(ctx); err == nil && vm != nil {
		memoryPercent = roundTo2DP(vm.UsedPercent)
		memoryUsedBytes = vm.Used
		memoryTotalBytes = vm.Total
	}

	if ioStats, err := gopsutilnet.IOCountersWithContext(ctx, false); err == nil && len(ioStats) > 0 {
		networkRxBytes = ioStats[0].BytesRecv
		networkTxBytes = ioStats[0].BytesSent
	}

	// Count established inbound TCP connections to app port 8080.
	if conns, err := gopsutilnet.ConnectionsWithContext(ctx, "tcp"); err == nil {
		for i := range conns {
			conn := conns[i]
			if conn.Laddr.Port == 8080 && strings.EqualFold(strings.TrimSpace(conn.Status), "ESTABLISHED") {
				activeConns++
			}
		}
	}

	if stats, err := h.dashboardService.GetDashboardStats(ctx); err == nil {
		activeAPIKeys = stats.ActiveAPIKeys
	}

	response.Success(c, dashboardRuntimeNodeMetric{
		NodeName:          strings.TrimSpace(hostname),
		Timestamp:         time.Now().Unix(),
		ActiveConnections: activeConns,
		ActiveAPIKeys:     activeAPIKeys,
		CPUPercent:        cpuPercent,
		MemoryPercent:     memoryPercent,
		MemoryUsedBytes:   memoryUsedBytes,
		MemoryTotalBytes:  memoryTotalBytes,
		NetworkRxBytes:    networkRxBytes,
		NetworkTxBytes:    networkTxBytes,
		OK:                true,
	})
}

// GetRuntimeCluster aggregates runtime metrics from cluster nodes for dashboard charts.
// GET /api/v1/admin/dashboard/runtime-cluster
func (h *DashboardHandler) GetRuntimeCluster(c *gin.Context) {
	ctx := c.Request.Context()
	nodes := resolveDashboardClusterNodes()
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))

	client := &http.Client{
		Timeout: 4 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // internal admin query
		},
	}

	out := make([]dashboardRuntimeNodeMetric, 0, len(nodes))
	var latestActiveAPIKeys int64
	for i := range nodes {
		node := nodes[i]
		metric, err := fetchDashboardRuntimeNode(ctx, client, node, authHeader)
		if err != nil {
			out = append(out, dashboardRuntimeNodeMetric{
				Node:      node,
				Timestamp: time.Now().Unix(),
				OK:        false,
				Error:     err.Error(),
			})
			continue
		}
		if strings.TrimSpace(metric.Node) == "" {
			metric.Node = node
		}
		metric.OK = true
		if metric.ActiveAPIKeys > 0 {
			latestActiveAPIKeys = metric.ActiveAPIKeys
		}
		out = append(out, metric)
	}

	if latestActiveAPIKeys == 0 {
		if stats, err := h.dashboardService.GetDashboardStats(ctx); err == nil {
			latestActiveAPIKeys = stats.ActiveAPIKeys
		}
	}

	response.Success(c, gin.H{
		"timestamp":             time.Now().Unix(),
		"active_api_keys":       latestActiveAPIKeys,
		"refresh_interval_secs": 5,
		"nodes":                 out,
	})
}

func resolveDashboardClusterNodes() []string {
	raw := strings.TrimSpace(os.Getenv("DASHBOARD_CLUSTER_NODES"))
	if raw == "" {
		return []string{"38.175.200.213", "38.175.200.245", "38.175.200.178"}
	}

	parts := strings.Split(raw, ",")
	seen := make(map[string]struct{}, len(parts))
	nodes := make([]string, 0, len(parts))
	for i := range parts {
		node := strings.TrimSpace(parts[i])
		if node == "" {
			continue
		}
		if _, ok := seen[node]; ok {
			continue
		}
		seen[node] = struct{}{}
		nodes = append(nodes, node)
	}
	if len(nodes) == 0 {
		return []string{"38.175.200.213", "38.175.200.245", "38.175.200.178"}
	}
	return nodes
}

func fetchDashboardRuntimeNode(
	ctx context.Context,
	client *http.Client,
	node string,
	authHeader string,
) (dashboardRuntimeNodeMetric, error) {
	url := fmt.Sprintf("https://%s/api/v1/admin/dashboard/runtime-local", strings.TrimSpace(node))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return dashboardRuntimeNodeMetric{}, err
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := client.Do(req)
	if err != nil {
		return dashboardRuntimeNodeMetric{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return dashboardRuntimeNodeMetric{}, fmt.Errorf("http %d", resp.StatusCode)
	}

	var envelope dashboardRuntimeResponseEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return dashboardRuntimeNodeMetric{}, err
	}
	if envelope.Code != 0 {
		msg := strings.TrimSpace(envelope.Message)
		if msg == "" {
			msg = "invalid response"
		}
		return dashboardRuntimeNodeMetric{}, errors.New(msg)
	}
	return envelope.Data, nil
}

func roundTo2DP(v float64) float64 {
	if v == 0 {
		return 0
	}
	return float64(int(v*100+0.5)) / 100
}

// GetUsageTrend handles getting usage trend data
// GET /api/v1/admin/dashboard/trend
// Query params: start_date, end_date (YYYY-MM-DD), granularity (day/hour), user_id, api_key_id, model, account_id, group_id, request_type, stream, billing_type
func (h *DashboardHandler) GetUsageTrend(c *gin.Context) {
	startTime, endTime := parseTimeRange(c)
	granularity := c.DefaultQuery("granularity", "day")

	// Parse optional filter params
	var userID, apiKeyID, accountID, groupID int64
	var model string
	var requestType *int16
	var stream *bool
	var billingType *int8

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			userID = id
		}
	}
	if apiKeyIDStr := c.Query("api_key_id"); apiKeyIDStr != "" {
		if id, err := strconv.ParseInt(apiKeyIDStr, 10, 64); err == nil {
			apiKeyID = id
		}
	}
	if accountIDStr := c.Query("account_id"); accountIDStr != "" {
		if id, err := strconv.ParseInt(accountIDStr, 10, 64); err == nil {
			accountID = id
		}
	}
	if groupIDStr := c.Query("group_id"); groupIDStr != "" {
		if id, err := strconv.ParseInt(groupIDStr, 10, 64); err == nil {
			groupID = id
		}
	}
	if modelStr := c.Query("model"); modelStr != "" {
		model = modelStr
	}
	if requestTypeStr := strings.TrimSpace(c.Query("request_type")); requestTypeStr != "" {
		parsed, err := service.ParseUsageRequestType(requestTypeStr)
		if err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		value := int16(parsed)
		requestType = &value
	} else if streamStr := c.Query("stream"); streamStr != "" {
		if streamVal, err := strconv.ParseBool(streamStr); err == nil {
			stream = &streamVal
		} else {
			response.BadRequest(c, "Invalid stream value, use true or false")
			return
		}
	}
	if billingTypeStr := c.Query("billing_type"); billingTypeStr != "" {
		if v, err := strconv.ParseInt(billingTypeStr, 10, 8); err == nil {
			bt := int8(v)
			billingType = &bt
		} else {
			response.BadRequest(c, "Invalid billing_type")
			return
		}
	}

	trend, err := h.dashboardService.GetUsageTrendWithFilters(c.Request.Context(), startTime, endTime, granularity, userID, apiKeyID, accountID, groupID, model, requestType, stream, billingType)
	if err != nil {
		response.Error(c, 500, "Failed to get usage trend")
		return
	}

	response.Success(c, gin.H{
		"trend":       trend,
		"start_date":  startTime.Format("2006-01-02"),
		"end_date":    endTime.Add(-24 * time.Hour).Format("2006-01-02"),
		"granularity": granularity,
	})
}

// GetModelStats handles getting model usage statistics
// GET /api/v1/admin/dashboard/models
// Query params: start_date, end_date (YYYY-MM-DD), user_id, api_key_id, account_id, group_id, request_type, stream, billing_type
func (h *DashboardHandler) GetModelStats(c *gin.Context) {
	startTime, endTime := parseTimeRange(c)

	// Parse optional filter params
	var userID, apiKeyID, accountID, groupID int64
	var requestType *int16
	var stream *bool
	var billingType *int8

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			userID = id
		}
	}
	if apiKeyIDStr := c.Query("api_key_id"); apiKeyIDStr != "" {
		if id, err := strconv.ParseInt(apiKeyIDStr, 10, 64); err == nil {
			apiKeyID = id
		}
	}
	if accountIDStr := c.Query("account_id"); accountIDStr != "" {
		if id, err := strconv.ParseInt(accountIDStr, 10, 64); err == nil {
			accountID = id
		}
	}
	if groupIDStr := c.Query("group_id"); groupIDStr != "" {
		if id, err := strconv.ParseInt(groupIDStr, 10, 64); err == nil {
			groupID = id
		}
	}
	if requestTypeStr := strings.TrimSpace(c.Query("request_type")); requestTypeStr != "" {
		parsed, err := service.ParseUsageRequestType(requestTypeStr)
		if err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		value := int16(parsed)
		requestType = &value
	} else if streamStr := c.Query("stream"); streamStr != "" {
		if streamVal, err := strconv.ParseBool(streamStr); err == nil {
			stream = &streamVal
		} else {
			response.BadRequest(c, "Invalid stream value, use true or false")
			return
		}
	}
	if billingTypeStr := c.Query("billing_type"); billingTypeStr != "" {
		if v, err := strconv.ParseInt(billingTypeStr, 10, 8); err == nil {
			bt := int8(v)
			billingType = &bt
		} else {
			response.BadRequest(c, "Invalid billing_type")
			return
		}
	}

	stats, err := h.dashboardService.GetModelStatsWithFilters(c.Request.Context(), startTime, endTime, userID, apiKeyID, accountID, groupID, requestType, stream, billingType)
	if err != nil {
		response.Error(c, 500, "Failed to get model statistics")
		return
	}

	response.Success(c, gin.H{
		"models":     stats,
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   endTime.Add(-24 * time.Hour).Format("2006-01-02"),
	})
}

// GetGroupStats handles getting group usage statistics
// GET /api/v1/admin/dashboard/groups
// Query params: start_date, end_date (YYYY-MM-DD), user_id, api_key_id, account_id, group_id, request_type, stream, billing_type
func (h *DashboardHandler) GetGroupStats(c *gin.Context) {
	startTime, endTime := parseTimeRange(c)

	var userID, apiKeyID, accountID, groupID int64
	var requestType *int16
	var stream *bool
	var billingType *int8

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			userID = id
		}
	}
	if apiKeyIDStr := c.Query("api_key_id"); apiKeyIDStr != "" {
		if id, err := strconv.ParseInt(apiKeyIDStr, 10, 64); err == nil {
			apiKeyID = id
		}
	}
	if accountIDStr := c.Query("account_id"); accountIDStr != "" {
		if id, err := strconv.ParseInt(accountIDStr, 10, 64); err == nil {
			accountID = id
		}
	}
	if groupIDStr := c.Query("group_id"); groupIDStr != "" {
		if id, err := strconv.ParseInt(groupIDStr, 10, 64); err == nil {
			groupID = id
		}
	}
	if requestTypeStr := strings.TrimSpace(c.Query("request_type")); requestTypeStr != "" {
		parsed, err := service.ParseUsageRequestType(requestTypeStr)
		if err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		value := int16(parsed)
		requestType = &value
	} else if streamStr := c.Query("stream"); streamStr != "" {
		if streamVal, err := strconv.ParseBool(streamStr); err == nil {
			stream = &streamVal
		} else {
			response.BadRequest(c, "Invalid stream value, use true or false")
			return
		}
	}
	if billingTypeStr := c.Query("billing_type"); billingTypeStr != "" {
		if v, err := strconv.ParseInt(billingTypeStr, 10, 8); err == nil {
			bt := int8(v)
			billingType = &bt
		} else {
			response.BadRequest(c, "Invalid billing_type")
			return
		}
	}

	stats, err := h.dashboardService.GetGroupStatsWithFilters(c.Request.Context(), startTime, endTime, userID, apiKeyID, accountID, groupID, requestType, stream, billingType)
	if err != nil {
		response.Error(c, 500, "Failed to get group statistics")
		return
	}

	response.Success(c, gin.H{
		"groups":     stats,
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   endTime.Add(-24 * time.Hour).Format("2006-01-02"),
	})
}

// GetAPIKeyUsageTrend handles getting API key usage trend data
// GET /api/v1/admin/dashboard/api-keys-trend
// Query params: start_date, end_date (YYYY-MM-DD), granularity (day/hour), limit (default 5)
func (h *DashboardHandler) GetAPIKeyUsageTrend(c *gin.Context) {
	startTime, endTime := parseTimeRange(c)
	granularity := c.DefaultQuery("granularity", "day")
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}

	trend, err := h.dashboardService.GetAPIKeyUsageTrend(c.Request.Context(), startTime, endTime, granularity, limit)
	if err != nil {
		response.Error(c, 500, "Failed to get API key usage trend")
		return
	}

	response.Success(c, gin.H{
		"trend":       trend,
		"start_date":  startTime.Format("2006-01-02"),
		"end_date":    endTime.Add(-24 * time.Hour).Format("2006-01-02"),
		"granularity": granularity,
	})
}

// GetUserUsageTrend handles getting user usage trend data
// GET /api/v1/admin/dashboard/users-trend
// Query params: start_date, end_date (YYYY-MM-DD), granularity (day/hour), limit (default 12)
func (h *DashboardHandler) GetUserUsageTrend(c *gin.Context) {
	startTime, endTime := parseTimeRange(c)
	granularity := c.DefaultQuery("granularity", "day")
	limitStr := c.DefaultQuery("limit", "12")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 12
	}

	trend, err := h.dashboardService.GetUserUsageTrend(c.Request.Context(), startTime, endTime, granularity, limit)
	if err != nil {
		response.Error(c, 500, "Failed to get user usage trend")
		return
	}

	response.Success(c, gin.H{
		"trend":       trend,
		"start_date":  startTime.Format("2006-01-02"),
		"end_date":    endTime.Add(-24 * time.Hour).Format("2006-01-02"),
		"granularity": granularity,
	})
}

// BatchUsersUsageRequest represents the request body for batch user usage stats
type BatchUsersUsageRequest struct {
	UserIDs []int64 `json:"user_ids" binding:"required"`
}

var dashboardBatchUsersUsageCache = newSnapshotCache(30 * time.Second)
var dashboardBatchAPIKeysUsageCache = newSnapshotCache(30 * time.Second)

// GetBatchUsersUsage handles getting usage stats for multiple users
// POST /api/v1/admin/dashboard/users-usage
func (h *DashboardHandler) GetBatchUsersUsage(c *gin.Context) {
	var req BatchUsersUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	userIDs := normalizeInt64IDList(req.UserIDs)
	if len(userIDs) == 0 {
		response.Success(c, gin.H{"stats": map[string]any{}})
		return
	}

	keyRaw, _ := json.Marshal(struct {
		UserIDs []int64 `json:"user_ids"`
	}{
		UserIDs: userIDs,
	})
	cacheKey := string(keyRaw)
	if cached, ok := dashboardBatchUsersUsageCache.Get(cacheKey); ok {
		c.Header("X-Snapshot-Cache", "hit")
		response.Success(c, cached.Payload)
		return
	}

	stats, err := h.dashboardService.GetBatchUserUsageStats(c.Request.Context(), userIDs, time.Time{}, time.Time{})
	if err != nil {
		response.Error(c, 500, "Failed to get user usage stats")
		return
	}

	payload := gin.H{"stats": stats}
	dashboardBatchUsersUsageCache.Set(cacheKey, payload)
	c.Header("X-Snapshot-Cache", "miss")
	response.Success(c, payload)
}

// BatchAPIKeysUsageRequest represents the request body for batch api key usage stats
type BatchAPIKeysUsageRequest struct {
	APIKeyIDs []int64 `json:"api_key_ids" binding:"required"`
}

// GetBatchAPIKeysUsage handles getting usage stats for multiple API keys
// POST /api/v1/admin/dashboard/api-keys-usage
func (h *DashboardHandler) GetBatchAPIKeysUsage(c *gin.Context) {
	var req BatchAPIKeysUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	apiKeyIDs := normalizeInt64IDList(req.APIKeyIDs)
	if len(apiKeyIDs) == 0 {
		response.Success(c, gin.H{"stats": map[string]any{}})
		return
	}

	keyRaw, _ := json.Marshal(struct {
		APIKeyIDs []int64 `json:"api_key_ids"`
	}{
		APIKeyIDs: apiKeyIDs,
	})
	cacheKey := string(keyRaw)
	if cached, ok := dashboardBatchAPIKeysUsageCache.Get(cacheKey); ok {
		c.Header("X-Snapshot-Cache", "hit")
		response.Success(c, cached.Payload)
		return
	}

	stats, err := h.dashboardService.GetBatchAPIKeyUsageStats(c.Request.Context(), apiKeyIDs, time.Time{}, time.Time{})
	if err != nil {
		response.Error(c, 500, "Failed to get API key usage stats")
		return
	}

	payload := gin.H{"stats": stats}
	dashboardBatchAPIKeysUsageCache.Set(cacheKey, payload)
	c.Header("X-Snapshot-Cache", "miss")
	response.Success(c, payload)
}
