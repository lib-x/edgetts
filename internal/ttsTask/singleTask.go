package ttsTask

import (
	"io"
	"sync"

	"github.com/lib-x/edgetts/internal/communicate"
)

type SingleTask struct {
	// Text to be synthesized
	Text string
	// Communicate
	Communicate *communicate.Communicate
	// Output
	Output io.Writer
}

// Start  start a tts single task
func (t *SingleTask) Start(wg *sync.WaitGroup) error {
	defer wg.Done()
	if err := t.Communicate.WriteStreamTo(t.Output); err != nil {
		return err
	}
	if closer, ok := t.Output.(io.Closer); ok {
		// log.Print("ttsTask.Start: close output writer\r\n")
		closer.Close()
	}
	return nil
}
