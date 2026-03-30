package management

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
)

// ── GET /v0/management/dynamic-proxy ─────────────────────────────────────────

// GetDynamicProxy 返回当前动态代理配置。
func (h *Handler) GetDynamicProxy(c *gin.Context) {
	if h == nil || h.cfg == nil {
		c.JSON(http.StatusOK, config.DynamicProxyConfig{})
		return
	}
	c.JSON(http.StatusOK, h.cfg.DynamicProxy)
}

// ── PUT /v0/management/dynamic-proxy ─────────────────────────────────────────

// PutDynamicProxy 保存动态代理配置并持久化到 config.yaml。
func (h *Handler) PutDynamicProxy(c *gin.Context) {
	var body config.DynamicProxyConfig
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body", "message": err.Error()})
		return
	}
	body.APIURL = strings.TrimSpace(body.APIURL)
	body.APIKey = strings.TrimSpace(body.APIKey)
	body.APIKeyHeader = strings.TrimSpace(body.APIKeyHeader)
	body.ResultField = strings.TrimSpace(body.ResultField)
	if body.TTLSeconds < 0 {
		body.TTLSeconds = 0
	}
	if body.RequestsPerIP < 0 {
		body.RequestsPerIP = 0
	}
	h.cfg.DynamicProxy = body
	h.persist(c)
}

// ── DELETE /v0/management/dynamic-proxy ──────────────────────────────────────

// DeleteDynamicProxy 禁用并清空动态代理配置。
func (h *Handler) DeleteDynamicProxy(c *gin.Context) {
	h.cfg.DynamicProxy = config.DynamicProxyConfig{}
	h.persist(c)
}

// ── POST /v0/management/dynamic-proxy/test ───────────────────────────────────

// TestDynamicProxy 测试动态代理 API URL 是否可达，并返回提取到的代理 URL 预览。
func (h *Handler) TestDynamicProxy(c *gin.Context) {
	var body struct {
		APIURL       string `json:"api-url"`
		APIKey       string `json:"api-key"`
		APIKeyHeader string `json:"api-key-header"`
		ResultField  string `json:"result-field"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body", "message": err.Error()})
		return
	}

	apiURL := strings.TrimSpace(body.APIURL)
	if apiURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_api_url", "message": "api-url 不能为空"})
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, apiURL, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_url", "message": err.Error()})
		return
	}

	apiKeyHeader := strings.TrimSpace(body.APIKeyHeader)
	if apiKeyHeader == "" {
		apiKeyHeader = "X-API-Key"
	}
	if key := strings.TrimSpace(body.APIKey); key != "" {
		req.Header.Set(apiKeyHeader, key)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "request_failed", "message": err.Error()})
		return
	}
	defer resp.Body.Close()

	rawBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	rawText := strings.TrimSpace(string(rawBytes))

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "api_error",
			"status":  resp.StatusCode,
			"message": rawText,
		})
		return
	}

	proxyURL := dynProxyExtractURL(rawText, strings.TrimSpace(body.ResultField))
	if proxyURL == "" {
		c.JSON(http.StatusOK, gin.H{
			"ok":        false,
			"raw":       rawText,
			"proxy_url": "",
			"message":   "无法从响应中提取代理 URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"proxy_url": proxyURL,
		"raw":       rawText,
		"message":   "测试成功",
	})
}

// dynProxyExtractURL 从 API 响应文本中提取代理 URL（支持 JSON 路径或纯文本）。
func dynProxyExtractURL(text, resultField string) string {
	proxyURL := ""

	isJSON := strings.HasPrefix(text, "{") || strings.HasPrefix(text, "[")
	if isJSON || resultField != "" {
		var data any
		if jsonErr := json.Unmarshal([]byte(text), &data); jsonErr == nil {
			if resultField != "" {
				node := data
				for _, key := range strings.Split(resultField, ".") {
					if m, ok := node.(map[string]any); ok {
						node = m[key]
					} else {
						node = nil
						break
					}
				}
				if node != nil {
					proxyURL = strings.TrimSpace(fmt.Sprintf("%v", node))
				}
			} else if m, ok := data.(map[string]any); ok {
				for _, key := range []string{"proxy", "url", "proxy_url", "data", "ip"} {
					if v, found := m[key]; found {
						proxyURL = strings.TrimSpace(fmt.Sprintf("%v", v))
						break
					}
				}
			}
		}
	}

	if proxyURL == "" {
		proxyURL = text
	}

	if proxyURL == "" {
		return ""
	}

	if !strings.HasPrefix(proxyURL, "http://") &&
		!strings.HasPrefix(proxyURL, "https://") &&
		!strings.HasPrefix(proxyURL, "socks5://") {
		proxyURL = "http://" + proxyURL
	}

	return proxyURL
}
