package edgetts

import (
	"context"
	"io"
)

// Speech is a compatibility wrapper around Client.
//
// Deprecated: prefer Client and the package-level helper functions.
type Speech struct {
	client *Client
	items  []speechTask
}

type speechTask interface {
	run(context.Context, *Client) error
}

type singleSpeechTask struct {
	request Request
	output  io.Writer
}

func (t singleSpeechTask) run(ctx context.Context, client *Client) error {
	_, err := client.WriteRequestTo(ctx, t.request, t.output)
	return err
}

type packSpeechTask struct {
	items []BatchItem
	write func(context.Context, *Client) error
}

func (t packSpeechTask) run(ctx context.Context, client *Client) error {
	return t.write(ctx, client)
}

// GetVoiceList retrieves the list of voices available for the speech.
//
// Deprecated: prefer Client.Voices.
func (s *Speech) GetVoiceList() ([]Voice, error) {
	return s.client.Voices(context.Background())
}

// AddSingleTask adds a single task to the speech.
//
// Deprecated: prefer Client.WriteTo, Client.Save, or Client.Do.
func (s *Speech) AddSingleTask(text string, output io.Writer) error {
	s.items = append(s.items, singleSpeechTask{request: Text(text), output: output})
	return nil
}

// AddPackTask adds a pack task to the speech.
//
// Deprecated: prefer Client.SaveBatch or Client.WriteZIP.
func (s *Speech) AddPackTask(dataEntries map[string]string, entryCreator func(name string) (io.Writer, error), output io.Writer, metaData ...map[string]any) error {
	return s.AddPackTaskWithCustomOptions(dataEntries, nil, entryCreator, output, metaData...)
}

// AddPackTaskWithCustomOptions adds a pack task with custom options.
//
// Deprecated: prefer Client.SaveBatch or Client.WriteZIP.
func (s *Speech) AddPackTaskWithCustomOptions(dataEntries map[string]string, entriesOption map[string][]Option, entryCreator func(name string) (io.Writer, error), output io.Writer, metaData ...map[string]any) error {
	if len(dataEntries) == 0 {
		return ErrBatchEmpty
	}

	items := make([]BatchItem, 0, len(dataEntries))
	for name, text := range dataEntries {
		items = append(items, BatchItem{Name: name, Request: Text(text, entriesOption[name]...)})
	}

	task := packSpeechTask{items: items, write: func(ctx context.Context, client *Client) error {
		for _, item := range items {
			writer, err := entryCreator(item.Name)
			if err != nil {
				return err
			}
			if _, err := client.WriteRequestTo(ctx, item.Request, writer); err != nil {
				return err
			}
		}
		for _, meta := range metaData {
			writer, err := entryCreator("metadata.json")
			if err != nil {
				return err
			}
			if err := writeJSON(writer, meta); err != nil {
				return err
			}
		}
		return nil
	}}

	s.items = append(s.items, task)
	return nil
}

// StartTasks starts all added tasks.
//
// Deprecated: prefer explicit Client calls.
func (s *Speech) StartTasks() error {
	for _, item := range s.items {
		if err := item.run(context.Background(), s.client); err != nil {
			return err
		}
	}
	return nil
}
