package executor

// dynamic_proxy_manager.go
// 动态代理管理器 —— Go 版 codex-console dynamic_proxy.py
//
// 功能：
//   - 通过外部 HTTP API 获取代理 URL（支持纯文本 / JSON 路径提取）
//   - TTL 缓存：优先使用 config.TTLSeconds，其次从 API URL time=N 参数推算，默认 540s
//   - 数量驱动轮换：当 requests-per-ip > 0 时，每个 IP 最多服务 N 次请求即换新 IP
//   - 线程安全：读写均受 sync.Mutex 保护
//   - 作用范围：对所有 executor 的所有凭证生效（globalDynamicProxyManager 全局单例）

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	log "github.com/sirupsen/logrus"
)

// dynamicProxyCache 保存当前动态代理的缓存状态。
type dynamicProxyCache struct {
	mu        sync.Mutex
	url       string        // 当前代理 URL
	transport *http.Transport // 与 url 绑定的 transport，URL 变化时重建
	expiresAt time.Time     // TTL 过期时间
	remaining int           // 剩余可服务请求数（RequestsPerIP 驱动，<=0 时不限制）
	cacheKey  string        // 配置指纹，配置变动时强制刷新
}

// globalDynamicProxyManager 是进程级别的全局动态代理单例。
var globalDynamicProxyManager dynamicProxyCache

// reTimeParam 匹配 API URL 中的 time=N 查询参数。
var reTimeParam = regexp.MustCompile(`[?&]time=(\d+)`)

// dynamicCacheKey 生成配置指纹，用于检测配置变更后强制刷新缓存。
func dynamicCacheKey(cfg *config.DynamicProxyConfig) string {
	return fmt.Sprintf("%s|%s|%s|%s|%d|%d",
		cfg.APIURL,
		cfg.APIKey,
		cfg.APIKeyHeader,
		cfg.ResultField,
		cfg.TTLSeconds,
		cfg.RequestsPerIP,
	)
}

// parseDynamicProxyTTL 计算代理缓存 TTL。
// 优先级：config.TTLSeconds > URL time=N 参数（分钟，提前 60s 过期）> 默认 540s。
func parseDynamicProxyTTL(cfg *config.DynamicProxyConfig) time.Duration {
	if cfg.TTLSeconds > 0 {
		return time.Duration(cfg.TTLSeconds) * time.Second
	}
	if m := reTimeParam.FindStringSubmatch(cfg.APIURL); len(m) == 2 {
		if minutes, err := strconv.Atoi(m[1]); err == nil && minutes > 0 {
			secs := minutes*60 - 60
			if secs < 60 {
				secs = 60
			}
			return time.Duration(secs) * time.Second
		}
	}
	return 540 * time.Second
}

