package edgetts

import (
	"fmt"
	"github.com/lib-x/edgetts/internal/communicate"
	"github.com/lib-x/edgetts/internal/communicateOption"
	"io"
	"sync"
)

type Speech struct {
	options []communicateOption.Option
	tasks   []*ttsTask
}

type ttsTask struct {
	// Text to be synthesized
	text string
	// communicate
	communicate *communicate.Communicate
	// output
	output io.WriteCloser
}

func (t ttsTask) start(wg *sync.WaitGroup) error {
	defer wg.Done()
	op, err := t.communicate.Stream()
	if err != nil {
		return err
	}
	defer t.output.Close()
	solveCount := 0
	audioData := make([][][]byte, t.communicate.AudioDataIndex)
	for i := range op {
		if _, ok := i["end"]; ok {
			solveCount++
			if solveCount == t.communicate.AudioDataIndex {
				break
			}
		}
		t, ok := i["type"]
		if ok && t == "audio" {
			data := i["data"].(communicate.AudioData)
			audioData[data.Index] = append(audioData[data.Index], data.Data)
		}
		e, ok := i["error"]
		if ok {
			fmt.Printf("has error err: %v\n", e)
		}
	}
	// write data, sort by index
	for _, v := range audioData {
		for _, data := range v {
			t.output.Write(data)
		}
	}
	return nil
}

func NewSpeech(options ...communicateOption.Option) (*Speech, error) {
	s := &Speech{
		options: options,
		tasks:   make([]*ttsTask, 0),
	}
	return s, nil
}

func (s *Speech) AddTask(text string, output io.WriteCloser) error {
	c, err := communicate.NewCommunicate(text, s.options...)
	if err != nil {
		return err
	}
	task := &ttsTask{
		text:        text,
		communicate: c,
		output:      output,
	}
	s.tasks = append(s.tasks, task)
	return nil
}

func (s *Speech) StartTasks() error {
	wg := &sync.WaitGroup{}
	wg.Add(len(s.tasks))
	for _, task := range s.tasks {
		go task.start(wg)
	}
	wg.Wait()
	return nil
}
