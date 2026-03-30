// Package config provides configuration management for the CLI Proxy API server.
// It handles loading and parsing YAML configuration files, and provides structured
// access to application settings including server port, authentication directory,
// debug settings, proxy configuration, and API keys.
package config

// SDKConfig represents the application's configuration, loaded from a YAML file.
type SDKConfig struct {
	// ProxyURL is the URL of an optional proxy server to use for outbound requests.
	ProxyURL string `yaml:"proxy-url" json:"proxy-url"`

	// ForceModelPrefix requires explicit model prefixes (e.g., "teamA/gemini-3-pro-preview")
	// to target prefixed credentials. When false, unprefixed model requests may use prefixed
	// credentials as well.
	ForceModelPrefix bool `yaml:"force-model-prefix" json:"force-model-prefix"`

	// RequestLog enables or disables detailed request logging functionality.
	RequestLog bool `yaml:"request-log" json:"request-log"`

	// APIKeys is a list of keys for authenticating clients to this proxy server.
	APIKeys []string `yaml:"api-keys" json:"api-keys"`

	// PassthroughHeaders controls whether upstream response headers are forwarded to downstream clients.
	// Default is false (disabled).
	PassthroughHeaders bool `yaml:"passthrough-headers" json:"passthrough-headers"`

	// Streaming configures server-side streaming behavior (keep-alives and safe bootstrap retries).
	Streaming StreamingConfig `yaml:"streaming" json:"streaming"`

	// NonStreamKeepAliveInterval controls how often blank lines are emitted for non-streaming responses.
	// <= 0 disables keep-alives. Value is in seconds.
	NonStreamKeepAliveInterval int `yaml:"nonstream-keepalive-interval,omitempty" json:"nonstream-keepalive-interval,omitempty"`

	// DynamicProxy configures the dynamic proxy provider for outbound requests.
	// When enabled, all executor requests route through a dynamically fetched proxy IP.
	// Priority: DynamicProxy > auth.ProxyURL (per-credential) > ProxyURL (global static).
	DynamicProxy DynamicProxyConfig `yaml:"dynamic-proxy" json:"dynamic-proxy"`
}

// DynamicProxyConfig configures dynamic proxy IP fetching for all outbound executor requests.
// The proxy IP is obtained from an external HTTP API and cached with TTL-based expiry.
// Optionally, a request-count limit per IP (requests-per-ip) can trigger early rotation
// before TTL expiry, aligning with codex-console's quantity-driven IP rotation strategy.
type DynamicProxyConfig struct {
	// Enable toggles dynamic proxy fetching. When false, static ProxyURL is used.
	Enable bool `yaml:"enable" json:"enable"`

	// APIURL is the HTTP endpoint that returns the proxy URL.
	// The response may be a plain proxy URL string (e.g. "http://user:pass@host:port")
	// or a JSON object. Use ResultField to extract from JSON.
	APIURL string `yaml:"api-url" json:"api-url"`

	// APIKey is an optional authentication key sent to the proxy API endpoint.
	APIKey string `yaml:"api-key" json:"api-key"`

	// APIKeyHeader is the HTTP header name used to carry APIKey.
	// Defaults to "X-API-Key" when empty.
	APIKeyHeader string `yaml:"api-key-header" json:"api-key-header"`

	// ResultField is a dot-separated JSON path to extract the proxy URL from a JSON response.
	// Example: "data.proxy". Leave empty to use the raw response body as the proxy URL.
	ResultField string `yaml:"result-field" json:"result-field"`

	// TTLSeconds overrides the proxy cache lifetime in seconds.
	// When 0, TTL is inferred from a "time=N" query parameter in APIURL (N minutes -> (N-1)*60 seconds).
	// Falls back to 540 seconds (9 minutes) when neither source provides a value.
	TTLSeconds int `yaml:"ttl-seconds" json:"ttl-seconds"`

	// RequestsPerIP sets the maximum number of requests a single proxy IP may serve
	// before the manager fetches a new IP, regardless of TTL.
	// Set to 0 to rely on TTL-only rotation (no count-based limit).
	// Mirrors codex-console's quantity-driven IP rotation (set_proxy_batch_size).
	RequestsPerIP int `yaml:"requests-per-ip" json:"requests-per-ip"`
}

// StreamingConfig holds server streaming behavior configuration.
type StreamingConfig struct {
	// KeepAliveSeconds controls how often the server emits SSE heartbeats (": keep-alive\n\n").
	// <= 0 disables keep-alives. Default is 0.
	KeepAliveSeconds int `yaml:"keepalive-seconds,omitempty" json:"keepalive-seconds,omitempty"`

	// BootstrapRetries controls how many times the server may retry a streaming request before any bytes are sent,
	// to allow auth rotation / transient recovery.
	// <= 0 disables bootstrap retries. Default is 0.
	BootstrapRetries int `yaml:"bootstrap-retries,omitempty" json:"bootstrap-retries,omitempty"`
}
