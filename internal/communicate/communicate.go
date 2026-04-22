package communicate

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lib-x/edgetts/internal/businessConsts"
	"github.com/lib-x/edgetts/internal/communicateOption"
	"github.com/lib-x/edgetts/internal/validate"
)

const (
	ssmlHeaderTemplate      = "X-RequestId:%s\r\nContent-Type:application/ssml+xml\r\nX-Timestamp:%sZ\r\nPath:ssml\r\n\r\n"
	wordBoundaryOffset      = 8_750_000
	binaryMessageHeaderSize = 2
)

type InputType int

const (
	InputText InputType = iota
	InputSSML
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
	inputType InputType
	input     string

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

var errNoAudioReceived = fmt.Errorf("no audio received")

func NewCommunicate(inputType InputType, input string, opt *communicateOption.CommunicateOption) (*Communicate, error) {
	if opt == nil {
		opt = &communicateOption.CommunicateOption{}
	}
	opt.CheckAndApplyDefaultOption()

	if err := validate.WithCommunicateOption(opt); err != nil {
		return nil, err
	}
	return &Communicate{
		inputType: inputType,
		input:     input,
		opt:       opt,
	}, nil
}

// WriteStreamTo writes audio to w using a background context.
func (c *Communicate) WriteStreamTo(w io.Writer) error {
	_, err := c.WriteStreamToContext(context.Background(), w)
	return err
}

// WriteStreamToContext writes audio to w and returns written bytes.
func (c *Communicate) WriteStreamToContext(ctx context.Context, w io.Writer) (int64, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	output, err := c.stream(ctx)
	if err != nil {
		return 0, err
	}

	var written int64
	for payload := range output {
		if errVal, ok := payload["error"]; ok {
			switch v := errVal.(type) {
			case webSocketError:
				return written, fmt.Errorf("websocket error: %s", v.Message)
			case unknownResponse:
				return written, fmt.Errorf("unknown response: %s", v.Message)
			case unexpectedResponse:
				return written, fmt.Errorf("unexpected response: %s", v.Message)
			case noAudioReceived:
				return written, fmt.Errorf("%w: %s", errNoAudioReceived, v.Message)
			default:
				return written, fmt.Errorf("stream error: %v", v)
			}
		}
		if t, isTypedData := payload["type"]; isTypedData && t == "audio" {
			data, ok := payload["data"].(audioData)
			if !ok {
				continue
			}
			n, err := w.Write(data.Data)
			written += int64(n)
			if err != nil {
				return written, fmt.Errorf("write audio payload: %w", err)
			}
		}
	}
	return written, nil
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
	return conn.WriteMessage(websocket.TextMessage, []byte(
		"X-Timestamp:"+currentTime+"\r\n"+
			"Content-Type:application/json; charset=utf-8\r\n"+
			"Path:speech.config\r\n\r\n"+
			`{"context":{"synthesis":{"audio":{"metadataoptions":{"sentenceBoundaryEnabled":false,"wordBoundaryEnabled":true},"outputFormat":"audio-24khz-48kbitrate-mono-mp3"}}}}`+"\r\n",
	))
}

func (c *Communicate) sendSSML(conn *websocket.Conn, currentTime string, text []byte) error {
	payload := string(text)
	if c.inputType == InputText {
		payload = makeSsml(string(text), c.opt.Pitch, c.opt.Voice, c.opt.Rate, c.opt.Volume)
	}
	return conn.WriteMessage(websocket.TextMessage,
		[]byte(appendRequestContextToSsmlHeaders(generateConnectID(), currentTime, payload)))
}

func (c *Communicate) stream(ctx context.Context) (chan map[string]interface{}, error) {
	output := make(chan map[string]interface{})
	texts := c.buildPayloads()
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
					output <- map[string]interface{}{"error": webSocketError{Message: err.Error()}}
					return
				}
				defer conn.Close()

				currentTime := currentTimeInMST()
				if err := c.sendSpeechGenerationConfig(conn, currentTime); err != nil {
					output <- map[string]interface{}{"error": unexpectedResponse{Message: err.Error()}}
					return
				}
				if err := c.sendSSML(conn, currentTime, text); err != nil {
					output <- map[string]interface{}{"error": unexpectedResponse{Message: err.Error()}}
					return
				}
				c.connStreamExchange(ctx, conn, output, idx)
			}()
		}
	}()

	return output, nil
}

func (c *Communicate) buildPayloads() [][]byte {
	switch c.inputType {
	case InputSSML:
		return [][]byte{[]byte(c.input)}
	default:
		return splitTextByByteLength(
			escape(removeIncompatibleCharacters(c.input)),
			getMaxMessageSize(c.opt.Pitch, c.opt.Voice, c.opt.Rate, c.opt.Volume),
		)
	}
}

func (c *Communicate) connStreamExchange(ctx context.Context, conn *websocket.Conn, output chan map[string]interface{}, idx int) {
	downloadAudio := false
	audioWasReceived := false

	for {
		select {
		case <-ctx.Done():
			if !audioWasReceived {
				output <- map[string]interface{}{"error": noAudioReceived{Message: "context cancelled before audio was received"}}
			}
			return
		default:
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				if !audioWasReceived {
					output <- map[string]interface{}{"error": webSocketError{Message: err.Error()}}
				}
				return
			}
			continueProcessing := c.handleWebSocketMessage(msgType, message, output, idx, &downloadAudio, &audioWasReceived)
			if !continueProcessing {
				if !audioWasReceived {
					output <- map[string]interface{}{"error": noAudioReceived{Message: "no audio data returned by service"}}
				}
				return
			}
		}
	}
}
