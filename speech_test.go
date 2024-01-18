package edgetts

import (
	"archive/zip"
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

	audio2, err := os.OpenFile("testdata/test1.mp3", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatal(err)
	}
	err = speech.AddTask("莫听穿林打叶声，何妨吟啸且徐行。竹杖芒鞋轻胜马，谁怕？一蓑烟雨任平生。料峭春风吹酒醒，微冷，山头斜照却相迎。回首向来萧瑟处，归去，也无风雨也无晴。", audio2)
	if err != nil {
		t.Fatal(err)
	}

	speech.StartTasks()
}

func TestSpeech_StartTasksToZip(t *testing.T) {
	speech, err := NewSpeech()
	if err != nil {
		t.Fatal(err)
	}
	w, err := os.OpenFile("testdata/tts.zip", os.O_RDWR|os.O_CREATE, 0666)
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()
	zipWriter.SetComment("This is a test")
	a, err := zipWriter.Create("a.mp3")
	if err != nil {
		t.Fatal(err)
	}
	err = speech.AddTask("种一棵树最好的时间是十年前，其次是现在.The best time to plant a tree is 20 years ago. The second-best time is now.", a)
	if err != nil {
		t.Fatal(err)
	}

	b, err := zipWriter.Create("b.mp3")
	if err != nil {
		t.Fatal(err)
	}

	err = speech.AddTask("莫听穿林打叶声，何妨吟啸且徐行。竹杖芒鞋轻胜马，谁怕？一蓑烟雨任平生。料峭春风吹酒醒，微冷，山头斜照却相迎。回首向来萧瑟处，归去，也无风雨也无晴。", b)
	if err != nil {
		t.Fatal(err)
	}
	speech.StartTasks()
	zipWriter.Flush()

}
