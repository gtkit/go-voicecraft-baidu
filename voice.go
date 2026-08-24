package voicraftbaidu

import (
	"context"
	"fmt"
	"net/http"
)

// CreateVoice 通过上传训练音频创建音色。
//
// 支持两种音频上传方式：
//   - AudioURL: 提供音频文件的公网链接
//   - AudioFile: 提供音频文件的 base64 编码内容
//
// 两者同时传入时，以 AudioFile 为准。
//
// 注意：通过此接口创建的音色，若 1 年内没有调用合成记录，该音色将被自动删除。
//
// 约定：err == nil 时表示成功，此时 resp 非 nil 且 resp.Status == 0；
// 任意失败（HTTP 错误、JSON 解析失败或业务 status != 0）时仅返回非 nil 的 error，
// resp 为 nil，请用 errors.Is/IsAPIError 等判断错误类型，勿在未检查 err 时使用 resp。
//
// 使用示例：
//
//	resp, err := client.CreateVoice(ctx, &voicraftbaidu.CreateVoiceRequest{
//	    VoiceName: "my-voice",
//	    VoiceDesc: "温柔细腻的音色",
//	    AudioURL:  "https://example.com/audio.wav",
//	    Lang:      voicraftbaidu.LangChinese,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("voice_id: %d\n", resp.Data.VoiceID)
func (c *Client) CreateVoice(ctx context.Context, req *CreateVoiceRequest) (*CreateVoiceResponse, error) {
	// 参数校验
	if err := req.validate(); err != nil {
		return nil, err
	}

	// 构建鉴权 query（access_token 模式追加 query；API Key 模式经默认头鉴权）
	authQuery, err := c.buildAuthQuery(ctx)
	if err != nil {
		return nil, fmt.Errorf("voicraftbaidu: auth failed: %w", err)
	}

	reqURL := c.baseURL + createVoiceEndpoint
	if encoded := authQuery.Encode(); encoded != "" {
		reqURL += "?" + encoded
	}

	var resp CreateVoiceResponse
	status, err := c.rest.PostJSON(ctx, reqURL, req, &resp)
	if err != nil {
		return nil, fmt.Errorf("voicraftbaidu: request failed: %w", err)
	}

	// 业务层面错误（百度以 HTTP 200 + status != 0 表达）
	if resp.Status != 0 {
		return nil, &APIError{StatusCode: status, Code: resp.Status, Message: resp.Message}
	}
	// HTTP 层面错误但响应体未携带业务 status
	if status != http.StatusOK {
		return nil, &APIError{StatusCode: status, Message: fmt.Sprintf("unexpected http status %d", status)}
	}

	return &resp, nil
}

// GetCloneText 获取一段用于音色复刻的训练文本。
//
// 返回的 TextID 有效期 24 小时，朗读 Text 内容录制音频后，
// 可在 CreateVoice 时通过 CreateVoiceRequest.TextID 关联以提升复刻质量。
// 使用自定义文本复刻时无需调用此接口。
//
// 失败时返回 *APIError（服务端 status != 0）。
func (c *Client) GetCloneText(ctx context.Context) (*CloneTextResponse, error) {
	var resp CloneTextResponse
	if err := c.postClone(ctx, cloneTextEndpoint, struct{}{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListVoices 分页查询当前用户已创建的音色列表。
//
// page 为页码，须 ≥ 1；传入 0 或负数时省略该参数，由服务端返回默认页。
//
// 失败时返回 *APIError（服务端 status != 0）。
func (c *Client) ListVoices(ctx context.Context, page int) (*ListVoicesResponse, error) {
	req := listVoicesRequest{}
	if page > 0 {
		req.Page = page
	}
	var resp ListVoicesResponse
	if err := c.postClone(ctx, listVoicesEndpoint, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// VoiceDetail 查询单个音色的详细信息。
//
// voiceID 是通过 CreateVoice 创建的音色 ID，必填。
//
// 失败时返回 *ValidationError（voiceID ≤ 0）或 *APIError（服务端 status != 0）。
func (c *Client) VoiceDetail(ctx context.Context, voiceID int) (*VoiceDetailResponse, error) {
	if voiceID <= 0 {
		return nil, &ValidationError{Field: "voice_id", Reason: "voice_id is required"}
	}
	var resp VoiceDetailResponse
	if err := c.postClone(ctx, voiceDetailEndpoint, voiceIDRequest{VoiceID: voiceID}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteVoice 删除指定音色。
//
// voiceID 是通过 CreateVoice 创建的音色 ID，必填。
// 删除成功返回 nil；失败返回 *ValidationError（voiceID ≤ 0）或 *APIError（服务端 status != 0）。
func (c *Client) DeleteVoice(ctx context.Context, voiceID int) error {
	if voiceID <= 0 {
		return &ValidationError{Field: "voice_id", Reason: "voice_id is required"}
	}
	var resp deleteVoiceResponse
	return c.postClone(ctx, deleteVoiceEndpoint, voiceIDRequest{VoiceID: voiceID}, &resp)
}

// postClone 向音色管理类 REST 接口发送 JSON 请求并解码响应。
//
// 这些接口（获取训练文本 / 列表 / 详情 / 删除）共享同一交互模式：
// POST JSON + 鉴权 query/header，响应体形如 {"status":..,"message":..,"data":..}，
// status != 0 表示业务错误。out 必须是指向内嵌 cloneHeader 的响应结构的指针。
func (c *Client) postClone(ctx context.Context, endpoint string, reqBody any, out cloneResult) error {
	authQuery, err := c.buildAuthQuery(ctx)
	if err != nil {
		return fmt.Errorf("voicraftbaidu: auth failed: %w", err)
	}

	reqURL := c.baseURL + endpoint
	if encoded := authQuery.Encode(); encoded != "" {
		reqURL += "?" + encoded
	}

	status, err := c.rest.PostJSON(ctx, reqURL, reqBody, out)
	if err != nil {
		return fmt.Errorf("voicraftbaidu: request failed: %w", err)
	}

	// 业务层面错误（百度以 HTTP 200 + status != 0 表达）
	if s, message := out.result(); s != 0 {
		return &APIError{StatusCode: status, Code: s, Message: message}
	}
	// HTTP 层面错误但响应体未携带业务 status
	if status != http.StatusOK {
		return &APIError{StatusCode: status, Message: fmt.Sprintf("unexpected http status %d", status)}
	}

	return nil
}