// fetchDynamicProxyURL 通过 HTTP 请求向 API 获取代理 URL。
// 支持纯文本响应和 JSON 路径提取（ResultField 点号分隔路径）。
func fetchDynamicProxyURL(cfg *config.DynamicProxyConfig) (string, error) {
	req, err := http.NewRequest(http.MethodGet, cfg.APIURL, nil)
	if err != nil {
		return "", fmt.Errorf("dynamic proxy: build request failed: %w", err)
	}

	// 注入 API Key 请求头
	if cfg.APIKey != "" {
		header := strings.TrimSpace(cfg.APIKeyHeader)
		if header == "" {
			header = "X-API-Key"
		}
		req.Header.Set(header, cfg.APIKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("dynamic proxy: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("dynamic proxy: API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return "", fmt.Errorf("dynamic proxy: read response failed: %w", err)
	}
	text := strings.TrimSpace(string(body))

	// 尝试 JSON 解析
	proxyURL := ""
	resultField := strings.TrimSpace(cfg.ResultField)
	isJSON := strings.HasPrefix(text, "{") || strings.HasPrefix(text, "[")

	if isJSON || resultField != "" {
		var data any
		if jsonErr := json.Unmarshal([]byte(text), &data); jsonErr == nil {
			if resultField != "" {
				// 点号路径提取，如 "data.proxy"
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
				// 自动搜索常见字段
				for _, key := range []string{"proxy", "url", "proxy_url", "data", "ip"} {
					if v, found := m[key]; found {
						proxyURL = strings.TrimSpace(fmt.Sprintf("%v", v))
						break
					}
				}
			}
		}
	}

	// 回退：响应原文作为代理 URL
	if proxyURL == "" {
		proxyURL = text
	}

	if proxyURL == "" {
		return "", fmt.Errorf("dynamic proxy: API returned empty proxy URL")
	}

	// 确保有协议前缀
	if !strings.HasPrefix(proxyURL, "http://") &&
		!strings.HasPrefix(proxyURL, "https://") &&
		!strings.HasPrefix(proxyURL, "socks5://") {
		proxyURL = "http://" + proxyURL
	}

	return proxyURL, nil
}

// getDynamicProxyTransport 返回当前有效的动态代理 transport。
// 当缓存有效时直接复用；当请求数耗尽或 TTL 过期时重新获取新 IP。
// 返回 nil 表示动态代理不可用，调用方应回退到静态代理。
func getDynamicProxyTransport(cfg *config.Config) *http.Transport {
	if cfg == nil || !cfg.DynamicProxy.Enable || strings.TrimSpace(cfg.DynamicProxy.APIURL) == "" {
		return nil
	}

	dynCfg := &cfg.DynamicProxy
	cacheKey := dynamicCacheKey(dynCfg)
	now := time.Now()

	m := &globalDynamicProxyManager
	m.mu.Lock()
	defer m.mu.Unlock()

	// 判断当前缓存是否有效
	countOK := dynCfg.RequestsPerIP <= 0 || m.remaining > 0
	ttlOK := !m.expiresAt.IsZero() && now.Before(m.expiresAt)
	configOK := m.cacheKey == cacheKey

	if m.url != "" && countOK && ttlOK && configOK {
		// 缓存有效：消耗一个名额
		if dynCfg.RequestsPerIP > 0 {
			m.remaining--
			log.Debugf("dynamic proxy: reusing cached IP (remaining=%d/%d)", m.remaining, dynCfg.RequestsPerIP)
		}
		return m.transport
	}

	// 需要获取新 IP
	reason := "cache miss"
	if m.url != "" {
		if !configOK {
			reason = "config changed"
		} else if !ttlOK {
			reason = "TTL expired"
		} else if dynCfg.RequestsPerIP > 0 && m.remaining <= 0 {
			reason = fmt.Sprintf("requests-per-ip exhausted (%d)", dynCfg.RequestsPerIP)
		}
	}
	log.Infof("dynamic proxy: fetching new proxy IP (%s)", reason)

	newURL, err := fetchDynamicProxyURL(dynCfg)
	if err != nil {
		log.Errorf("dynamic proxy: failed to fetch proxy IP: %v", err)
		// 若旧缓存仍在 TTL 内且配置未变，降级继续使用
		if m.url != "" && ttlOK && configOK {
			log.Warnf("dynamic proxy: falling back to previous cached IP")
			if dynCfg.RequestsPerIP > 0 {
				m.remaining--
			}
			return m.transport
		}
		return nil
	}

	ttl := parseDynamicProxyTTL(dynCfg)
	transport := buildProxyTransport(newURL)
	if transport == nil {
		log.Errorf("dynamic proxy: failed to build transport for %s", newURL)
		return nil
	}

	m.url = newURL
	m.transport = transport
	m.expiresAt = now.Add(ttl)
	m.cacheKey = cacheKey
	if dynCfg.RequestsPerIP > 0 {
		// 本次调用消耗 1 个名额
		m.remaining = dynCfg.RequestsPerIP - 1
		log.Infof("dynamic proxy: new IP assigned, requests-per-ip=%d, TTL=%s, url=%s",
			dynCfg.RequestsPerIP, ttl, maskProxyURL(newURL))
	} else {
		m.remaining = 0
		log.Infof("dynamic proxy: new IP assigned, TTL=%s, url=%s", ttl, maskProxyURL(newURL))
	}

	return m.transport
}

// maskProxyURL 遮蔽代理 URL 中的密码，用于日志输出。
// 例：http://user:pass@1.2.3.4:8080 -> http://user:***@1.2.3.4:8080
func maskProxyURL(raw string) string {
	if idx := strings.Index(raw, "@"); idx > 0 {
		scheme := ""
		rest := raw
		if i := strings.Index(raw, "://"); i >= 0 {
			scheme = raw[:i+3]
			rest = raw[i+3:]
		}
		// 在 @ 之前遮蔽 :password
		creds := rest[:strings.Index(rest, "@")]
		if ci := strings.Index(creds, ":"); ci >= 0 {
			creds = creds[:ci+1] + "***"
		}
		return scheme + creds + rest[strings.Index(rest, "@"):]
	}
	return raw
}
