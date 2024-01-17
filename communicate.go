package edgetts

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	processLimitMax = 16
)

type turnContext struct {
	ServiceTag string `json:"serviceTag"`
}

type turnAudio struct {
	Type     string `json:"type"`
	StreamID string `json:"streamId"`
}

type turnStart struct {
	Context turnContext `json:"context"`
}

type turnResp struct {
	Context turnContext `json:"context"`
	Audio   turnAudio   `json:"audio"`
}

type turnMetaInnerText struct {
	Text         string `json:"Text"`
	Length       int    `json:"Length"`
	BoundaryType string `json:"BoundaryType"`
}

type turnMetaInnerData struct {
	Offset   int               `json:"Offset"`
	Duration int               `json:"Duration"`
	Text     turnMetaInnerText `json:"text"`
}

type turnMetadata struct {
	Type string            `json:"Type"`
	Data turnMetaInnerData `json:"Data"`
}

type turnMeta struct {
	Metadata []turnMetadata `json:"Metadata"`
}

type communicateChunk struct {
	Type     string
	Data     []byte
	Offset   int
	Duration int
	Text     string
}

type communicateTextTask struct {
	id     int
	text   string
	option communicateTextOption

	chunk      chan communicateChunk
	speechData []byte
}

type communicateTextOption struct {
	voice  string
	rate   string
	volume string
}

type communicate struct {
	connList []*websocket.Conn

	option communicateTextOption
	proxy  string

	processorLimit int
	tasks          chan *communicateTextTask
}

func newCommunicate() *communicate {
	return &communicate{
		option: communicateTextOption{
			voice:  "Microsoft Server Speech Text to Speech Voice (zh-CN, XiaoxiaoNeural)",
			rate:   "+0%",
			volume: "+0%",
		},
		processorLimit: processLimitMax,
		tasks:          make(chan *communicateTextTask, processLimitMax),
		connList:       make([]*websocket.Conn, 0, processLimitMax),
	}
}

func (c *communicate) withVoice(voice string) *communicate {
	if voice == "" {
		return c
	}
	match := voiceMatcher.FindStringSubmatch(voice)
	if match != nil {
		lang := match[1]
		region := match[2]
		name := match[3]
		if i := strings.Index(name, "-"); i != -1 {
			region = region + "-" + name[:i]
			name = name[i+1:]
		}
		voice = fmt.Sprintf("Microsoft Server Speech Text to Speech Voice (%s-%s, %s)", lang, region, name)
		if !isValidVoice(voice) {
			return c
		}
		c.option.voice = voice
	}
	return c
}

func (c *communicate) withRate(rate string) *communicate {
	if !isValidRate(rate) {
		return c
	}
	c.option.rate = rate
	return c
}

func (c *communicate) withVolume(volume string) *communicate {
	if !isValidVolume(volume) {
		return c
	}
	c.option.volume = volume
	return c
}

func (c *communicate) withProxy(proxy string) *communicate {
	if proxy == "" {
		return c
	}
	c.proxy = proxy
	return c
}

func (c *communicate) applyDefaultOption(text *communicateTextOption) {
	if text.voice == "" {
		text.voice = c.option.voice
	}
	if text.rate == "" {
		text.rate = c.option.rate
	}
	if text.volume == "" {
		text.volume = c.option.volume
	}
}

func (c *communicate) openWss() *websocket.Conn {
	headers := createWssConnectHeader()

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(fmt.Sprintf("%s&ConnectionId=%s", EdgeWssEndpoint, uuidWithOutDashes()), headers)
	if err != nil {
		log.Fatal("dial:", err)
	}
	c.connList = append(c.connList, conn)
	return conn
}

func createWssConnectHeader() http.Header {
	headers := http.Header{}
	headers.Add("Pragma", "no-cache")
	headers.Add("Cache-Control", "no-cache")
	headers.Add("Origin", "chrome-extension://jdiccldimpdaibmpdkjnbmckianbfold")
	headers.Add("Accept-Encoding", "gzip, deflate, br")
	headers.Add("Accept-Language", "en-US,en;q=0.9")
	headers.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36 Edg/91.0.864.41")
	return headers
}

