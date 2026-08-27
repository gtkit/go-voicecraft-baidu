package voicecraftbaidu

// ============================================================================
// 返回码枚举
//
// 这些常量对应百度大模型声音复刻接口返回的错误码：
//   - 音色管理 / 创建接口在响应 JSON 的 status 字段返回（见 APIError.Code）
//   - 在线合成接口在握手、初始化、文本阶段以 code 字段返回（见 WebSocketError.Code）
//
// 调用方可直接与这些常量比较，例如：
//
//	if apiErr, ok := voicecraftbaidu.IsAPIError(err); ok && apiErr.Code == voicecraftbaidu.CodeVoiceIDNotFound {
//	    // 音色不存在
//	}
//
// 通过 CodeDescription 可获取某个码的中文说明。
// ============================================================================

// 业务码：音色管理与创建接口在 status 字段返回。
const (
	CodeSuccess              = 0     // 处理成功
	CodeUserConcurrencyLimit = 10014 // 用户并发超限
	CodeUserQuotaExceeded    = 10015 // 用户配额超限
	CodeServiceTemporaryErr  = 10020 // 服务临时错误，请稍候再试
	CodeVoiceTextMismatch    = 10021 // 音频与文本匹配校验失败
	CodeTextIDExpired        = 10022 // text_id 不存在或已超过 24 小时
	CodeAudioDownloadLimit   = 10023 // 音频 URL 不可访问或上传文件超过 5M
	CodeVoiceNotFound        = 10025 // 音色不存在或已被删除
	CodeInvalidAudioFormat   = 10026 // 文件内容无效
	CodeAudioSensitivePerson = 10027 // 音频可能涉及敏感人物
	CodeSynthesisException   = 11000 // 合成异常，请检查音色与合成语种是否一致
	CodeQuotaThrottle        = 11002 // 限流：配额超限
	CodeConcurrencyThrottle  = 11003 // 限流：并发超限
	CodeTextTooLong          = 11004 // 文本超长
	CodeVoiceIDInvalid       = 11006 // voice_id 错误（无访问权限）
	CodeMissingAuth          = 11007 // 未传递有效鉴权信息
	CodeServiceRetry         = 11008 // 服务临时异常，请稍候重试
	CodeMissingParameter     = 11009 // 参数缺失
	CodeInvalidParameter     = 11010 // 参数无效
	CodeVoiceIDNotFound      = 11011 // voice_id 不存在
	CodeInvalidPage          = 11012 // page 参数无效，须 ≥ 1
	CodeInvalidBase64Audio   = 11013 // 音频内容无效（base64 编码无效）
	CodeNoDataPermission     = 11014 // 无访问该数据的权限
	CodeTextSensitive        = 11015 // 文本包含敏感信息
	CodeFileDownloadFailed   = 12000 // 文件下载失败
	CodeAudioTooShort        = 12001 // 音频内容太短
	CodeWERCheckFailed       = 12002 // 未检测到有效音频（文本匹配校验失败）
	CodeAudioSNRFailed       = 12003 // 未检测到有效音频（信噪比校验失败）
	CodeRecognitionFailed    = 12004 // 无有效人声（识别失败）
	CodeAudioLevelFailed     = 12005 // 无有效人声（音量异常）
	CodeAudioSpeedFailed     = 12006 // 无有效人声（语速异常）
	CodeAudioQualityPoor     = 12007 // 音频质量较差
)

// 协议码：在线合成（流式 WebSocket / 非流式 HTTP）在 code 字段返回。
const (
	CodeNoAPIPermission          = 6      // appid 无相应接口权限
	CodeOpenAPIConcurrencyLimit  = 15     // 触发并发限流
	CodeOpenAPIUsageLimit        = 17     // 无剩余可用额度
	CodeAccessTokenInvalid       = 110    // access_token 校验不通过
	CodeAccessTokenExpired       = 111    // access_token 过期
	CodeIAMAuthError             = 217    // API Key 校验不通过
	CodeInvalidValue             = 216100 // 参数值非法
	CodeMissingRequiredParam     = 216101 // 缺少必填参数
	CodeTextOverLimit            = 216103 // 单次文本超过 1000 字
	CodePendingTextTooLong       = 216429 // 待处理文本过长（发送频率过快）
	CodeTTSVoiceNoPermission     = 216403 // 对该 voice_id 无权限
	CodeTTSVoiceNotFound         = 216404 // voice_id 不存在
	CodeOpenAPIUsageLimitReached = 216604 // 额度已用完
	CodeInternalError            = 282000 // 服务器内部错误
	CodeVoiceServiceNotStarted   = 282101 // voice_id 服务未启动
)

