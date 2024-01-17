package edgetts

import (
	"os"
	"testing"
)

func Test_EdgeTTSSpeak(t *testing.T) {
	audio, err := os.OpenFile("testdata/test.mp3", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Error(err)
	}
	tts := NewTTS(WithWriterCloser(audio))
	tts.AddTextDefault("我爱go编程,you are my best friend")
	tts.Speak()
}
