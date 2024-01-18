package edgetts

type MessageType int

const (
	trustedClientToken = "6A5AA1D4EAFF4E9FB37E23D68491D6F4"
	edgeWssEndpoint    = "wss://speech.platform.bing.com/consumer/speech/synthesize/" + "readaloud/edge/v1?trustedClientToken=" + trustedClientToken

	defaultVoice = "zh-CN-XiaoxiaoNeural"
)
