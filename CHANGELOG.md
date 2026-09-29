# Changelog

本项目遵循 [Keep a Changelog 1.1.0](https://keepachangelog.com/zh-CN/1.1.0/) 与 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

## [Unreleased]

> ⚠ 破坏性变更（fail-closed）：`New` 开始校验 `WithBaseURL` 传入的基址，scheme 不是 http / https / ws / wss 或缺少主机名时返回 `*ValidationError`（`Field` 为 `base_url`）；WebSocket 握手失败的错误类型改为 `*HandshakeError`。
>
> 迁移：
> - 基址写成完整地址，如 `https://gw.example.com` 或 `http://127.0.0.1:3000/baidu`；不带 scheme 的 `gw.example.com` 需补上 scheme。
> - 基址中的 query、fragment、userinfo 在 `New` 中被丢弃；依赖它们的调用方改为在网关或 `WithHTTPClient` 的 Transport 上处理。
> - 按错误文本 `voicecraftbaidu: websocket dial:` 判断握手失败的代码，改用 `errors.AsType[*voicecraftbaidu.HandshakeError](err)`；底层错误仍可用 `errors.Is` 穿透。

### Added

- `HandshakeError`：WebSocket 握手失败时返回，`StatusCode` 为握手响应的 HTTP 状态码（无响应时为 0），`URL` 为不含查询串的目标地址，实现 `Unwrap`。
- `WithBaseURL` 接受 `ws` / `wss` 基址，分别按 `http` / `https` 处理。

### Changed

- `New` 规范化基址：去掉 query、fragment、userinfo 与末尾斜杠。

### Fixed

- WebSocket 地址跟随基址 scheme：`http` 基址走 `ws://`，`https` 基址走 `wss://`；此前固定为 `wss://`，经只开 HTTP 的网关时无法连接。
- WebSocket 地址保留基址的路径前缀：此前路径被百度路径覆盖，基址为 `https://gw.example.com/baidu` 时 REST 与 WebSocket 打到不同地址。
