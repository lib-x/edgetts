package edgetts

type communicateOption struct {
	Voice           string
	VoiceLangRegion string
	Rate            string
	Volume          string
	Proxy           string
}

func (c *communicateOption) applyDefaultOption() {
	// Default values
	if c.Voice == "" {
		c.Voice = defaultVoice
		c.VoiceLangRegion = defaultVoice
	}
	if c.Rate == "" {
		c.Rate = "+0%"
	}
	if c.Volume == "" {
		c.Volume = "+0%"
	}

}

type Option func(*communicateOption)

func WithVoice(voice string) Option {
	return func(option *communicateOption) {
		option.Voice = voice
	}
}

func WithVoiceLangRegion(voiceLangRegion string) Option {
	return func(option *communicateOption) {
		option.VoiceLangRegion = voiceLangRegion
	}

}

func WithRate(rate string) Option {
	return func(option *communicateOption) {
		option.Rate = rate
	}
}

func WithVolume(volume string) Option {
	return func(option *communicateOption) {
		option.Volume = volume
	}
}

func WithProxy(proxy string) Option {
	return func(option *communicateOption) {
		option.Proxy = proxy
	}
}
