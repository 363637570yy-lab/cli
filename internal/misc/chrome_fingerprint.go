// Package misc provides miscellaneous utility functions for the CLIProxyAPI server.
// chrome_fingerprint.go: Chrome 浏览器指纹生成器，对齐 codex-console 的请求头随机化策略。
// 用于需要模拟真实浏览器行为的 HTTP 请求（OpenAI / Anthropic Amp 等端点）。
package misc

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
)

// chromeProfile 描述一个 Chrome 版本的指纹特征
type chromeProfile struct {
	major     int
	buildBase int
	patchMin  int
	patchMax  int
	secChUA   string
}

// chromeProfiles 对齐 codex-console 的 _CHROME_PROFILES，覆盖 Chrome 131/133/136 三个版本
var chromeProfiles = []chromeProfile{
	{
		major:     131,
		buildBase: 6778,
		patchMin:  69,
		patchMax:  205,
		secChUA:   `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`,
	},
	{
		major:     133,
		buildBase: 6943,
		patchMin:  33,
		patchMax:  153,
		secChUA:   `"Not(A:Brand";v="99", "Google Chrome";v="133", "Chromium";v="133"`,
	},
	{
		major:     136,
		buildBase: 7103,
		patchMin:  48,
		patchMax:  175,
		secChUA:   `"Chromium";v="136", "Google Chrome";v="136", "Not.A/Brand";v="99"`,
	},
}

// chromeAcceptLanguages 对齐 codex-console 的 _ACCEPT_LANGUAGES
var chromeAcceptLanguages = []string{
	"en-US,en;q=0.9",
	"en-US,en;q=0.9,zh-CN;q=0.8",
	"en,en-US;q=0.9",
	"en-US,en;q=0.8",
}

// ChromeFingerprint 保存一次随机化的 Chrome 浏览器指纹
type ChromeFingerprint struct {
	Major              int
	FullVersion        string // e.g. "133.0.6943.98"
	UserAgent          string
	SecChUA            string
	SecChUAFullVerList string
	PlatformVersion    string // e.g. "12.0.0"
	AcceptLanguage     string
}

// buildSecChUAFullVersionList 根据 sec-ch-ua 生成 sec-ch-ua-full-version-list。
// 对齐 codex-console 的 _build_sec_ch_ua_full_version_list()。
func buildSecChUAFullVersionList(secChUA, fullVer string) string {
	// 简单解析 brand;v="X" 格式
	entries := []string{}
	parts := strings.Split(secChUA, ", ")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// brand 是 "Not..." 类，保持版本号不变
		if strings.Contains(part, "Not") || strings.Contains(part, "Brand") {
			// 替换短版本为 major.0.0.0
			idx := strings.Index(part, `";v="`)
			if idx == -1 {
				entries = append(entries, part)
				continue
			}
			brand := part[:idx+1] // e.g. "Not_A Brand"
			ver := part[idx+5:]   // after `";v="`
			ver = strings.TrimSuffix(ver, `"`)
			entries = append(entries, fmt.Sprintf(`%s";v="%s.0.0.0"`, brand, ver))
		} else {
			// Google Chrome / Chromium → 完整版本号
			idx := strings.Index(part, `";v="`)
			if idx == -1 {
				entries = append(entries, part)
				continue
			}
			brand := part[:idx+1]
			entries = append(entries, fmt.Sprintf(`%s";v="%s"`, brand, fullVer))
		}
	}
	return strings.Join(entries, ", ")
}

// NewChromeFingerprint 生成一个随机化的 Chrome 指纹，对齐 codex-console _random_chrome_profile()
func NewChromeFingerprint() *ChromeFingerprint {
	p := chromeProfiles[rand.Intn(len(chromeProfiles))]
	patch := p.patchMin + rand.Intn(p.patchMax-p.patchMin+1)
	fullVer := fmt.Sprintf("%d.0.%d.%d", p.major, p.buildBase, patch)
	ua := fmt.Sprintf(
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36",
		fullVer,
	)
	platformVer := fmt.Sprintf("%d.0.0", 10+rand.Intn(6)) // Windows 10.0.0 ~ 15.0.0
	fullVerList := buildSecChUAFullVersionList(p.secChUA, fullVer)
	lang := chromeAcceptLanguages[rand.Intn(len(chromeAcceptLanguages))]

	return &ChromeFingerprint{
		Major:              p.major,
		FullVersion:        fullVer,
		UserAgent:          ua,
		SecChUA:            p.secChUA,
		SecChUAFullVerList: fullVerList,
		PlatformVersion:    platformVer,
		AcceptLanguage:     lang,
	}
}

// InjectBrowserFingerprint 将 Chrome 浏览器指纹注入 HTTP 请求头。
// 对齐 codex-console OpenAIHTTPClient._build_default_headers()。
// 若 fp 为 nil，则自动生成一个随机指纹。
// 仅在目标请求头为空或明确为非浏览器 UA 时覆盖，不破坏已有浏览器头。
func InjectBrowserFingerprint(req *http.Request, fp *ChromeFingerprint) {
	if req == nil {
		return
	}
	if fp == nil {
		fp = NewChromeFingerprint()
	}

	// 只有在 UA 为空或 Go 默认 UA 时才覆盖
	existingUA := req.Header.Get("User-Agent")
	if existingUA == "" || strings.HasPrefix(existingUA, "Go-http-client") {
		req.Header.Set("User-Agent", fp.UserAgent)
	}

	// sec-ch-ua 族：补充缺失的头
	if req.Header.Get("sec-ch-ua") == "" {
		req.Header.Set("sec-ch-ua", fp.SecChUA)
		req.Header.Set("sec-ch-ua-mobile", "?0")
		req.Header.Set("sec-ch-ua-platform", `"Windows"`)
		req.Header.Set("sec-ch-ua-arch", `"x86"`)
		req.Header.Set("sec-ch-ua-bitness", `"64"`)
		req.Header.Set("sec-ch-ua-full-version", fmt.Sprintf(`"%s"`, fp.FullVersion))
		req.Header.Set("sec-ch-ua-platform-version", fmt.Sprintf(`"%s"`, fp.PlatformVersion))
		req.Header.Set("sec-ch-ua-full-version-list", fp.SecChUAFullVerList)
	}

	// Fetch meta
	if req.Header.Get("Sec-Fetch-Dest") == "" {
		req.Header.Set("Sec-Fetch-Dest", "empty")
		req.Header.Set("Sec-Fetch-Mode", "cors")
		req.Header.Set("Sec-Fetch-Site", "same-site")
	}

	// Accept-Encoding：保证 gzip/br 在场
	if req.Header.Get("Accept-Encoding") == "" {
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	}

	// Accept-Language
	if req.Header.Get("Accept-Language") == "" {
		req.Header.Set("Accept-Language", fp.AcceptLanguage)
	}
}
