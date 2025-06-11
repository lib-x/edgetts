package communicate

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/lib-x/edgetts/internal/businessConsts"
	"github.com/lib-x/edgetts/internal/communicateOption"
	"github.com/lib-x/edgetts/internal/validate"
	"io"
	"log"
	"net/http"
	"sync"
)

const (
	ssmlHeaderTemplate      = "X-RequestId:%s\r\nContent-Type:application/ssml+xml\r\nX-Timestamp:%sZ\r\nPath:ssml\r\n\r\n"
	wordBoundaryOffset      = 8_750_000
	binaryMessageHeaderSize = 2
)

var (
	headerOnce        = &sync.Once{}
	communicateHeader http.Header
)

func init() {
	headerOnce.Do(func() {
		communicateHeader = makeDefaultHeaders()
	})
}

type Communicate struct {
	text string

	audioDataIndex int
	prevIdx        int
	shiftTime      int
	finalUtterance map[int]int
	opt            *communicateOption.CommunicateOption
}

type textEntry struct {
	Text         string `json:"text"`
	BoundaryType string `json:"BoundaryType"`
	Length       int64  `json:"Length"`
}
type dataEntry struct {
	Offset   int       `json:"Offset"`
	Duration int       `json:"Duration"`
	Text     textEntry `json:"text"`
}
type metaDataEntry struct {
	Type string    `json:"Type"`
	Data dataEntry `json:"Data"`
}
type metaDataContext struct {
	Metadata []metaDataEntry `json:"Metadata"`
}

type audioData struct {
	Data  []byte
	Index int
}

type unknownResponse struct {
	Message string
}

type unexpectedResponse struct {
	Message string
}

type noAudioReceived struct {
	Message string
}

type webSocketError struct {
	Message string
}

func NewCommunicate(text string, opt *communicateOption.CommunicateOption) (*Communicate, error) {
	if opt == nil {
		opt = &communicateOption.CommunicateOption{}
	}
	opt.CheckAndApplyDefaultOption()

	if err := validate.WithCommunicateOption(opt); err != nil {
		return nil, err
	}
	return &Communicate{
		text: text,
		opt:  opt,
	}, nil
}

// WriteStreamTo  write audio stream to io.WriteCloser
func (c *Communicate) WriteStreamTo(rc io.Writer) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	output, err := c.stream(ctx)
	if err != nil {
		return err
	}

	for payload := range output {
		if t, isTypedData := payload["type"]; isTypedData && t == "audio" {
			if data, ok := payload["data"].(audioData); ok {
				rc.Write(data.Data)
			}
		}
	}
	return nil
}

func makeDefaultHeaders() http.Header {
	header := make(http.Header)
	header.Set("Pragma", "no-cache")
	header.Set("Cache-Control", "no-cache")
	header.Set("Origin", "chrome-extension://jdiccldimpdaibmpdkjnbmckianbfold")
	header.Set("Accept-Encoding", "gzip, deflate, br")
	header.Set("Accept-Language", "en-US,en;q=0.9")
	header.Set("User-Agent", fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36  Edg/%s.0.0.0",
		businessConsts.ChromiumMajorVersion,
		businessConsts.ChromiumMajorVersion,
	))
	return header
}

func (c *Communicate) sendSpeechGenerationConfig(conn *websocket.Conn, currentTime string) error {
	// Prepare the request to be sent to the service.
	//
	// Note sentenceBoundaryEnabled and wordBoundaryEnabled are actually supposed
	// to be booleans, but Edge Browser seems to send them as strings.
	//
	// This is a bug in Edge as Azure Cognitive Services actually sends them as
	// bool and not string. For now I will send them as bool unless it causes
	// any problems.
	//
	// Also pay close attention to double { } in request (escape for f-string).
	return conn.WriteMessage(websocket.TextMessage, []byte(
		"X-Timestamp:"+currentTime+"\r\n"+
			"Content-Type:application/json; charset=utf-8\r\n"+
			"Path:speech.config\r\n\r\n"+
			`{"context":{"synthesis":{"audio":{"metadataoptions":{"sentenceBoundaryEnabled":false,"wordBoundaryEnabled":true},"outputFormat":"audio-24khz-48kbitrate-mono-mp3"}}}}`+"\r\n",
	))
}

func (c *Communicate) sendSSML(conn *websocket.Conn, currentTime string, text []byte) error {
	return conn.WriteMessage(websocket.TextMessage,
		[]byte(
			appendRequestContextToSsmlHeaders(
				generateConnectID(),
				currentTime,
				makeSsml(string(text), c.opt.Pitch, c.opt.Voice, c.opt.Rate, c.opt.Volume),
			),
		))
}

func (c *Communicate) stream(ctx context.Context) (chan map[string]interface{}, error) {
	output := make(chan map[string]interface{})
	texts := splitTextByByteLength(
		escape(removeIncompatibleCharacters(c.text)),
		getMaxMessageSize(c.opt.Pitch, c.opt.Voice, c.opt.Rate, c.opt.Volume),
	)
	c.audioDataIndex = len(texts)
	c.finalUtterance = make(map[int]int)
	c.prevIdx = -1
	c.shiftTime = -1
	go func() {
		defer close(output)
		for idx, text := range texts {
			func() {
				wsURL := generateWssEndpoint()
				dialer := websocket.Dialer{}
				c.applyWebSocketProxyIfSet(&dialer)
				conn, _, err := dialer.Dial(wsURL, communicateHeader)
				if err != nil {
					output <- map[string]interface{}{
						"error": webSocketError{Message: err.Error()},
					}
					return
				}
				defer conn.Close()
				currentTime := currentTimeInMST()
				err = c.sendSpeechGenerationConfig(conn, currentTime)
				if err != nil {
					log.Println("sendSpeechGenerationConfig error:", err)
					return
				}
				if err = c.sendSSML(conn, currentTime, text); err != nil {
					log.Println("sendSSML error:", err)
					return
				}
				c.connStreamExchange(ctx, conn, output, idx)
			}()
		}
	}()

	return output, nil
}

func (c *Communicate) connStreamExchange(ctx context.Context, conn *websocket.Conn, output chan map[string]interface{}, idx int) {
	// download indicates whether we should be expecting audio data,
	// this is so what we avoid getting binary data from the websocket
	// and falsely thinking it's audio data.
	downloadAudio := false

	// audioWasReceived indicates whether we have received audio data
	// from the websocket. This is so we can raise an exception if we
	// don't receive any audio data.
	audioWasReceived := false

	for {
		select {
		case <-ctx.Done():
			if !audioWasReceived {
				log.Println("No audio was received. Please verify that your parameters are correct.")
			}
			return
		default:
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read message from conn err", err)
				return
			}
			continueProcessing := c.handleWebSocketMessage(msgType, message, output, idx, &downloadAudio, &audioWasReceived)
			if !continueProcessing {
				return
			}
		}
	}
}
