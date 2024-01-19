package communicateOption

import (
	"fmt"
	"github.com/lib-x/edgetts/internal/businessConsts"
	"strings"
)

type CommunicateOption struct {
	Voice            string
	VoiceLangRegion  string
	Rate             string
	Volume           string
	HttpProxy        string
	Socket5Proxy     string
	Socket5ProxyUser string
	Socket5ProxyPass string
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
	if c.Rate == "" {
		c.Rate = "+0%"
	}
	if c.Volume == "" {
		c.Volume = "+0%"
	}

}
