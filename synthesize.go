package voicraftbaidu

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gtkit/json/v2"
)

// Synthesize 通过声音复刻非流式在线合成接口，将文本一次性合成为完整音频。
//
// 与 NewTTSSession（声音复刻流式合成）不同，此方法是单次 REST 请求，
// 直接返回完整的音频字节，适合短文本、无需边合成边播放的场景。
//
// voiceID 是通过 CreateVoice 创建的音色 ID，必填。
// text 是待合成文本，必填，不超过 500 个字符（按 rune 计算，中文算一个字符）。
// cfg 可为 nil，使用服务端默认参数（wav 格式、语速/音调/音量均为 5）。
// 其中 cfg.Emotion 为非流式合成专用字段（happy / surprise / angry / disgust）。
//
// 返回的字节是 cfg.MediaType 指定格式的音频（默认 wav）；
// 失败时返回 *APIError（服务端 status != 0）或 *ValidationError（参数非法）。
//
// 使用示例：
//
//	audio, err := client.Synthesize(ctx, 100001, "你好，世界。", &voicraftbaidu.TTSConfig{
//	    MediaType: voicraftbaidu.MediaMP3,
//	    Emotion:   voicraftbaidu.EmotionHappy,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("output.mp3", audio, 0o644)
func (c *Client) Synthesize(ctx context.Context, voiceID int, text string, cfg *TTSConfig) ([]byte, error) {
	if err := validateSynthesize(voiceID, text, cfg); err != nil {
		return nil, err
	}

	fields, err := buildSynthesizeBody(voiceID, text, cfg)
	if err != nil {
		return nil, fmt.Errorf("voicraftbaidu: marshal request: %w", err)
	}

	// 构建鉴权 query（access_token 模式追加 query；API Key 模式经默认头鉴权）
	authQuery, err := c.buildAuthQuery(ctx)
	if err != nil {
		return nil, fmt.Errorf("voicraftbaidu: auth failed: %w", err)
	}

	reqURL := c.baseURL + synthesizeEndpoint
	if encoded := authQuery.Encode(); encoded != "" {
		reqURL += "?" + encoded
	}

	// 成功返回二进制音频、失败返回 JSON 错误体，用 Raw 取原始字节并按 Content-Type 区分。
	// 显式以 Accept: */* 覆盖 httpc 默认的 application/json，避免内容协商干扰二进制音频返回。
	header, respBody, status, err := c.rest.RequestRawWithHeader(
		ctx, http.MethodPost, reqURL, map[string]string{"Accept": "*/*"}, fields)
	if err != nil {
		return nil, fmt.Errorf("voicraftbaidu: request failed: %w", err)
	}

	if status == http.StatusOK && strings.HasPrefix(header.Get("Content-Type"), "audio") {
		return respBody, nil
	}

	// 错误响应：{"status": 非0, "message": "..."}
	var errResp struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}
	if jsonErr := json.Unmarshal(respBody, &errResp); jsonErr == nil &&
		(errResp.Status != 0 || errResp.Message != "") {
		return nil, &APIError{
			StatusCode: status,
			Code:       errResp.Status,
			Message:    errResp.Message,
		}
	}

	return nil, fmt.Errorf("voicraftbaidu: unexpected response (http=%d, content-type=%q): %s",
		status, header.Get("Content-Type"), string(respBody))
}

// validateSynthesize 校验非流式合成的请求参数。
func validateSynthesize(voiceID int, text string, cfg *TTSConfig) error {
	if voiceID <= 0 {
		return &ValidationError{Field: "voice_id", Reason: "voice_id is required"}
	}
	if text == "" {
		return &ValidationError{Field: "text", Reason: "text is empty"}
	}
	if utf8.RuneCountInString(text) > maxSynthesizeTextLength {
		return &ValidationError{Field: "text", Reason: "text exceeds 500 characters limit"}
	}
	if cfg != nil && cfg.Emotion != "" {
		switch cfg.Emotion {
		case EmotionHappy, EmotionSurprise, EmotionAngry, EmotionDisgust:
		default:
			return &ValidationError{Field: "emotion", Reason: "emotion must be one of happy/surprise/angry/disgust"}
		}
	}
	return nil
}

// buildSynthesizeBody 构造非流式合成的请求体字段。
//
// 复用 TTSConfig 的 JSON 编码（含“显式零值”能力），再注入 text 与 voice_id，
// 避免重复维护一套合成参数的序列化逻辑。返回 map 交由 httpc 统一序列化。
func buildSynthesizeBody(voiceID int, text string, cfg *TTSConfig) (map[string]any, error) {
	fields := map[string]any{}
	if cfg != nil {
		cfgJSON, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(cfgJSON, &fields); err != nil {
			return nil, err
		}
	}
	fields["text"] = text
	fields["voice_id"] = voiceID
	return fields, nil
}
