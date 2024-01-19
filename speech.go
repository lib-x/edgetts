package edgetts

import (
	"errors"
	"github.com/lib-x/edgetts/internal/communicate"
	"github.com/lib-x/edgetts/internal/communicateOption"
	"github.com/lib-x/edgetts/internal/ttsTask"
	"io"
	"sync"
)

var (
	NoPackTaskEntries = errors.New("no pack task entries")
)

type Speech struct {
	vm        *VoiceManager
	options   []Option
	tasks     []*ttsTask.SingleTask
	packTasks []*ttsTask.PackTask
}

func (s *Speech) convertToInternalOpt() *communicateOption.CommunicateOption {
	opt := &option{}
	for _, apply := range s.options {
		apply(opt)
	}
	return opt.toInternalOption()

}

// NewSpeech creates a new Speech instance.
// It takes a variadic parameter:
// - options: a slice of communicateOption.Option that will be used to configure the Speech instance.
// The function returns a pointer to the newly created Speech instance and an error if any occurs during the creation process.
func NewSpeech(options ...Option) (*Speech, error) {
	s := &Speech{
		options:   options,
		tasks:     make([]*ttsTask.SingleTask, 0),
		packTasks: make([]*ttsTask.PackTask, 0),
		vm:        NewVoiceManager(),
	}
	return s, nil
}

// GetVoiceList retrieves the list of voices available for the speech.
// It returns a slice of Voice objects and an error if any occurs during the retrieval process.
func (s *Speech) GetVoiceList() ([]Voice, error) {
	return s.vm.ListVoices()
}

// AddSingleTask adds a single task to the speech.
// It takes two parameters:
// - text: the text to be synthesized.
// - output: the output of the single task, which will finally be written into a file.
// The function returns an error if there is an issue with the communication.
func (s *Speech) AddSingleTask(text string, output io.Writer) error {
	opt := s.convertToInternalOpt()
	c, err := communicate.NewCommunicate(text, opt)
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

// AddPackTask adds a pack task to the speech.
// It takes four parameters:
// - dataEntries: a map where the key is the entry name and the value is the entry text to be synthesized.
// - entryCreator: a function that creates a writer for each entry. This can be a packer context related writer, such as a zip writer.
// - output: the output of the pack task, which will finally be written into a file.
// - metaData: optional parameter. It is the data which will be serialized into a json file. The name uses the key and value as the key-value pair.
// The function returns an error if there are no pack task entries.
func (s *Speech) AddPackTask(dataEntries map[string]string, entryCreator func(name string) (io.Writer, error), output io.Writer, metaData ...map[string]any) error {
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

	packTask := &ttsTask.PackTask{
		CommunicateOpt:   s.convertToInternalOpt(),
		PackEntryCreator: entryCreator,
		PackEntries:      packEntries,
		Output:           output,
		MetaData:         metaData,
	}
	s.packTasks = append(s.packTasks, packTask)
	return nil
}

// StartTasks starts all the tasks in the Speech instance.
// It creates a WaitGroup and adds the total number of tasks to it.
// Then it starts each task in a separate goroutine and waits for all of them to finish.
// The function returns an error if any occurs during the execution of the tasks.
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
