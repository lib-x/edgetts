package ttsTask

import (
	"encoding/json"
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
	// EntryCommunicateOpt defines the options for communicating with the TTS engine.if note set, use the PackTask's CommunicateOpt.
	EntryCommunicateOpt *communicateOption.CommunicateOption
}

type PackTask struct {
	// PackEntryCreator defines the function to create a writer for each entry
	PackEntryCreator func(string) (io.Writer, error)
	// CommunicateOpt defines the options for communicating with the TTS engine
	CommunicateOpt *communicateOption.CommunicateOption
	// PackEntries defines the list of entries to be packed into a file
	PackEntries []*PackEntry
	// Output
	Output io.Writer
	// MetaData is the data which will be serialized into a json file,name use the key and value as the key-value pair.
	MetaData []map[string]any
}

func (p *PackTask) Start(wg *sync.WaitGroup) error {
	defer wg.Done()
	for _, entry := range p.PackEntries {
		// for zip file, the entry should be written after creation.
		opt := p.CommunicateOpt
		if entry.EntryCommunicateOpt != nil {
			opt = entry.EntryCommunicateOpt

		}
		c, err := communicate.NewCommunicate(entry.Text, opt)
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
	// after all entries are written, write the meta data into a json file. this process is optional.
	// so error is ignored.

	if len(p.MetaData) > 0 {
		for _, metaData := range p.MetaData {
			for entryName, entryPayload := range metaData {
				metaEntry, err := p.PackEntryCreator(entryName)
				if err != nil {
					log.Printf("create meta entry writer error:%v \r\n", err)
					continue
				}
				if err = json.NewEncoder(metaEntry).Encode(entryPayload); err != nil {
					log.Printf("write data to meta entry writer error:%v \r\n", err)
					continue
				}
			}
		}
	}
	return nil
}
