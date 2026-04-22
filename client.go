package edgetts

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lib-x/edgetts/internal/communicate"
)

// Client synthesizes text and SSML to audio.
type Client struct {
	options []Option
	vm      *VoiceManager
}

// New creates a reusable client.
func New(opts ...Option) *Client {
	return &Client{options: append([]Option(nil), opts...), vm: NewVoiceManager()}
}

// NewSpeech creates a compatibility wrapper around the new client-based API.
//
// Deprecated: prefer New and the package-level helper functions.
func NewSpeech(options ...Option) (*Speech, error) {
	return &Speech{client: New(options...)}, nil
}

// Do synthesizes one request and returns the audio bytes.
func (c *Client) Do(ctx context.Context, req Request) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := c.WriteRequestTo(ctx, req, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Bytes synthesizes text and returns the audio bytes.
func (c *Client) Bytes(ctx context.Context, text string, opts ...Option) ([]byte, error) {
	return c.Do(ctx, Text(text, opts...))
}

// BytesSSML synthesizes SSML and returns the audio bytes.
func (c *Client) BytesSSML(ctx context.Context, ssml string, opts ...Option) ([]byte, error) {
	return c.Do(ctx, SSML(ssml, opts...))
}

// WriteRequestTo writes synthesized audio to w.
func (c *Client) WriteRequestTo(ctx context.Context, req Request, w io.Writer) (int64, error) {
	if strings.TrimSpace(req.Input) == "" {
		return 0, ErrEmptyInput
	}

	comm, err := c.newCommunicate(req)
	if err != nil {
		return 0, err
	}
	return comm.WriteStreamToContext(ctx, w)
}

// WriteTo writes synthesized text audio to w.
func (c *Client) WriteTo(ctx context.Context, text string, w io.Writer, opts ...Option) (int64, error) {
	return c.WriteRequestTo(ctx, Text(text, opts...), w)
}

// WriteSSMLTo writes synthesized SSML audio to w.
func (c *Client) WriteSSMLTo(ctx context.Context, ssml string, w io.Writer, opts ...Option) (int64, error) {
	return c.WriteRequestTo(ctx, SSML(ssml, opts...), w)
}

// Save writes synthesized text audio to a file.
func (c *Client) Save(ctx context.Context, text, path string, opts ...Option) error {
	return c.saveRequest(ctx, Text(text, opts...), path)
}

// SaveSSML writes synthesized SSML audio to a file.
func (c *Client) SaveSSML(ctx context.Context, ssml, path string, opts ...Option) error {
	return c.saveRequest(ctx, SSML(ssml, opts...), path)
}

func (c *Client) saveRequest(ctx context.Context, req Request, path string) error {
	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create file %s: %w", tmpPath, err)
	}

	if _, err := c.WriteRequestTo(ctx, req, f); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close file %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename file %s to %s: %w", tmpPath, path, err)
	}
	return nil
}

// Stream synthesizes text and returns a streaming reader.
func (c *Client) Stream(ctx context.Context, text string, opts ...Option) (io.ReadCloser, error) {
	return c.streamRequest(ctx, Text(text, opts...))
}

// StreamSSML synthesizes SSML and returns a streaming reader.
func (c *Client) StreamSSML(ctx context.Context, ssml string, opts ...Option) (io.ReadCloser, error) {
	return c.streamRequest(ctx, SSML(ssml, opts...))
}

func (c *Client) streamRequest(ctx context.Context, req Request) (io.ReadCloser, error) {
	if strings.TrimSpace(req.Input) == "" {
		return nil, ErrEmptyInput
	}

	pr, pw := io.Pipe()
	go func() {
		_, err := c.WriteRequestTo(ctx, req, pw)
		_ = pw.CloseWithError(err)
	}()
	return pr, nil
}

// Batch synthesizes all items and returns structured per-item results.
func (c *Client) Batch(ctx context.Context, items []BatchItem) ([]BatchResult, error) {
	if len(items) == 0 {
		return nil, ErrBatchEmpty
	}

	results := make([]BatchResult, len(items))
	for i, item := range items {
		results[i].Name = item.Name
		data, err := c.Do(ctx, item.Request)
		results[i].Bytes = data
		results[i].Err = err
		results[i].N = int64(len(data))
	}
	return results, nil
}

// SaveBatch writes synthesized items into a directory.
func (c *Client) SaveBatch(ctx context.Context, dir string, items []BatchItem) ([]BatchResult, error) {
	if len(items) == 0 {
		return nil, ErrBatchEmpty
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create dir %s: %w", dir, err)
	}

	results := make([]BatchResult, len(items))
	for i, item := range items {
		results[i].Name = item.Name
		path := filepath.Join(dir, item.Name)
		if err := c.saveRequest(ctx, item.Request, path); err != nil {
			results[i].Err = err
			continue
		}
		info, statErr := os.Stat(path)
		if statErr == nil {
			results[i].N = info.Size()
		}
	}
	return results, nil
}

