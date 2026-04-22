package edgetts

import "github.com/lib-x/edgetts/internal/communicateOption"

type option struct {
	Voice                 string
	VoiceLangRegion       string
	Pitch                 string
	Rate                  string
	Volume                string
	HTTPProxy             string
	SOCKS5Proxy           string
	SOCKS5ProxyUser       string
	SOCKS5ProxyPass       string
	IgnoreSSLVerification bool
}

func (o *option) toInternalOption() *communicateOption.CommunicateOption {
	return &communicateOption.CommunicateOption{
		Voice:            o.Voice,
		VoiceLangRegion:  o.VoiceLangRegion,
		Pitch:            o.Pitch,
		Rate:             o.Rate,
		Volume:           o.Volume,
		HttpProxy:        o.HTTPProxy,
		Socket5Proxy:     o.SOCKS5Proxy,
		Socket5ProxyUser: o.SOCKS5ProxyUser,
		Socket5ProxyPass: o.SOCKS5ProxyPass,
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

// WithPitch sets pitch, e.g. +50Hz or -50Hz.
func WithPitch(pitch string) Option {
	return func(option *option) {
		option.Pitch = pitch
	}
}

// WithRate sets rate, e.g. -50% or +50%.
func WithRate(rate string) Option {
	return func(option *option) {
		option.Rate = rate
	}
}

// WithVolume sets volume, e.g. -50% or +50%.
func WithVolume(volume string) Option {
	return func(option *option) {
		option.Volume = volume
	}
}

func WithHttpProxy(proxy string) Option { return WithHTTPProxy(proxy) }

func WithHTTPProxy(proxy string) Option {
	return WithHTTPProxyEx(proxy, false)
}

func WithHttpProxyEx(proxy string, ignoreSSLVerification bool) Option {
	return WithHTTPProxyEx(proxy, ignoreSSLVerification)
}

func WithHTTPProxyEx(proxy string, ignoreSSLVerification bool) Option {
	return func(option *option) {
		option.HTTPProxy = proxy
		option.IgnoreSSLVerification = ignoreSSLVerification
	}
}

func WithSocket5Proxy(proxy, userName, password string) Option {
	return WithSOCKS5Proxy(proxy, userName, password)
}

func WithSOCKS5Proxy(proxy, userName, password string) Option {
	return WithSOCKS5ProxyEx(proxy, userName, password, false)
}

func WithSocket5ProxyEx(proxy, userName, password string, ignoreSSLVerification bool) Option {
	return WithSOCKS5ProxyEx(proxy, userName, password, ignoreSSLVerification)
}

func WithSOCKS5ProxyEx(proxy, userName, password string, ignoreSSLVerification bool) Option {
	return func(option *option) {
		option.SOCKS5Proxy = proxy
		option.SOCKS5ProxyUser = userName
		option.SOCKS5ProxyPass = password
		option.IgnoreSSLVerification = ignoreSSLVerification
	}
}

func WithInsecureSkipVerify() Option {
	return func(option *option) {
		option.IgnoreSSLVerification = true
	}
}
