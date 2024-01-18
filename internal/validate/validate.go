package validate

import (
	"errors"
	"fmt"
	"github.com/lib-x/edgetts/internal/businessConsts"
	"github.com/lib-x/edgetts/internal/communicateOption"
	"regexp"
	"strings"
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
	if validVoicePattern.MatchString(c.Voice) {
		voiceParsed := strings.Split(c.Voice, "-")
		lang := voiceParsed[0]
		region := voiceParsed[1]
		name := voiceParsed[2]
		c.Voice = fmt.Sprintf(businessConsts.VoiceNameTemplate, lang, region, name)
	} else {
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
