package edgetts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/lib-x/edgetts/internal/businessConsts"
)

var (
	getVoiceHeader http.Header
	theaderOnce    = &sync.Once{}
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
	client *http.Client
}

func NewVoiceManager() *VoiceManager {
	theaderOnce.Do(func() {
		getVoiceHeader = makeVoiceListRequestHeader()
	})
	return &VoiceManager{client: &http.Client{}}
}

func (m *VoiceManager) ListVoices() ([]Voice, error) {
	return m.ListVoicesContext(context.Background())
}

func (m *VoiceManager) ListVoicesContext(ctx context.Context) ([]Voice, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, businessConsts.VoiceListEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create voice list request: %w", err)
	}
	req.Header = getVoiceHeader.Clone()

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request voices: %w", err)
	}
	defer resp.Body.Close()

	var voices []Voice
	if err := json.NewDecoder(resp.Body).Decode(&voices); err != nil {
		return nil, fmt.Errorf("decode voices: %w", err)
	}
	return voices, nil
}

func makeVoiceListRequestHeader() http.Header {
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	header.Set("Accept", "*/*")
	header.Set("Authority", "speech.platform.bing.com")
	header.Set("Sec-CH-UA", fmt.Sprintf(`" Not;A Brand";v="99", "Microsoft Edge";v="%s", "Chromium";v="%s"`,
		businessConsts.ChromiumMajorVersion,
		businessConsts.ChromiumMajorVersion,
	))
	header.Set("Sec-CH-UA-Mobile", "?0")
	header.Set("User-Agent", fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36 Edg/%s.0.0.0",
		businessConsts.ChromiumMajorVersion,
		businessConsts.ChromiumMajorVersion,
	))
	header.Set("Sec-Fetch-Site", "none")
	header.Set("Sec-Fetch-Mode", "cors")
	header.Set("Sec-Fetch-Dest", "empty")
	header.Set("Accept-Encoding", "gzip, deflate, br")
	header.Set("Accept-Language", "en-US,en;q=0.9")
	return header
}
