package validate

import (
	"errors"
	"fmt"
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

// ValidateCommunicateOptions validate CommunicateOptions
func ValidateCommunicateOptions(c *communicateOption.CommunicateOption) error {
	// ValidateCommunicateOptions voice
	if validVoicePattern.MatchString(c.Voice) {
		voiceParsed := strings.Split(c.Voice, "-")
		lang := voiceParsed[0]
		region := voiceParsed[1]
		name := voiceParsed[2]
		c.Voice = fmt.Sprintf("Microsoft Server Speech Text to Speech Voice (%s-%s, %s)", lang, region, name)
	} else {
		return InvalidVoiceError
	}

	// ValidateCommunicateOptions rate

	if !validRateVolumePattern.MatchString(c.Rate) {
		return InvalidRateError
	}

	// ValidateCommunicateOptions volume
	if !validRateVolumePattern.MatchString(c.Volume) {
		return InvalidVolumeError
	}
	return nil
}