// codeDescriptions 维护已知返回码到中文说明的映射，由 CodeDescription 查询。
var codeDescriptions = map[int]string{
	CodeSuccess:              "处理成功",
	CodeUserConcurrencyLimit: "用户并发超限，如有高并发需求请提交工单",
	CodeUserQuotaExceeded:    "用户配额超限，如有高额度需求请提交工单",
	CodeServiceTemporaryErr:  "服务临时错误，请稍候再试",
	CodeVoiceTextMismatch:    "音频与文本匹配校验失败，请按照返回 text 朗读",
	CodeTextIDExpired:        "text_id 不存在或已超过 24 小时，请重新获取训练文本",
	CodeAudioDownloadLimit:   "音频 URL 不可访问或上传文件超过 5M",
	CodeVoiceNotFound:        "该音色不存在或已被删除",
	CodeInvalidAudioFormat:   "文件内容无效，请提供格式正确的音频",
	CodeAudioSensitivePerson: "音频可能涉及敏感人物，请重新上传或指定文本复刻",

	CodeSynthesisException:  "合成异常，请检查音色与合成语种是否一致",
	CodeQuotaThrottle:       "限流：用户配额超限",
	CodeConcurrencyThrottle: "限流：用户并发超限",
	CodeTextTooLong:         "文本超长，请缩短文本重试",
	CodeVoiceIDInvalid:      "voice_id 错误，请检查是否正确",
	CodeMissingAuth:         "未传递有效的鉴权信息",
	CodeServiceRetry:        "服务临时异常，请稍候重试",
	CodeMissingParameter:    "参数缺失，请检查输入参数",
	CodeInvalidParameter:    "参数无效，请检查输入参数",
	CodeVoiceIDNotFound:     "voice_id 不存在，请检查是否正确",
	CodeInvalidPage:         "page 参数无效，须大于等于 1",
	CodeInvalidBase64Audio:  "音频内容无效，请使用有效的 base64 编码",
	CodeNoDataPermission:    "无访问该数据的权限",
	CodeTextSensitive:       "文本包含敏感信息，请去掉后重试",

	CodeFileDownloadFailed: "文件下载失败，请检查音频下载链接",
	CodeAudioTooShort:      "音频内容太短，请更换音频",
	CodeWERCheckFailed:     "未检测到有效音频，请根据返回 text 朗读",
	CodeAudioSNRFailed:     "未检测到有效音频，请根据返回 text 朗读",
	CodeRecognitionFailed:  "无有效的人声，请更换音频",
	CodeAudioLevelFailed:   "无有效的人声，请调整音量后重新复刻",
	CodeAudioSpeedFailed:   "无有效的人声，请调整语速后重新复刻",
	CodeAudioQualityPoor:   "音频质量较差，请更换音频",

	CodeNoAPIPermission:          "appid 无相应接口权限",
	CodeOpenAPIConcurrencyLimit:  "触发并发限流",
	CodeOpenAPIUsageLimit:        "无剩余可用额度",
	CodeAccessTokenInvalid:       "access_token 校验不通过",
	CodeAccessTokenExpired:       "access_token 过期，请更新",
	CodeIAMAuthError:             "API Key 校验不通过",
	CodeInvalidValue:             "参数值非法",
	CodeMissingRequiredParam:     "缺少必填参数",
	CodeTextOverLimit:            "单次文本超过 1000 字",
	CodePendingTextTooLong:       "发送频率过快，待处理文本过长",
	CodeTTSVoiceNoPermission:     "对该 voice_id 无权限",
	CodeTTSVoiceNotFound:         "voice_id 不存在",
	CodeOpenAPIUsageLimitReached: "额度已用完",
	CodeInternalError:            "服务器内部错误",
	CodeVoiceServiceNotStarted:   "voice_id 服务未启动",
}

// CodeDescription 返回返回码的中文说明。
// 若 code 不在已知枚举内，known 为 false、desc 为空字符串。
//
// 用法：
//
//	if apiErr, ok := voicecraftbaidu.IsAPIError(err); ok {
//	    if desc, known := voicecraftbaidu.CodeDescription(apiErr.Code); known {
//	        log.Printf("接口返回 %d：%s", apiErr.Code, desc)
//	    }
//	}
func CodeDescription(code int) (desc string, known bool) {
	desc, known = codeDescriptions[code]
	return desc, known
}
