package edgetts

import (
	"io"
)

type EdgeTTS struct {
	communicator *communicate
	texts        []*communicateTextTask
	outCome      io.WriteCloser
}

func NewTTS(options ...Option) *EdgeTTS {
	args := &ttsOption{}
	for _, opt := range options {
		opt(args)
	}

	tts := newCommunicate().withVoice(args.Voice).withRate(args.Rate).withVolume(args.Volume)
	tts.openWss()
	return &EdgeTTS{
		communicator: tts,
		outCome:      args.wc,
		texts:        []*communicateTextTask{},
	}
}

func (eTTS *EdgeTTS) task(text string, voice string, rate string, volume string) *communicateTextTask {
	return &communicateTextTask{
		text: text,
		option: communicateTextOption{
			voice:  voice,
			rate:   rate,
			volume: volume,
		},
	}
}

func (eTTS *EdgeTTS) AddTextDefault(text string) *EdgeTTS {
	eTTS.texts = append(eTTS.texts, eTTS.task(text, "", "", ""))
	return eTTS
}

func (eTTS *EdgeTTS) AddTextWithVoice(text string, voice string) *EdgeTTS {
	eTTS.texts = append(eTTS.texts, eTTS.task(text, voice, "", ""))
	return eTTS
}

func (eTTS *EdgeTTS) AddText(text string, voice string, rate string, volume string) *EdgeTTS {
	eTTS.texts = append(eTTS.texts, eTTS.task(text, voice, rate, volume))
	return eTTS
}

func (eTTS *EdgeTTS) Speak() {
	defer eTTS.communicator.close()
	defer eTTS.outCome.Close()

	go eTTS.communicator.allocateTask(eTTS.texts)
	eTTS.communicator.createPool()
	for _, text := range eTTS.texts {
		eTTS.outCome.Write(text.speechData)
	}
}
