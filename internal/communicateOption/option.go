package communicateOption

import (
	"fmt"
	"strings"

	"github.com/lib-x/edgetts/internal/businessConsts"
)

type CommunicateOption struct {
	Voice            string
	VoiceLangRegion  string
	Pitch            string
	Rate             string
	Volume           string
	HttpProxy        string
	Socket5Proxy     string
	Socket5ProxyUser string
	Socket5ProxyPass string
	IgnoreSSL        bool
}

func (c *CommunicateOption) CheckAndApplyDefaultOption() {
	// Default values
	if c.Voice == "" {
		c.Voice = businessConsts.DefaultVoice
		c.VoiceLangRegion = businessConsts.DefaultVoice
	}
	// try auto fill voiceLangRegion
	if c.VoiceLangRegion == "" {
		voiceParsed := strings.Split(c.Voice, "-")
		lang := voiceParsed[0]
		region := voiceParsed[1]
		name := voiceParsed[2]
		c.VoiceLangRegion = fmt.Sprintf(businessConsts.VoiceNameTemplate, lang, region, name)
	}
	if c.Pitch == "" {
		c.Pitch = "+0Hz"
	}
	if c.Rate == "" {
		c.Rate = "+0%"
	}
	if c.Volume == "" {
		c.Volume = "+0%"
	}

}
