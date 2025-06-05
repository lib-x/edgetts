package communicate

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/lib-x/edgetts/internal/businessConsts"
	"github.com/lib-x/edgetts/internal/communicateOption"
	"github.com/lib-x/edgetts/internal/validate"
	"golang.org/x/net/proxy"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	ssmlHeaderTemplate      = "X-RequestId:%s\r\nContent-Type:application/ssml+xml\r\nX-Timestamp:%sZ\r\nPath:ssml\r\n\r\n"
	wordBoundaryOffset      = 8_750_000
	binaryMessageHeaderSize = 2
)

var (
	headerOnce        = &sync.Once{}
	escapeReplacer    = strings.NewReplacer(">", "&gt;", "<", "&lt;")
	communicateHeader http.Header
)

func init() {
	headerOnce.Do(func() {
		communicateHeader = makeHeaders()
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
type metaHub struct {
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
		if t, ok := payload["type"]; ok && t == "audio" {
			data := payload["data"].(audioData)
			rc.Write(data.Data)
		}
	}
	return nil
}

func makeHeaders() http.Header {
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

func (c *Communicate) sendConfig(conn *websocket.Conn, currentTime string) error {
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
			ssmlHeadersAppendExtraData(
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
		calculateMaxMessageSize(c.opt.Pitch, c.opt.Voice, c.opt.Rate, c.opt.Volume),
	)
	c.audioDataIndex = len(texts)
	c.finalUtterance = make(map[int]int)
	c.prevIdx = -1
	c.shiftTime = -1
	
	go func() {
		defer close(output)
		for idx, text := range texts {
			func() {
				wsURL := businessConsts.EdgeWssEndpoint +
					"&Sec-MS-GEC=" + GenerateSecMsGecToken() +
					"&Sec-MS-GEC-Version=" + GenerateSecMsGecVersion() +
					"&ConnectionId=" + generateConnectID()
				dialer := websocket.Dialer{}
				setupWebSocketProxy(&dialer, c)

				conn, _, err := dialer.Dial(wsURL, communicateHeader)
				if err != nil {
					output <- map[string]interface{}{
						"error": webSocketError{Message: err.Error()},
					}
					return
				}
				defer conn.Close()
				currentTime := currentTimeInMST()
				err = c.sendConfig(conn, currentTime)
				if err != nil {
					log.Println("sendConfig error:", err)
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

func sumWithMap(idx int, m map[int]int) int {
	sumResult := 0
	for i := 0; i < idx; i++ {
		sumResult += m[i]
	}
	return sumResult
}

func setupWebSocketProxy(dialer *websocket.Dialer, c *Communicate) {
	if c.opt.HttpProxy != "" {
		proxyUrl, _ := url.Parse(c.opt.HttpProxy)
		dialer.Proxy = http.ProxyURL(proxyUrl)
	}
	if c.opt.Socket5Proxy != "" {
		auth := &proxy.Auth{User: c.opt.Socket5ProxyUser, Password: c.opt.Socket5ProxyPass}
		socket5ProxyDialer, _ := proxy.SOCKS5("tcp", c.opt.Socket5Proxy, auth, proxy.Direct)
		dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
			return socket5ProxyDialer.Dial(network, address)
		}
		dialer.NetDialContext = dialContext
	}
	if c.opt.IgnoreSSL {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
}

func getMetaHubFrom(data []byte) (*metaHub, error) {
	metadata := &metaHub{}
	err := json.Unmarshal(data, &metadata)
	if err != nil {
		return nil, fmt.Errorf("err=%s, data=%s", err.Error(), string(data))
	}
	return metadata, nil
}

// processWebsocketTextMessage parses a websocket text message into headers and body.
// It returns a map of headers and the message body as a byte slice.
func processWebsocketTextMessage(data []byte) (headers map[string]string, body []byte) {
	headers = make(map[string]string)
	// Find the end of the headers section
	headerEndIndex := bytes.Index(data, []byte("\r\n\r\n"))
	if headerEndIndex == -1 {
		// If there's no header separator, treat the entire message as body
		return headers, data
	}
	// Split headers into individual lines
	headerLines := bytes.Split(data[:headerEndIndex], []byte("\r\n"))
	// Parse each header line
	for _, line := range headerLines {
		parts := bytes.SplitN(line, []byte(":"), 2)
		if len(parts) == 2 {
			key := string(bytes.TrimSpace(parts[0]))
			value := string(bytes.TrimSpace(parts[1]))
			headers[key] = value
		}
		// Ignore malformed header lines
	}
	// The body starts after the headers
	body = data[headerEndIndex+4:]

	return headers, body
}

func removeIncompatibleCharacters(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			return ' '
		}
		return r
	}, str)
}

func generateConnectID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// splitTextByByteLength splits the input text into chunks of a specified byte length.
// The function ensures that the text is not split in the middle of a word. If a word exceeds the specified byte length,
// the word is placed in the next chunk. The function returns a slice of byte slices, each representing a chunk of the original text.
//
// Parameters:
// text: The input text to be split.
// byteLength: The maximum byte length for each chunk of text.
//
// Returns:
// A slice of byte slices, each representing a chunk of the original text.
func splitTextByByteLength(text string, byteLength int) [][]byte {
	var result [][]byte
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if len(data) > byteLength {
			return byteLength, data[:byteLength], nil
		}
		return len(data), data, bufio.ErrFinalToken
	})

	for scanner.Scan() {
		result = append(result, bytes.TrimSpace(scanner.Bytes()))
	}

	return result
}

func makeSsml(text string, pitch, voice string, rate string, volume string) string {
	ssml := &Speak{
		XMLName: xml.Name{Local: "speak"},
		Version: "1.0",
		Xmlns:   "http://www.w3.org/2001/10/synthesis",
		Lang:    "en-US",
		Voice: []Voice{{
			Name: voice,
			Prosody: Prosody{
				Pitch:  pitch,
				Rate:   rate,
				Volume: volume,
				Text:   text,
			},
		}},
	}

	output, err := xml.MarshalIndent(ssml, "", "  ")
	if err != nil {
		return ""
	}
	return string(output)
}

func currentTimeInMST() string {
	// Use time.FixedZone to represent a fixed timezone offset of 0 (UTC)
	zone := time.FixedZone("UTC", 0)
	now := time.Now().In(zone)
	return now.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
}

func ssmlHeadersAppendExtraData(requestID string, timestamp string, ssml string) string {
	headers := fmt.Sprintf(
		ssmlHeaderTemplate,
		requestID,
		timestamp,
	)
	return headers + ssml
}

func calculateMaxMessageSize(pitch, voice string, rate string, volume string) int {
	websocketMaxSize := 1 << 16
	overheadPerMessage := len(ssmlHeadersAppendExtraData(generateConnectID(), currentTimeInMST(), makeSsml("", pitch, voice, rate, volume))) + 50
	return websocketMaxSize - overheadPerMessage
}

func escape(data string) string {
	// Must do ampersand first
	entities := make(map[string]string)
	data = escapeReplacer.Replace(data)
	if entities != nil {
		data = replaceWithDict(data, entities)
	}
	return data
}

func replaceWithDict(data string, entities map[string]string) string {
	for key, value := range entities {
		data = strings.ReplaceAll(data, key, value)
	}
	return data
}

func (c *Communicate) handleWebSocketMessage(msgType int, message []byte, output chan map[string]interface{}, idx int, downloadAudio *bool, audioWasReceived *bool) bool {
	switch msgType {
	case websocket.TextMessage:
		return c.handleTextMessage(message, output, idx, downloadAudio)
	case websocket.BinaryMessage:
		return c.handleBinaryMessage(message, output, idx, downloadAudio, audioWasReceived)
	default:
		log.Printf("Received unknown message type: %d", msgType)
		return true
	}
}

func (c *Communicate) handleTextMessage(message []byte, output chan map[string]interface{}, idx int, downloadAudio *bool) bool {
	parameters, data := processWebsocketTextMessage(message)
	path := parameters["Path"]

	switch path {
	case "turn.start":
		*downloadAudio = true
	case "turn.end":
		output <- map[string]interface{}{
			"end": "",
		}
		*downloadAudio = false
		return false // End of audio data

	case "audio.metadata":
		meta, err := getMetaHubFrom(data)
		if err != nil {
			output <- map[string]interface{}{
				"error": unknownResponse{Message: err.Error()},
			}
			return false
		}

		for _, metaObj := range meta.Metadata {
			metaType := metaObj.Type
			if idx != c.prevIdx {
				c.shiftTime = sumWithMap(idx, c.finalUtterance)
				c.prevIdx = idx
			}
			switch metaType {
			case "WordBoundary":
				c.finalUtterance[idx] = metaObj.Data.Offset + metaObj.Data.Duration + wordBoundaryOffset
				output <- map[string]interface{}{
					"type":     metaType,
					"offset":   metaObj.Data.Offset + c.shiftTime,
					"duration": metaObj.Data.Duration,
					"text":     metaObj.Data.Text,
				}
			case "SessionEnd":
				// do nothing
			default:
				output <- map[string]interface{}{
					"error": unknownResponse{Message: "Unknown metadata type: " + metaType},
				}
				return false
			}
		}
	case "response":
		// do nothing
	default:
		output <- map[string]interface{}{
			"error": unknownResponse{Message: "The response from the service is not recognized.\n" + string(message)},
		}
		return false
	}
	return true
}

func (c *Communicate) handleBinaryMessage(message []byte, output chan map[string]interface{}, idx int, downloadAudio *bool, audioWasReceived *bool) bool {
	if !*downloadAudio {
		output <- map[string]interface{}{
			"error": unknownResponse{"We received a binary message, but we are not expecting one."},
		}
		return false
	}

	if len(message) < binaryMessageHeaderSize {
		output <- map[string]interface{}{
			"error": unknownResponse{"We received a binary message, but it is missing the header length."},
		}
		return false
	}

	headerLength := int(binary.BigEndian.Uint16(message[:2]))
	if len(message) < headerLength+2 {
		output <- map[string]interface{}{
			"error": unknownResponse{"We received a binary message, but it is missing the audio data."},
		}
		return false
	}

	audioBinaryData := message[headerLength+2:]
	output <- map[string]interface{}{
		"type": "audio",
		"data": audioData{
			Data:  audioBinaryData,
			Index: idx,
		},
	}
	*audioWasReceived = true
	return true
}