// WriteZIP writes a batch into a zip archive.
func (c *Client) WriteZIP(ctx context.Context, w io.Writer, items []BatchItem, meta map[string]any) error {
	if len(items) == 0 {
		return ErrBatchEmpty
	}

	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, item := range items {
		entryWriter, err := zw.Create(item.Name)
		if err != nil {
			return fmt.Errorf("create zip entry %s: %w", item.Name, err)
		}
		if _, err := c.WriteRequestTo(ctx, item.Request, entryWriter); err != nil {
			return fmt.Errorf("write zip entry %s: %w", item.Name, err)
		}
	}

	if meta != nil {
		entryWriter, err := zw.Create("metadata.json")
		if err != nil {
			return fmt.Errorf("create metadata entry: %w", err)
		}
		if err := writeJSON(entryWriter, meta); err != nil {
			return fmt.Errorf("write metadata: %w", err)
		}
	}

	return nil
}

// Voices lists available voices.
func (c *Client) Voices(ctx context.Context) ([]Voice, error) {
	return c.vm.ListVoicesContext(ctx)
}

// FindVoice finds the first matching voice.
func (c *Client) FindVoice(ctx context.Context, filter VoiceFilter) (Voice, error) {
	voices, err := c.Voices(ctx)
	if err != nil {
		return Voice{}, err
	}
	for _, voice := range FilterVoices(voices, filter) {
		return voice, nil
	}
	return Voice{}, ErrVoiceNotFound
}

func (c *Client) newCommunicate(req Request) (*communicate.Communicate, error) {
	merged := c.mergeOptions(req.Options...).toInternalOption()
	switch req.Type {
	case InputText:
		return communicate.NewCommunicate(communicate.InputText, req.Input, merged)
	case InputSSML:
		return communicate.NewCommunicate(communicate.InputSSML, req.Input, merged)
	default:
		return communicate.NewCommunicate(communicate.InputText, req.Input, merged)
	}
}

func (c *Client) mergeOptions(extra ...Option) *option {
	merged := &option{}
	for _, apply := range c.options {
		apply(merged)
	}
	for _, apply := range extra {
		apply(merged)
	}
	return merged
}

// FilterVoices filters the provided voice list.
func FilterVoices(voices []Voice, filter VoiceFilter) []Voice {
	result := make([]Voice, 0, len(voices))
	for _, voice := range voices {
		if filter.Name != "" && !strings.EqualFold(voice.Name, filter.Name) {
			continue
		}
		if filter.ShortName != "" && !strings.EqualFold(voice.ShortName, filter.ShortName) {
			continue
		}
		if filter.Locale != "" && !strings.EqualFold(voice.Locale, filter.Locale) {
			continue
		}
		if filter.Gender != "" && !strings.EqualFold(voice.Gender, filter.Gender) {
			continue
		}
		if filter.SuggestedCodec != "" && !strings.EqualFold(voice.SuggestedCodec, filter.SuggestedCodec) {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(voice.Status, filter.Status) {
			continue
		}
		if filter.Language != "" && !strings.EqualFold(voice.Language, filter.Language) {
			continue
		}
		result = append(result, voice)
	}
	return result
}

func writeJSON(w io.Writer, value any) error {
	return json.NewEncoder(w).Encode(value)
}

// Bytes synthesizes text with a temporary client.
func Bytes(ctx context.Context, text string, opts ...Option) ([]byte, error) {
	return New(opts...).Bytes(ctx, text)
}

// BytesSSML synthesizes SSML with a temporary client.
func BytesSSML(ctx context.Context, ssml string, opts ...Option) ([]byte, error) {
	return New(opts...).BytesSSML(ctx, ssml)
}

// Save synthesizes text to a file with a temporary client.
func Save(ctx context.Context, text, path string, opts ...Option) error {
	return New(opts...).Save(ctx, text, path)
}

// SaveSSML synthesizes SSML to a file with a temporary client.
func SaveSSML(ctx context.Context, ssml, path string, opts ...Option) error {
	return New(opts...).SaveSSML(ctx, ssml, path)
}

// WriteTo synthesizes text to an io.Writer with a temporary client.
func WriteTo(ctx context.Context, text string, w io.Writer, opts ...Option) (int64, error) {
	return New(opts...).WriteTo(ctx, text, w)
}

// WriteSSMLTo synthesizes SSML to an io.Writer with a temporary client.
func WriteSSMLTo(ctx context.Context, ssml string, w io.Writer, opts ...Option) (int64, error) {
	return New(opts...).WriteSSMLTo(ctx, ssml, w)
}

// Stream synthesizes text to a stream with a temporary client.
func Stream(ctx context.Context, text string, opts ...Option) (io.ReadCloser, error) {
	return New(opts...).Stream(ctx, text)
}

// StreamSSML synthesizes SSML to a stream with a temporary client.
func StreamSSML(ctx context.Context, ssml string, opts ...Option) (io.ReadCloser, error) {
	return New(opts...).StreamSSML(ctx, ssml)
}
