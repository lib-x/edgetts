package edgetts

import (
	"fmt"
	"io"
)

type Speech struct {
	*Communicate

	voice string

	output io.WriteCloser
}

func NewSpeech(c *Communicate, output io.WriteCloser) (*Speech, error) {
	s := &Speech{
		Communicate: c,
		output:      output,
	}
	return s, nil
}

func (s *Speech) GenerateTTS() error {
	op, err := s.Stream()
	if err != nil {
		return err
	}
	defer s.CloseOutput()
	solveCount := 0
	audioData := make([][][]byte, s.AudioDataIndex)
	for i := range op {
		if _, ok := i["end"]; ok {
			solveCount++
			if solveCount == s.AudioDataIndex {
				break
			}
		}
		t, ok := i["type"]
		if ok && t == "audio" {
			data := i["data"].(AudioData)
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
			s.output.Write(data)
		}
	}
	s.output.Close()
	return nil
}
