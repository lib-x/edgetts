package businessConsts

const (
	WindowsFileTimeEpoch = 11644473600
	ChromiumFullVersion  = "103.0.5060.66"
	TrustedClientToken   = "6A5AA1D4EAFF4E9FB37E23D68491D6F4"
	EdgeWssEndpoint      = "wss://speech.platform.bing.com/consumer/speech/synthesize/" + "readaloud/edge/v1?trustedClientToken=" + TrustedClientToken
	VoiceListEndpoint    = "https://speech.platform.bing.com/consumer/speech/synthesize/readaloud/voices/list?trustedclienttoken=" + TrustedClientToken
)

const (
	DefaultVoice = "zh-CN-XiaoxiaoNeural"
)
