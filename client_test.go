package edgetts

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilterVoices(t *testing.T) {
	voices := []Voice{
		{ShortName: "zh-CN-XiaoxiaoNeural", Locale: "zh-CN", Gender: "Female"},
		{ShortName: "en-US-GuyNeural", Locale: "en-US", Gender: "Male"},
	}

	got := FilterVoices(voices, VoiceFilter{Locale: "zh-CN", Gender: "female"})
	if len(got) != 1 || got[0].ShortName != "zh-CN-XiaoxiaoNeural" {
		t.Fatalf("unexpected filter result: %+v", got)
	}
}

func TestEmptyInput(t *testing.T) {
	client := New()
	if _, err := client.Bytes(context.Background(), ""); !errors.Is(err, ErrEmptyInput) {
		t.Fatalf("expected ErrEmptyInput, got %v", err)
	}
	if _, err := client.Stream(context.Background(), "   "); !errors.Is(err, ErrEmptyInput) {
		t.Fatalf("expected ErrEmptyInput from stream, got %v", err)
	}
}

func TestBatchEmpty(t *testing.T) {
	client := New()
	if _, err := client.Batch(context.Background(), nil); !errors.Is(err, ErrBatchEmpty) {
		t.Fatalf("expected ErrBatchEmpty, got %v", err)
	}
	if err := client.WriteZIP(context.Background(), io.Discard, nil, nil); !errors.Is(err, ErrBatchEmpty) {
		t.Fatalf("expected ErrBatchEmpty from zip, got %v", err)
	}
}

func TestWriteJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	if err := writeJSON(buf, map[string]any{"name": "demo", "count": 2}); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, `"name":"demo"`) {
		t.Fatalf("unexpected json output: %s", output)
	}
}

func TestSaveBatchCreatesFiles(t *testing.T) {
	dir := t.TempDir()
	client := New()
	items := []BatchItem{{Name: "a.mp3", Request: Text("")}}
	results, err := client.SaveBatch(context.Background(), dir, items)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("unexpected results len: %d", len(results))
	}
	if results[0].Err == nil {
		t.Fatalf("expected item error for empty input")
	}
	if _, statErr := os.Stat(filepath.Join(dir, "a.mp3")); !os.IsNotExist(statErr) {
		t.Fatalf("expected file not to exist, got %v", statErr)
	}
}

func TestWriteZIPWritesMetadata(t *testing.T) {
	buf := &bytes.Buffer{}
	client := New()
	err := client.WriteZIP(context.Background(), buf, []BatchItem{{Name: "a.mp3", Request: Text("")}}, map[string]any{"k": "v"})
	if err == nil {
		t.Fatal("expected error for empty request in zip")
	}

	buf.Reset()
	zw := zip.NewWriter(buf)
	w, err := zw.Create("x")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("ok")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}
