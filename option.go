package edgetts

import "github.com/lib-x/edgetts/internal/communicateOption"

type option struct {
	Voice                 string
	VoiceLangRegion       string
	Pitch                 string
	Rate                  string
	Volume                string
	HttpProxy             string
	Socket5Proxy          string
	Socket5ProxyUser      string
	Socket5ProxyPass      string
	IgnoreSSLVerification bool
}

func (o *option) toInternalOption() *communicateOption.CommunicateOption {
	return &communicateOption.CommunicateOption{
		Voice:            o.Voice,
		VoiceLangRegion:  o.VoiceLangRegion,
		Pitch:            o.Pitch,
		Rate:             o.Rate,
		Volume:           o.Volume,
		HttpProxy:        o.HttpProxy,
		Socket5Proxy:     o.Socket5Proxy,
		Socket5ProxyUser: o.Socket5ProxyUser,
		Socket5ProxyPass: o.Socket5ProxyPass,
		IgnoreSSL:        o.IgnoreSSLVerification,
	}
}

type Option func(option *option)

func WithVoice(voice string) Option {
	return func(option *option) {
		option.Voice = voice
	}
}

func WithVoiceLangRegion(voiceLangRegion string) Option {
	return func(option *option) {
		option.VoiceLangRegion = voiceLangRegion
	}

}

// WithPitch set pitch of the tts output.such as +50Hz,-50Hz
func WithPitch(pitch string) Option {
	return func(option *option) {
		option.Pitch = pitch
	}
}

// WithRate set rate of the tts output.rate=-50% means rate down 50%,rate=+50% means rate up 50%
func WithRate(rate string) Option {
	return func(option *option) {
		option.Rate = rate
	}
}

// WithVolume set volume of the tts output.volume=-50% means volume down 50%,volume=+50% means volume up 50%
func WithVolume(volume string) Option {
	return func(option *option) {
		option.Volume = volume
	}
}

func WithHttpProxy(proxy string) Option {
	return WithHttpProxyEx(proxy, false)
}

func WithHttpProxyEx(proxy string, ignoreSSLVerification bool) Option {
	return func(option *option) {
		option.HttpProxy = proxy
		option.IgnoreSSLVerification = ignoreSSLVerification
	}
}

func WithSocket5Proxy(proxy, userName, password string) Option {
	return WithSocket5ProxyEx(proxy, userName, password, false)
}

func WithSocket5ProxyEx(proxy, userName, password string, ignoreSSLVerification bool) Option {
	return func(option *option) {
		option.Socket5Proxy = proxy
		option.Socket5ProxyUser = userName
		option.Socket5ProxyPass = password
		option.IgnoreSSLVerification = ignoreSSLVerification
	}

}
