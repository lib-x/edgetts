package edgetts

import "io"

type Option func(option *ttsOption)
type ttsOption struct {
	Text   string
	Voice  string
	Proxy  string
	Rate   string
	Volume string
	//WordsInCue     float64
	wc io.WriteCloser
	//WriteSubtitles string
}

func WithText(text string) Option {
	return func(option *ttsOption) {
		option.Text = text
	}
}

func WithVoice(voice string) Option {
	return func(option *ttsOption) {
		option.Voice = voice
	}
}

func WithProxy(proxy string) Option {
	return func(option *ttsOption) {
		option.Proxy = proxy
	}
}

func WithRate(rate string) Option {
	return func(option *ttsOption) {
		option.Rate = rate
	}
}

func WithVolume(volume string) Option {
	return func(option *ttsOption) {
		option.Volume = volume
	}
}

func WithWriterCloser(writer io.WriteCloser) Option {
	return func(option *ttsOption) {
		option.wc = writer
	}

}
