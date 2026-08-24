package voicraftbaidu

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gtkit/json/v2"
)

func TestValidateSynthesize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		voiceID   int
		text      string
		cfg       *TTSConfig
		wantField string // 非空表示期望 ValidationError 且 Field 匹配
	}{
		{name: "valid without cfg", voiceID: 100001, text: "你好"},
		{name: "valid with emotion", voiceID: 100001, text: "你好", cfg: &TTSConfig{Emotion: EmotionHappy}},
		{name: "valid with empty emotion", voiceID: 1, text: "x", cfg: &TTSConfig{}},
		{name: "zero voice id", voiceID: 0, text: "你好", wantField: "voice_id"},
		{name: "negative voice id", voiceID: -1, text: "你好", wantField: "voice_id"},
		{name: "empty text", voiceID: 1, text: "", wantField: "text"},
		{name: "text too long", voiceID: 1, text: strings.Repeat("字", maxSynthesizeTextLength+1), wantField: "text"},
		{name: "text at limit", voiceID: 1, text: strings.Repeat("字", maxSynthesizeTextLength)},
		{name: "invalid emotion", voiceID: 1, text: "x", cfg: &TTSConfig{Emotion: "sad"}, wantField: "emotion"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateSynthesize(tt.voiceID, tt.text, tt.cfg)
			if tt.wantField == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			ve, ok := IsValidationError(err)
			if !ok {
				t.Fatalf("want ValidationError, got %T: %v", err, err)
			}
			if ve.Field != tt.wantField {
				t.Errorf("want field %q, got %q", tt.wantField, ve.Field)
			}
		})
	}
}

func TestBuildSynthesizeBody(t *testing.T) {
	t.Parallel()

	t.Run("nil cfg only text and voice_id", func(t *testing.T) {
		t.Parallel()
		got, err := buildSynthesizeBody(100001, "你好", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got["text"] != "你好" {
			t.Errorf("text = %v", got["text"])
		}
		if got["voice_id"].(int) != 100001 {
			t.Errorf("voice_id = %v", got["voice_id"])
		}
		if len(got) != 2 {
			t.Errorf("want exactly 2 fields, got %d: %v", len(got), got)
		}
	})

	t.Run("with cfg merges config fields", func(t *testing.T) {
		t.Parallel()
		cfg := &TTSConfig{MediaType: MediaMP3, Emotion: EmotionAngry}
		cfg.SetSpeed(0) // 显式零值必须写入
		got, err := buildSynthesizeBody(1, "x", cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got["media_type"] != MediaMP3 {
			t.Errorf("media_type = %v", got["media_type"])
		}
		if got["emotion"] != EmotionAngry {
			t.Errorf("emotion = %v", got["emotion"])
		}
		if _, ok := got["speed"]; !ok {
			t.Errorf("explicit zero speed should be present, got %v", got)
		}
		if got["text"] != "x" || got["voice_id"].(int) != 1 {
			t.Errorf("text/voice_id = %v / %v", got["text"], got["voice_id"])
		}
	})
}

func TestSynthesize_Success(t *testing.T) {
	t.Parallel()

	wantAudio := []byte("RIFF....fake-wav-bytes")
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, synthesizeEndpoint) {
			t.Errorf("path = %s", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("content-type = %s", ct)
		}
		if acc := r.Header.Get("Accept"); acc != "*/*" {
			t.Errorf("accept = %s, want */* (避免内容协商干扰二进制音频)", acc)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)

		w.Header().Set("Content-Type", "audio/wav")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(wantAudio)
	}))
	defer server.Close()

	client, err := New(WithAPIKey("k"), WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	audio, err := client.Synthesize(context.Background(), 100001, "你好，世界。", &TTSConfig{Speed: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(audio) != string(wantAudio) {
		t.Errorf("audio = %q, want %q", audio, wantAudio)
	}
	if gotBody["text"] != "你好，世界。" {
		t.Errorf("sent text = %v", gotBody["text"])
	}
	if gotBody["voice_id"].(float64) != 100001 {
		t.Errorf("sent voice_id = %v", gotBody["voice_id"])
	}
	if gotBody["speed"].(float64) != 7 {
		t.Errorf("sent speed = %v", gotBody["speed"])
	}
}

func TestSynthesize_APIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 百度错误体也可能携带 200，靠 Content-Type 判定
		_, _ = w.Write([]byte(`{"status":3300,"message":"input invalid"}`))
	}))
	defer server.Close()

	client, err := New(WithAPIKey("k"), WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Synthesize(context.Background(), 100001, "你好", nil)
	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatalf("want APIError, got %T: %v", err, err)
	}
	if apiErr.Code != 3300 || apiErr.Message != "input invalid" {
		t.Errorf("got code=%d message=%q", apiErr.Code, apiErr.Message)
	}
}

func TestSynthesize_UnexpectedResponse(t *testing.T) {
	t.Parallel()

	// 非音频且非 {status,message} 的响应，应返回普通 error 而非 APIError。
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<html>oops</html>"))
	}))
	defer server.Close()

	client, err := New(WithAPIKey("k"), WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Synthesize(context.Background(), 1, "你好", nil)
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if _, ok := IsAPIError(err); ok {
		t.Errorf("want plain error, got APIError: %v", err)
	}
}

func TestSynthesize_WithAccessToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, tokenEndpoint) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok-123","expires_in":2592000}`))
			return
		}
		if got := r.URL.Query().Get("access_token"); got != "tok-123" {
			t.Errorf("access_token = %q", got)
		}
		w.Header().Set("Content-Type", "audio/mp3")
		_, _ = w.Write([]byte("mp3-bytes"))
	}))
	defer server.Close()

	client, err := New(WithClientCredentials("id", "secret"), WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	audio, err := client.Synthesize(context.Background(), 1, "你好", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(audio) != "mp3-bytes" {
		t.Errorf("audio = %q", audio)
	}
}

func TestSynthesize_ValidationShortCircuit(t *testing.T) {
	t.Parallel()

	// 参数非法时不应发起任何 HTTP 请求。
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called on invalid input")
	}))
	defer server.Close()

	client, err := New(WithAPIKey("k"), WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.Synthesize(context.Background(), 0, "你好", nil); err == nil {
		t.Fatal("want validation error, got nil")
	}
}
