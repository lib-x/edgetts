package edgetts

import (
	"os"
	"testing"
)

func TestSpeech_StartTasks(t *testing.T) {
	speech, err := NewSpeech()
	if err != nil {
		t.Fatal(err)
	}
	audio, err := os.OpenFile("testdata/test.mp3", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatal(err)
	}
	err = speech.AddTask("种一棵树最好的时间是十年前，其次是现在.The best time to plant a tree is 20 years ago. The second-best time is now.", audio)
	if err != nil {
		t.Fatal(err)
	}

	speech.StartTasks()
}
