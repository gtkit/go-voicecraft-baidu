package voicecraftbaidu

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newCloneServer 构造一个仅响应指定 endpoint 的 httptest 服务，返回固定 JSON。
// captureBody 非 nil 时会写入收到的请求体。
func newCloneServer(t *testing.T, wantPath, respJSON string, captureBody *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, wantPath) {
			t.Errorf("path = %s, want suffix %s", r.URL.Path, wantPath)
		}
		if captureBody != nil {
			raw, _ := io.ReadAll(r.Body)
			*captureBody = string(raw)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(respJSON))
	}))
}

func TestAPIKeyAuthorizationHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
		want string
	}{
		{name: "raw key gets Bearer prefix", key: "ALTAK-abc", want: "Bearer ALTAK-abc"},
		{name: "key with prefix not doubled", key: "Bearer ALTAK-abc", want: "Bearer ALTAK-abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotAuth string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotAuth = r.Header.Get("Authorization")
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"status":0,"data":{"text_id":"t","text":"x"}}`))
			}))
			defer server.Close()

			client, err := New(WithAPIKey(tt.key), WithBaseURL(server.URL))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := client.GetCloneText(context.Background()); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotAuth != tt.want {
				t.Errorf("Authorization = %q, want %q", gotAuth, tt.want)
			}
		})
	}
}

func TestGetCloneText_Success(t *testing.T) {
	t.Parallel()

	server := newCloneServer(t, cloneTextEndpoint,
		`{"status":0,"message":"","data":{"text_id":"tid-123","text":"请朗读这段文本"}}`, nil)
	defer server.Close()

	client, err := New(WithAPIKey("k"), WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.GetCloneText(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data == nil || resp.Data.TextID != "tid-123" || resp.Data.Text != "请朗读这段文本" {
		t.Errorf("data = %+v", resp.Data)
	}
}

func TestListVoices_Success(t *testing.T) {
	t.Parallel()

	var gotBody string
	server := newCloneServer(t, listVoicesEndpoint,
		`{"status":0,"message":"","data":{"total":2,"page":1,"page_size":10,"items":[{"voice_id":100001,"voice_name":"v1","lang":"zh","create_time":1700000000,"status":0}]}}`,
		&gotBody)
	defer server.Close()

	client, err := New(WithAPIKey("k"), WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.ListVoices(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data == nil || resp.Data.Total != 2 || len(resp.Data.Items) != 1 {
		t.Fatalf("data = %+v", resp.Data)
	}
	if resp.Data.Items[0].VoiceID != 100001 {
		t.Errorf("item voice_id = %d", resp.Data.Items[0].VoiceID)
	}
	if resp.Data.Items[0].CreateTime != 1700000000 {
		t.Errorf("item create_time = %d", resp.Data.Items[0].CreateTime)
	}
	if !strings.Contains(gotBody, `"page":1`) {
		t.Errorf("request body = %s, want page=1", gotBody)
	}
}

func TestListVoices_OmitNonPositivePage(t *testing.T) {
	t.Parallel()

	var gotBody string
	server := newCloneServer(t, listVoicesEndpoint,
		`{"status":0,"data":{"total":0,"page":1,"page_size":10,"items":[]}}`, &gotBody)
	defer server.Close()

	client, err := New(WithAPIKey("k"), WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.ListVoices(context.Background(), 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(gotBody, "page") {
		t.Errorf("request body = %s, page should be omitted when ≤ 0", gotBody)
	}
}

func TestVoiceDetail_Success(t *testing.T) {
	t.Parallel()

	var gotBody string
	server := newCloneServer(t, voiceDetailEndpoint,
		`{"status":0,"message":"","data":{"voice_id":100001,"voice_name":"v1","voice_desc":"d","lang":"zh","status":0,"create_time":1700000000}}`,
		&gotBody)
	defer server.Close()

	client, err := New(WithAPIKey("k"), WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.VoiceDetail(context.Background(), 100001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data == nil || resp.Data.VoiceID != 100001 || resp.Data.Lang != "zh" {
		t.Errorf("data = %+v", resp.Data)
	}
	if !strings.Contains(gotBody, `"voice_id":100001`) {
		t.Errorf("request body = %s", gotBody)
	}
}

func TestDeleteVoice_Success(t *testing.T) {
	t.Parallel()

	server := newCloneServer(t, deleteVoiceEndpoint, `{"status":0,"message":""}`, nil)
	defer server.Close()

	client, err := New(WithAPIKey("k"), WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	if err := client.DeleteVoice(context.Background(), 100001); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloneManage_ValidationShortCircuit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called on invalid input")
	}))
	defer server.Close()

	client, err := New(WithAPIKey("k"), WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.VoiceDetail(context.Background(), 0); err == nil {
		t.Error("VoiceDetail: want validation error, got nil")
	} else if _, ok := IsValidationError(err); !ok {
		t.Errorf("VoiceDetail: want ValidationError, got %T", err)
	}

	if err := client.DeleteVoice(context.Background(), -1); err == nil {
		t.Error("DeleteVoice: want validation error, got nil")
	} else if _, ok := IsValidationError(err); !ok {
		t.Errorf("DeleteVoice: want ValidationError, got %T", err)
	}
}

func TestCloneManage_APIError(t *testing.T) {
	t.Parallel()

	// status != 0 应转为 *APIError，Code 即返回码枚举值。
	server := newCloneServer(t, voiceDetailEndpoint,
		`{"status":11011,"message":"voice_id not exists"}`, nil)
	defer server.Close()

	client, err := New(WithAPIKey("k"), WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.VoiceDetail(context.Background(), 999)
	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatalf("want APIError, got %T: %v", err, err)
	}
	if apiErr.Code != CodeVoiceIDNotFound {
		t.Errorf("code = %d, want %d", apiErr.Code, CodeVoiceIDNotFound)
	}
}

func TestCloneManage_HTTPError(t *testing.T) {
	t.Parallel()

	t.Run("non-200 with business status maps to APIError", func(t *testing.T) {
		t.Parallel()
		// 百度业务错误以 status 字段表达；即便随非 200 返回，httpc 仍会解码 body，
		// postClone 据此映射为携带 HTTP 状态码的 APIError。
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"status":11006,"message":"no permission"}`))
		}))
		defer server.Close()

		client, _ := New(WithAPIKey("k"), WithBaseURL(server.URL))
		_, err := client.GetCloneText(context.Background())
		apiErr, ok := IsAPIError(err)
		if !ok {
			t.Fatalf("want APIError, got %T: %v", err, err)
		}
		if apiErr.StatusCode != http.StatusForbidden || apiErr.Code != CodeVoiceIDInvalid {
			t.Errorf("got http=%d code=%d", apiErr.StatusCode, apiErr.Code)
		}
	})

	t.Run("non-200 without code is plain error", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("oops"))
		}))
		defer server.Close()

		client, _ := New(WithAPIKey("k"), WithBaseURL(server.URL))
		_, err := client.ListVoices(context.Background(), 1)
		if err == nil {
			t.Fatal("want error, got nil")
		}
		if _, ok := IsAPIError(err); ok {
			t.Errorf("want plain error, got APIError: %v", err)
		}
	})

	t.Run("invalid json body returns decode error", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("not-json"))
		}))
		defer server.Close()

		client, _ := New(WithAPIKey("k"), WithBaseURL(server.URL))
		_, err := client.VoiceDetail(context.Background(), 1)
		if err == nil {
			t.Fatal("want decode error, got nil")
		}
		if _, ok := IsAPIError(err); ok {
			t.Errorf("want plain error, got APIError: %v", err)
		}
	})
}
