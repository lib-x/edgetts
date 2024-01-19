package validate

import (
	"errors"
	"github.com/lib-x/edgetts/internal/communicateOption"
	"regexp"
)

var (
	validVoicePattern      = regexp.MustCompile(`^([a-z]{2,})-([A-Z]{2,})-(.+Neural)$`)
	validRateVolumePattern = regexp.MustCompile(`^[+-]\d+%$`)
)

var (
	InvalidVoiceError  = errors.New("invalid voice")
	InvalidRateError   = errors.New("invalid rate")
	InvalidVolumeError = errors.New("invalid volume")
)

// WithCommunicateOption validate With a CommunicateOption
func WithCommunicateOption(c *communicateOption.CommunicateOption) error {
	// WithCommunicateOption voice
	if !validVoicePattern.MatchString(c.Voice) {
		return InvalidVoiceError
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
