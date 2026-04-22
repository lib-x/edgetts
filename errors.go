package edgetts

import "errors"

var (
	ErrEmptyInput      = errors.New("empty input")
	ErrBatchEmpty      = errors.New("empty batch")
	ErrVoiceNotFound   = errors.New("voice not found")
	ErrNoAudioReceived = errors.New("no audio received")
)
