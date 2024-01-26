package validate

import (
	"errors"
	"regexp"

	"github.com/lib-x/edgetts/internal/communicateOption"
)

var (
	validPitchPattern      = regexp.MustCompile(`^[+-]\d+Hz$`)
	validVoicePattern      = regexp.MustCompile(`^([a-z]{2,})-([A-Z]{2,})-(.+Neural)$`)
	validRateVolumePattern = regexp.MustCompile(`^[+-]\d+%$`)
)

var (
	InvalidVoiceError  = errors.New("invalid voice")
	InvalidPitchError  = errors.New("invalid pitch")
	InvalidRateError   = errors.New("invalid rate")
	InvalidVolumeError = errors.New("invalid volume")
)

// WithCommunicateOption validate With a CommunicateOption
func WithCommunicateOption(c *communicateOption.CommunicateOption) error {
	// WithCommunicateOption voice
	if !validVoicePattern.MatchString(c.Voice) {
		return InvalidVoiceError
	}

	// WithCommunicateOption pitch
	if !validPitchPattern.MatchString(c.Pitch) {
		return InvalidPitchError
	}
	// WithCommunicateOption rate
	if !validRateVolumePattern.MatchString(c.Rate) {
		return InvalidRateError
	}

	// WithCommunicateOption volume
	if !validRateVolumePattern.MatchString(c.Volume) {
		return InvalidVolumeError
	}

	return nil
}
