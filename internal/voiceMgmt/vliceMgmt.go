package voiceMgmt

import (
	"encoding/json"
	"github.com/lib-x/edgetts/internal/businessConsts"
	"net/http"
	"sync"
)

var (
	getVoiceHeader http.Header
	headerOnce     = &sync.Once{}
)

type Voice struct {
	Name           string `json:"Name"`
	ShortName      string `json:"ShortName"`
	Gender         string `json:"Gender"`
	Locale         string `json:"Locale"`
	SuggestedCodec string `json:"SuggestedCodec"`
	FriendlyName   string `json:"FriendlyName"`
	Status         string `json:"Status"`
	Language       string
	VoiceTag       VoiceTag `json:"VoiceTag"`
}
type VoiceTag struct {
	ContentCategories  []string `json:"ContentCategories"`
	VoicePersonalities []string `json:"VoicePersonalities"`
}

type VoiceManager struct {
}

func NewVoiceManager() *VoiceManager {
	headerOnce.Do(func() {
		getVoiceHeader = makeVoiceListRequestHeader()
	})
	return &VoiceManager{}
}

func (m *VoiceManager) ListVoices() ([]Voice, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", businessConsts.VoiceListEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header = getVoiceHeader
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var voices []Voice
	if err = json.NewDecoder(resp.Body).Decode(&voices); err != nil {
		return nil, err
	}
	return voices, nil
}

func makeVoiceListRequestHeader() http.Header {
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	header.Set("Accept", "*/*")
	header.Set("Authority", "speech.platform.bing.com")
	header.Set("Sec-CH-UA", `" Not;A Brand";v="99", "Microsoft Edge";v="91", "Chromium";v="91"`)
	header.Set("Sec-CH-UA-Mobile", "?0")
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36 Edg/91.0.864.41")
	header.Set("Sec-Fetch-Site", "none")
	header.Set("Sec-Fetch-Mode", "cors")
	header.Set("Sec-Fetch-Dest", "empty")
	header.Set("Accept-Encoding", "gzip, deflate, br")
	header.Set("Accept-Language", "en-US,en;q=0.9")
	return header
}
