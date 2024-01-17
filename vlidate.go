package edgetts

import "regexp"

var (
	voiceMatcher       = regexp.MustCompile(`^([a-z]{2,})-([A-Z]{2,})-(.+Neural)$`)
	voiceNameMatcher   = regexp.MustCompile(`^Microsoft Server Speech Text to Speech Voice \(.+,.+\)$`)
	numberParamMatcher = regexp.MustCompile(`^[+-]\d+%$`)
)

func isValidVoice(voice string) bool {
	return voiceNameMatcher.MatchString(voice)
}

func isValidRate(rate string) bool {
	if rate == "" {
		return false
	}
	return numberParamMatcher.MatchString(rate)
}

func isValidVolume(volume string) bool {
	if volume == "" {
		return false
	}
	return numberParamMatcher.MatchString(volume)
}
