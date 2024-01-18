package edgetts

import (
	"os"
	"testing"
)

func TestSpeech_GenerateTTS(t *testing.T) {
	c, err := NewCommunicate("种一棵树最好的时间是十年前，其次是现在")
	if err != nil {
		t.Fatal(err)
	}
	audio, err := os.OpenFile("testdata/xxx.mp3", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Fatal(err)
	}
	speech, err := NewSpeech(c, audio)
	speech.GenerateTTS()
}