func (c *communicate) close() {
	for _, conn := range c.connList {
		conn.Close()
	}
}

func (c *communicate) stream(text *communicateTextTask) chan communicateChunk {
	text.chunk = make(chan communicateChunk)
	// texts := splitTextByByteLength(removeIncompatibleCharacters(c.text), calcMaxMsgSize(c.voice, c.rate, c.volume))
	conn := c.openWss()
	date := dateToString()
	c.applyDefaultOption(&text.option)
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("X-Timestamp:%s\r\nContent-Type:application/json; charset=utf-8\r\nPath:speech.config\r\n\r\n{\"context\":{\"synthesis\":{\"audio\":{\"metadataoptions\":{\"sentenceBoundaryEnabled\":false,\"wordBoundaryEnabled\":true},\"outputFormat\":\"audio-24khz-48kbitrate-mono-mp3\"}}}}\r\n", date)))
	conn.WriteMessage(websocket.TextMessage, []byte(ssmlHeadersPlusData(uuidWithOutDashes(), date, mkSSML(
		text.text, text.option.voice, text.option.rate, text.option.volume,
	))))

	go func() {
		// download indicates whether we should be expecting audio data,
		// this is so what we avoid getting binary data from the websocket
		// and falsely thinking it's audio data.
		downloadAudio := false

		// audio_was_received indicates whether we have received audio data
		// from the websocket. This is so we can raise an exception if we
		// don't receive any audio data.
		// audioWasReceived := false

		// finalUtterance := make(map[int]int)
		for {
			// 读取消息
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			switch messageType {
			case websocket.TextMessage:
				parameters, data, _ := getHeadersAndData(data)
				path := parameters["Path"]

				switch path {
				case "turn.start":
					downloadAudio = true
				case "turn.end":
					downloadAudio = false
					text.chunk <- communicateChunk{
						Type: ChunkTypeEnd,
					}
				case "audio.metadata":
					meta := &turnMeta{}
					if err := json.Unmarshal(data, meta); err != nil {
						log.Fatalf("We received a text message, but unmarshal failed.")
					}
					for _, v := range meta.Metadata {
						if v.Type == ChunkTypeWordBoundary {
							text.chunk <- communicateChunk{
								Type:     v.Type,
								Offset:   v.Data.Offset,
								Duration: v.Data.Duration,
								Text:     v.Data.Text.Text,
							}
						} else if v.Type == ChunkTypeSessionEnd {
							continue
						} else {
							log.Fatalf("Unknown metadata type: %s", v.Type)
						}
					}
				case "response":
					log.Println("get response")
				default:
					log.Fatalf("The path from the service is not recognized.\n%s", data)
				}

			case websocket.BinaryMessage:
				if !downloadAudio {
					log.Fatalf("We received a binary message, but we are not expecting one.")
				}
				if len(data) < 2 {
					log.Fatalf("We received a binary message, but it is missing the header length.")
				}
				headerLength := int(binary.BigEndian.Uint16(data[:2]))
				if len(data) < headerLength+2 {
					log.Fatalf("We received a binary message, but it is missing the audio data.")
				}
				text.chunk <- communicateChunk{
					Type: ChunkTypeAudio,
					Data: data[headerLength+2:],
				}
				// audioWasReceived = true
			}
		}
	}()

	return text.chunk
}

func (c *communicate) allocateTask(tasks []*communicateTextTask) {
	for id, t := range tasks {
		t.id = id
		c.tasks <- t
	}
	close(c.tasks)
}

func (c *communicate) process(wg *sync.WaitGroup) {
	defer wg.Done()
	for t := range c.tasks {
		chunk := c.stream(t)
		for {
			v, ok := <-chunk
			if ok {
				if v.Type == ChunkTypeAudio {
					t.speechData = append(t.speechData, v.Data...)
					// } else if v.Type == ChunkTypeWordBoundary {
				} else if v.Type == ChunkTypeEnd {
					close(t.chunk)
					break
				}
			}
		}
	}
}

func (c *communicate) createPool() {
	var wg sync.WaitGroup
	for i := 0; i < c.processorLimit; i++ {
		wg.Add(1)
		go c.process(&wg)
	}
	wg.Wait()
}
