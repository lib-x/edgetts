package edgetts

import (
	"errors"
	"fmt"
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

func validate(c *communicateOption) error {
	// Validate voice
	if validVoicePattern.MatchString(c.Voice) {
		voiceParsed := strings.Split(c.Voice, "-")
		lang := voiceParsed[0]
		region := voiceParsed[1]
		name := voiceParsed[2]
		c.Voice = fmt.Sprintf("Microsoft Server Speech Text to Speech Voice (%s-%s, %s)", lang, region, name)
	} else {
		return InvalidVoiceError
	}

	// Validate rate

	if !validRateVolumePattern.MatchString(c.Rate) {
		return InvalidRateError
	}

	// Validate volume
	if !validRateVolumePattern.MatchString(c.Volume) {
		return InvalidVolumeError
	}
	return nil
}
