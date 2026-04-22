package edgetts

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestNewSpeech(t *testing.T) {
	speech, err := NewSpeech(WithVoice("zh-CN-YunxiaNeural"))
	if err != nil {
		t.Fatal(err)
	}
	if speech == nil || speech.client == nil {
		t.Fatal("expected initialized speech client")
	}
}

func TestSpeechAddPackTaskEmpty(t *testing.T) {
	speech, _ := NewSpeech()
	if err := speech.AddPackTask(nil, func(name string) (io.Writer, error) { return io.Discard, nil }, io.Discard); !errors.Is(err, ErrBatchEmpty) {
		t.Fatalf("expected ErrBatchEmpty, got %v", err)
	}
}

func TestSpeechStartTasksSingleTaskEmptyInput(t *testing.T) {
	speech, _ := NewSpeech()
	if err := speech.AddSingleTask("", io.Discard); err != nil {
		t.Fatal(err)
	}
	if err := speech.StartTasks(); !errors.Is(err, ErrEmptyInput) {
		t.Fatalf("expected ErrEmptyInput, got %v", err)
	}
}

func TestSpeechAddPackTaskWithMetadataWritesMetadata(t *testing.T) {
	speech, _ := NewSpeech()
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	entries := map[string]string{"a.mp3": ""}
	created := map[string]bool{}
	err := speech.AddPackTask(entries, func(name string) (io.Writer, error) {
		created[name] = true
		return zw.Create(name)
	}, io.Discard, map[string]any{"name": "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if err := speech.StartTasks(); err == nil {
		t.Fatal("expected empty input error")
	}
}
