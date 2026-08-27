package voicecraftbaidu

import (
	"testing"

	"github.com/gtkit/json/v2"
)

// BenchmarkTTSConfig_MarshalJSON 基准测试声音复刻合成参数的自定义 JSON 编码。
// 该路径在每次 WebSocket 初始化帧和非流式合成请求中都会执行。
func BenchmarkTTSConfig_MarshalJSON(b *testing.B) {
	cfg := (&TTSConfig{
		Lang:      LangChinese,
		Dialect:   DialectSichuan,
		MediaType: MediaMP3,
		Emotion:   EmotionHappy,
	}).SetPitch(0).SetVolume(8).SetSpeed(7)

	b.ReportAllocs()
	for b.Loop() {
		if _, err := json.Marshal(cfg); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStreamTTSConfig_MarshalJSON 基准测试公有云流式合成参数的自定义 JSON 编码。
func BenchmarkStreamTTSConfig_MarshalJSON(b *testing.B) {
	cfg := (&StreamTTSConfig{Aue: AudioEncodingMP3}).SetSpd(7).SetPit(0).SetVol(5).WithSampleRate16K()

	b.ReportAllocs()
	for b.Loop() {
		if _, err := json.Marshal(cfg); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBuildSynthesizeBody 基准测试非流式合成请求体的构造（含配置合并）。
func BenchmarkBuildSynthesizeBody(b *testing.B) {
	cfg := &TTSConfig{MediaType: MediaMP3, Emotion: EmotionHappy, Speed: 7}
	const text = "当春风拂过，大地渐渐回暖，万物复苏的季节到来了。"

	b.ReportAllocs()
	for b.Loop() {
		if _, err := buildSynthesizeBody(100001, text, cfg); err != nil {
			b.Fatal(err)
		}
	}
}
