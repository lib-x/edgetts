package edgetts

const (
	TrustedClientToken = "6A5AA1D4EAFF4E9FB37E23D68491D6F4"
	EdgeWssEndpoint    = "wss://speech.platform.bing.com/consumer/speech/synthesize/readaloud/edge/v1?TrustedClientToken=" + TrustedClientToken
	VoiceListEndpoint  = "https://speech.platform.bing.com/consumer/speech/synthesize/readaloud/voices/list?trustedclienttoken=" + TrustedClientToken
)

// Locale
const (
	ZhCN = "zh-CN"
	EnUS = "en-US"
)

const (
	ChunkTypeAudio        = "Audio"
	ChunkTypeWordBoundary = "WordBoundary"
	ChunkTypeSessionEnd   = "SessionEnd"
	ChunkTypeEnd          = "ChunkEnd"
)
