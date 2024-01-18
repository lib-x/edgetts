package communicateOption

import "github.com/lib-x/edgetts/internal/businessConsts"

type CommunicateOption struct {
	Voice            string
	VoiceLangRegion  string
	Rate             string
	Volume           string
	HttpProxy        string
	Socket5Proxy     string
	Socket5ProxyUser string
	Socket5ProxyPass string
}

func (c *CommunicateOption) ApplyDefaultOption() {
	// Default values
	if c.Voice == "" {
		c.Voice = businessConsts.DefaultVoice
		c.VoiceLangRegion = businessConsts.DefaultVoice
	}
	if c.Rate == "" {
		c.Rate = "+0%"
	}
	if c.Volume == "" {
		c.Volume = "+0%"
	}

}

type Option func(*CommunicateOption)

func WithVoice(voice string) Option {
	return func(option *CommunicateOption) {
		option.Voice = voice
	}
}

func WithVoiceLangRegion(voiceLangRegion string) Option {
	return func(option *CommunicateOption) {
		option.VoiceLangRegion = voiceLangRegion
	}

}

func WithRate(rate string) Option {
	return func(option *CommunicateOption) {
		option.Rate = rate
	}
}

func WithVolume(volume string) Option {
	return func(option *CommunicateOption) {
		option.Volume = volume
	}
}

func WithHttpProxy(proxy string) Option {
	return func(option *CommunicateOption) {
		option.HttpProxy = proxy
	}
}

func WithSocket5Proxy(proxy, userName, password string) Option {
	return func(option *CommunicateOption) {
		option.Socket5Proxy = proxy
		option.Socket5ProxyUser = userName
		option.Socket5ProxyPass = password
	}
}
