package ttsTask

import (
	"github.com/lib-x/edgetts/internal/communicate"
	"github.com/lib-x/edgetts/internal/communicateOption"
	"io"
	"log"
	"sync"
)

type PackEntry struct {
	// Text to be synthesized
	Text string
	// Entry name to be packed into a file
	EntryName string
}

type PackTask struct {
	// PackEntryCreator defines the function to create a writer for each entry
	PackEntryCreator func(string) (io.Writer, error)
	// CommunicateOpt defines the options for communicating with the TTS engine
	CommunicateOpt []communicateOption.Option
	// PackEntries defines the list of entries to be packed into a file
	PackEntries []*PackEntry
	// Output
	Output io.Writer
}

func (p *PackTask) Start(wg *sync.WaitGroup) error {
	defer wg.Done()
	for _, entry := range p.PackEntries {
		c, err := communicate.NewCommunicate(entry.Text, p.CommunicateOpt...)
		if err != nil {
			log.Printf("create communicate error:%v \r\n", err)
			continue
		}
		entryWriter, err := p.PackEntryCreator(entry.EntryName)
		err = c.WriteStreamTo(entryWriter)
		if err != nil {
			log.Printf("write data to entry writer error:%v \r\n", err)
			return err
		}
	}
	return nil
}
