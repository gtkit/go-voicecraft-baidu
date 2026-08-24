// package voicraft-baidu 提供百度智能云大模型声音复刻 API 的 Go SDK。
//
// 支持以下核心能力：
//   - 创建音色（REST API）：上传音频创建自定义音色
//   - 音色管理（REST API）：获取训练文本、查询音色列表与详情、删除音色
//   - 流式语音合成（WebSocket TTS）：基于已创建音色进行实时语音合成
//   - 非流式在线合成（REST API）：基于已创建音色，将文本一次性合成为完整音频
//   - 流式文本在线合成（WebSocket TTS）：流式文本在线合成基于 websocket 协议，可以将输入的文本合成为二进制格式的语音数据
//
// 接口返回码统一枚举于 codes.go，可通过 CodeDescription 获取中文说明。
//
// 鉴权方式支持 access_token（OAuth）和 API Key 两种模式。
package voicraftbaidu
