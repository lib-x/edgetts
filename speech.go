package edgetts

import (
	"errors"
	"github.com/lib-x/edgetts/internal/communicate"
	"github.com/lib-x/edgetts/internal/communicateOption"
	"github.com/lib-x/edgetts/internal/ttsTask"
	"github.com/lib-x/edgetts/internal/voiceList"
	"io"
	"sync"
)

var (
	NoPackTaskEntries = errors.New("no pack task entries")
)

type Speech struct {
	options   []communicateOption.Option
	tasks     []*ttsTask.SingleTask
	packTasks []*ttsTask.PackTask
}

func NewSpeech(options ...communicateOption.Option) (*Speech, error) {
	s := &Speech{
		options:   options,
		tasks:     make([]*ttsTask.SingleTask, 0),
		packTasks: make([]*ttsTask.PackTask, 0),
	}
	return s, nil
}

// GetVoiceList  get the list of voices.
func (s *Speech) GetVoiceList() ([]voiceList.Voice, error) {
	return voiceList.ListVoices()
}

// AddSingleTask add a single task to speech.
func (s *Speech) AddSingleTask(text string, output io.Writer) error {
	c, err := communicate.NewCommunicate(text, s.options...)
	if err != nil {
		return err
	}
	task := &ttsTask.SingleTask{
		Text:        text,
		Communicate: c,
		Output:      output,
	}
	s.tasks = append(s.tasks, task)
	return nil
}

// AddPackTask add a pack task to speech.
func (s *Speech) AddPackTask(dataEntries map[string]string, entryCreator func(name string) (io.Writer, error), output io.Writer) error {
	taskCount := len(dataEntries)
	if taskCount == 0 {
		return NoPackTaskEntries
	}
	packEntries := make([]*ttsTask.PackEntry, 0, taskCount)
	for name, text := range dataEntries {
		packEntry := &ttsTask.PackEntry{
			Text:      text,
			EntryName: name,
		}
		packEntries = append(packEntries, packEntry)
	}
	packTask := ttsTask.PackTask{
		CommunicateOpt:   s.options,
		PackEntryCreator: entryCreator,
		PackEntries:      packEntries,
		Output:           output,
	}
	s.packTasks = append(s.packTasks, &packTask)
	return nil
}

func (s *Speech) StartTasks() error {
	wg := &sync.WaitGroup{}
	wg.Add(len(s.tasks) + len(s.packTasks))
	for _, task := range s.tasks {
		go task.Start(wg)
	}
	for _, task := range s.packTasks {
		go task.Start(wg)
	}
	wg.Wait()
	return nil
}
