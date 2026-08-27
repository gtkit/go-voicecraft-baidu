package voicecraftbaidu

import "testing"

func TestCodeDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		code      int
		wantKnown bool
		wantDesc  string
	}{
		{name: "success", code: CodeSuccess, wantKnown: true, wantDesc: "处理成功"},
		{name: "business code", code: CodeVoiceIDNotFound, wantKnown: true, wantDesc: "voice_id 不存在，请检查是否正确"},
		{name: "protocol code", code: CodeTTSVoiceNotFound, wantKnown: true},
		{name: "unknown code", code: 999999, wantKnown: false, wantDesc: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			desc, known := CodeDescription(tt.code)
			if known != tt.wantKnown {
				t.Fatalf("known = %v, want %v", known, tt.wantKnown)
			}
			if tt.wantDesc != "" && desc != tt.wantDesc {
				t.Errorf("desc = %q, want %q", desc, tt.wantDesc)
			}
			if !known && desc != "" {
				t.Errorf("unknown code should have empty desc, got %q", desc)
			}
		})
	}
}

func TestErrorDescription(t *testing.T) {
	t.Parallel()

	t.Run("APIError known code", func(t *testing.T) {
		t.Parallel()
		err := &APIError{Code: CodeVoiceIDNotFound}
		if got := err.Description(); got != "voice_id 不存在，请检查是否正确" {
			t.Errorf("desc = %q", got)
		}
	})

	t.Run("APIError unknown code", func(t *testing.T) {
		t.Parallel()
		err := &APIError{Code: 999999}
		if got := err.Description(); got != "" {
			t.Errorf("unknown code should be empty, got %q", got)
		}
	})

	t.Run("WebSocketError known code", func(t *testing.T) {
		t.Parallel()
		err := &WebSocketError{Code: CodeTextOverLimit}
		if got := err.Description(); got != "单次文本超过 1000 字" {
			t.Errorf("desc = %q", got)
		}
	})
}

// TestCodeDescriptionMapComplete 确保每个枚举码都在描述映射中（防止新增码漏登记）。
func TestCodeDescriptionMapComplete(t *testing.T) {
	t.Parallel()

	codes := []int{
		CodeSuccess, CodeUserConcurrencyLimit, CodeUserQuotaExceeded, CodeServiceTemporaryErr,
		CodeVoiceTextMismatch, CodeTextIDExpired, CodeAudioDownloadLimit, CodeVoiceNotFound,
		CodeInvalidAudioFormat, CodeAudioSensitivePerson, CodeSynthesisException, CodeQuotaThrottle,
		CodeConcurrencyThrottle, CodeTextTooLong, CodeVoiceIDInvalid, CodeMissingAuth,
		CodeServiceRetry, CodeMissingParameter, CodeInvalidParameter, CodeVoiceIDNotFound,
		CodeInvalidPage, CodeInvalidBase64Audio, CodeNoDataPermission, CodeTextSensitive,
		CodeFileDownloadFailed, CodeAudioTooShort, CodeWERCheckFailed, CodeAudioSNRFailed,
		CodeRecognitionFailed, CodeAudioLevelFailed, CodeAudioSpeedFailed, CodeAudioQualityPoor,
		CodeNoAPIPermission, CodeOpenAPIConcurrencyLimit, CodeOpenAPIUsageLimit, CodeAccessTokenInvalid,
		CodeAccessTokenExpired, CodeIAMAuthError, CodeInvalidValue, CodeMissingRequiredParam,
		CodeTextOverLimit, CodePendingTextTooLong, CodeTTSVoiceNoPermission, CodeTTSVoiceNotFound,
		CodeOpenAPIUsageLimitReached, CodeInternalError, CodeVoiceServiceNotStarted,
	}

	for _, code := range codes {
		if _, known := CodeDescription(code); !known {
			t.Errorf("code %d 未登记到 codeDescriptions", code)
		}
	}
}
