package edgetts

import (
	"github.com/lib-x/edgetts/internal/communicate"
	"github.com/lib-x/edgetts/internal/communicateOption"
	"github.com/lib-x/edgetts/internal/ttsTask"
	"io"
	"sync"
)

type Speech struct {
	options []communicateOption.Option
	tasks   []*ttsTask.TTSTask
}

func NewSpeech(options ...communicateOption.Option) (*Speech, error) {
	s := &Speech{
		options: options,
		tasks:   make([]*ttsTask.TTSTask, 0),
	}
	return s, nil
}

func (s *Speech) AddTask(text string, output io.Writer) error {
	c, err := communicate.NewCommunicate(text, s.options...)
	if err != nil {
		return err
	}
	task := &ttsTask.TTSTask{
		Text:        text,
		Communicate: c,
		Output:      output,
	}
	s.tasks = append(s.tasks, task)
	return nil
}

func (s *Speech) StartTasks() error {
	wg := &sync.WaitGroup{}
	wg.Add(len(s.tasks))
	for _, task := range s.tasks {
		go task.Start(wg)
	}
	wg.Wait()
	return nil
}
