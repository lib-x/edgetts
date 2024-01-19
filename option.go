package edgetts

import "github.com/lib-x/edgetts/internal/communicateOption"

type option struct {
	Voice            string
	VoiceLangRegion  string
	Rate             string
	Volume           string
	HttpProxy        string
	Socket5Proxy     string
	Socket5ProxyUser string
	Socket5ProxyPass string
}

func (o *option) toInternalOption() *communicateOption.CommunicateOption {
	return &communicateOption.CommunicateOption{
		Voice:            o.Voice,
		VoiceLangRegion:  o.VoiceLangRegion,
		Rate:             o.Rate,
		Volume:           o.Volume,
		HttpProxy:        o.HttpProxy,
		Socket5Proxy:     o.Socket5Proxy,
		Socket5ProxyUser: o.Socket5ProxyUser,
		Socket5ProxyPass: o.Socket5ProxyPass,
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

func WithRate(rate string) Option {
	return func(option *option) {
		option.Rate = rate
	}
}

func WithVolume(volume string) Option {
	return func(option *option) {
		option.Volume = volume
	}
}

func WithHttpProxy(proxy string) Option {
	return func(option *option) {
		option.HttpProxy = proxy
	}
}

func WithSocket5Proxy(proxy, userName, password string) Option {
	return func(option *option) {
		option.Socket5Proxy = proxy
		option.Socket5ProxyUser = userName
		option.Socket5ProxyPass = password
	}
}
