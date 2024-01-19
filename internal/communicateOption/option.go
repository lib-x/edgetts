package communicateOption

import "github.com/lib-x/edgetts/internal/businessConsts"

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
	if c.Rate == "" {
		c.Rate = "+0%"
	}
	if c.Volume == "" {
		c.Volume = "+0%"
	}

}
