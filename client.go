package voicraftbaidu

import (
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gtkit/httpc"
	"golang.org/x/sync/singleflight"
)

// Client 是百度智能云声音复刻服务的客户端。
//
// Client 在创建后是并发安全且不可变的（immutable after New），
// 可以在多个 goroutine 中共享使用。
//
// 使用示例：
//
//	// 方式一：使用 client_id + client_secret（推荐，SDK 自动管理 token）
//	client, err := voicraftbaidu.New(
//	    voicraftbaidu.WithClientCredentials("your-client-id", "your-client-secret"),
//	)
//
//	// 方式二：使用 API Key
//	client, err := voicraftbaidu.New(
//	    voicraftbaidu.WithAPIKey("your-api-key"),
//	)
type Client struct {
	httpClient   *http.Client  // 底层 HTTP 客户端，供 WebSocket dialer 继承传输配置
	rest         *httpc.Client // 基于 httpClient 构造，承载所有 REST/JSON 请求
	authMode     AuthMode
	apiKey       string // AuthAPIKey 模式下使用（用户传入的原值）
	apiKeyHeader string // AuthAPIKey 模式下的 Authorization 头值（含 "Bearer " 前缀）
	clientID     string // AuthAccessToken 模式下使用
	clientSecret string // AuthAccessToken 模式下使用
	baseURL      string // API 基础地址

	// token 管理（仅 AuthAccessToken 模式）
	token      atomic.Pointer[tokenCache] // 无锁读取缓存的 token
	tokenGroup singleflight.Group         // 防止并发刷新 token

	// WebSocket TTS 配置
	idleTimeout int // WebSocket 空闲超时（秒）
}

// New 创建一个新的 Client 实例。
//
// 必须通过 WithClientCredentials 或 WithAPIKey 配置鉴权方式，
// 否则返回 ErrNoAuth。
//
// 示例：
//
//	client, err := voicraftbaidu.New(
//	    voicraftbaidu.WithClientCredentials(clientID, clientSecret),
//	    voicraftbaidu.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
//	    voicraftbaidu.WithIdleTimeout(120),
//	)
func New(opts ...Option) (*Client, error) {
	c := &Client{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		baseURL:     defaultBaseURL,
		idleTimeout: defaultIdleTimeout,
	}

	for _, opt := range opts {
		opt(c)
	}

	// 校验鉴权配置
	if c.apiKey == "" && (c.clientID == "" || c.clientSecret == "") {
		return nil, ErrNoAuth
	}

	// 确定鉴权模式
	if c.apiKey != "" {
		c.authMode = AuthAPIKey
	} else {
		c.authMode = AuthAccessToken
	}

	// 基于最终的 httpClient 构造 REST 客户端，统一 HTTP/JSON 请求路径。
	// 响应体上限取音频上限（兼顾非流式合成的二进制音频）；API Key 鉴权通过默认头透传。
	restOpts := []httpc.Option{
		httpc.WithHTTPClient(c.httpClient),
		httpc.WithMaxResponseBytes(maxSynthesizeAudioBytes),
	}
	if c.authMode == AuthAPIKey {
		// 百度要求 Authorization 头为 "Bearer <API Key>"；用户已带前缀则不重复添加。
		c.apiKeyHeader = c.apiKey
		if !strings.HasPrefix(c.apiKeyHeader, "Bearer ") {
			c.apiKeyHeader = "Bearer " + c.apiKeyHeader
		}
		restOpts = append(restOpts, httpc.WithDefaultHeaders(map[string]string{
			"Authorization": c.apiKeyHeader,
		}))
	}
	c.rest = httpc.New(restOpts...)

	return c, nil
}

// ============================================================================
// Functional Options
// ============================================================================

// Option 是 Client 的配置选项函数。
type Option func(*Client)

// WithHTTPClient 设置自定义的 HTTP 客户端。
// 可用于配置代理、自定义 TLS、超时等。
//
// 默认值：&http.Client{Timeout: 30 * time.Second}
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithAPIKey 设置 API Key 鉴权模式。
// 使用此模式时，SDK 通过 Authorization 头以 "Bearer <API Key>" 形式传递；
// 传入的 key 已带 "Bearer " 前缀时不会重复添加。
func WithAPIKey(key string) Option {
	return func(c *Client) {
		c.apiKey = key
	}
}

// WithClientCredentials 设置 OAuth client_credentials 鉴权模式。
// SDK 将自动获取和续期 access_token。
func WithClientCredentials(clientID, clientSecret string) Option {
	return func(c *Client) {
		c.clientID = clientID
		c.clientSecret = clientSecret
	}
}

// WithBaseURL 自定义 API 基础地址。
// 默认值：https://aip.baidubce.com
//
// 主要用于测试环境或私有化部署。
func WithBaseURL(url string) Option {
	return func(c *Client) {
		if url != "" {
			c.baseURL = url
		}
	}
}

// WithIdleTimeout 设置 WebSocket 连接的空闲超时时间（秒）。
// 取值范围 [5, 600]，默认 60。
//
// 当 WebSocket 连接在指定时间内无任何消息交互时，服务端将主动断开连接。
func WithIdleTimeout(seconds int) Option {
	return func(c *Client) {
		if seconds >= 5 && seconds <= 600 {
			c.idleTimeout = seconds
		}
	}
}
