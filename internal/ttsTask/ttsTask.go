package ttsTask

import (
	"github.com/lib-x/edgetts/internal/communicate"
	"io"
	"sync"
)

type TTSTask struct {
	// Text to be synthesized
	Text string
	// Communicate
	Communicate *communicate.Communicate
	// Output
	Output io.WriteCloser
}

func (t *TTSTask) Start(wg *sync.WaitGroup) error {
	defer wg.Done()
	if err := t.Communicate.WriteStreamTo(t.Output); err != nil {
		return err
	}
	t.Output.Close()
	return nil
}
