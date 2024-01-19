package communicate

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lib-x/edgetts/internal/businessConsts"
	"github.com/lib-x/edgetts/internal/communicateOption"
	"github.com/lib-x/edgetts/internal/validate"
	"golang.org/x/net/proxy"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	ssmlHeaderTemplate = "X-RequestId:%s\r\nContent-Type:application/ssml+xml\r\nX-Timestamp:%sZ\r\nPath:ssml\r\n\r\n"
	ssmlTemplate       = "<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='en-US'><voice name='%s'><prosody pitch='+0Hz' rate='%s' volume='%s'>%s</prosody></voice></speak>"
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
	text                string
	voice               string
	voiceLanguageRegion string
	rate                string
	volume              string

	httpProxy        string
	socket5Proxy     string
	socket5ProxyUser string
	socket5ProxyPass string
	op               chan map[string]interface{}

	AudioDataIndex int
}

type TextEntry struct {
	Text         string `json:"text"`
	Length       int64  `json:"Length"`
	BoundaryType string `json:"BoundaryType"`
}
type DataEntry struct {
	Offset   int       `json:"Offset"`
	Duration int       `json:"Duration"`
	Text     TextEntry `json:"text"`
}
type MetaDataEntry struct {
	Type string    `json:"Type"`
	Data DataEntry `json:"Data"`
}
type AudioData struct {
	Data  []byte
	Index int
}

type UnknownResponse struct {
	Message string
}

type UnexpectedResponse struct {
	Message string
}

type NoAudioReceived struct {
	Message string
}

type WebSocketError struct {
	Message string
}

func NewCommunicate(text string, options ...communicateOption.Option) (*Communicate, error) {
	opts := &communicateOption.CommunicateOption{}
	for _, optFn := range options {
		optFn(opts)
	}
	opts.ApplyDefaultOption()

	if err := validate.WithCommunicateOption(opts); err != nil {
		return nil, err
	}
	return &Communicate{
		text:                text,
		voice:               opts.Voice,
		voiceLanguageRegion: opts.VoiceLangRegion,
		rate:                opts.Rate,
		volume:              opts.Volume,
		httpProxy:           opts.HttpProxy,
		socket5Proxy:        opts.Socket5Proxy,
		socket5ProxyUser:    opts.Socket5ProxyUser,
		socket5ProxyPass:    opts.Socket5ProxyPass,
	}, nil
}

// WriteStreamTo  write audio stream to io.WriteCloser
func (c *Communicate) WriteStreamTo(rc io.Writer) error {
	op, err := c.stream()
	if err != nil {
		return err
	}
	solveCount := 0
	audioData := make([][][]byte, c.AudioDataIndex)
	for i := range op {
		if _, ok := i["end"]; ok {
			solveCount++
			if solveCount == c.AudioDataIndex {
				break
			}
		}
		t, ok := i["type"]
		if ok && t == "audio" {
			data := i["data"].(AudioData)
			audioData[data.Index] = append(audioData[data.Index], data.Data)
		}
		e, ok := i["error"]
		if ok {
			fmt.Printf("has error err: %v\n", e)
		}
	}
	for _, v := range audioData {
		for _, data := range v {
			rc.Write(data)
		}
	}
	return nil
}

func (c *Communicate) CloseOutput() {
	close(c.op)
}

func makeHeaders() http.Header {
	header := make(http.Header)
	header.Set("Pragma", "no-cache")
	header.Set("Cache-Control", "no-cache")
	header.Set("Origin", "chrome-extension://jdiccldimpdaibmpdkjnbmckianbfold")
	header.Set("Accept-Encoding", "gzip, deflate, br")
	header.Set("Accept-Language", "en-US,en;q=0.9")
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36 Edg/91.0.864.41")
	return header
}

func (c *Communicate) stream() (<-chan map[string]interface{}, error) {
	texts := splitTextByByteLength(
		escape(removeIncompatibleCharacters(c.text)),
		calculateMaxMessageSize(c.voice, c.rate, c.volume),
	)
	c.AudioDataIndex = len(texts)

	finalUtterance := make(map[int]int)
	prevIdx := -1
	shiftTime := -1

	output := make(chan map[string]interface{})

	for idx, text := range texts {
		wsURL := businessConsts.EdgeWssEndpoint + "&ConnectionId=" + generateConnectID()
		dialer := websocket.Dialer{}
		// http proxy
		if c.httpProxy != "" {
			proxyUrl, err := url.Parse(c.httpProxy)
			if err != nil {
				log.Println("proxy url parse error", err)
			} else {
				dialer.Proxy = http.ProxyURL(proxyUrl)
			}
		}
		// socket5 proxy
		if c.socket5Proxy != "" {
			var auth *proxy.Auth
			if c.socket5ProxyUser != "" || c.socket5ProxyPass != "" {
				auth = &proxy.Auth{User: c.socket5ProxyUser, Password: c.socket5ProxyPass}
			}
			socket5ProxyDialer, err := proxy.SOCKS5("tcp", c.socket5Proxy, auth, proxy.Direct)
			if err == nil {
				dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
					return socket5ProxyDialer.Dial(network, address)
				}
				dialer.NetDialContext = dialContext
			}
		}

		conn, _, err := dialer.Dial(wsURL, communicateHeader)
		if err != nil {
			return nil, err
		}

		// download indicates whether we should be expecting audio data,
		// this is so what we avoid getting binary data from the websocket
		// and falsely thinking it's audio data.
		downloadAudio := false

		// audioWasReceived indicates whether we have received audio data
		// from the websocket. This is so we can raise an exception if we
		// don't receive any audio data.
		audioWasReceived := false

		// Each message needs to have the proper date.
		date := currentTimeInMST()

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
		err = conn.WriteMessage(websocket.TextMessage, []byte(
			"X-Timestamp:"+date+"\r\n"+
				"Content-Type:application/json; charset=utf-8\r\n"+
				"Path:speech.config\r\n\r\n"+
				`{"context":{"synthesis":{"audio":{"metadataoptions":{"sentenceBoundaryEnabled":false,"wordBoundaryEnabled":true},"outputFormat":"audio-24khz-48kbitrate-mono-mp3"}}}}`+"\r\n",
		))
		if err != nil {
			conn.Close()
			return nil, err
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte(
			ssmlHeadersAppendExtraData(
				generateConnectID(),
				date,
				makeSsml(string(text), c.voice, c.rate, c.volume),
			),
		))
		if err != nil {
			conn.Close()
			return nil, err
		}

		go func(idx int) {
			defer conn.Close()
			defer func() {
				if err := recover(); err != nil {
					fmt.Printf("communicate.Stream recovered from panic: %v stack: %s", err, string(debug.Stack()))
				}
			}()

			for {
				msgType, message, err := conn.ReadMessage()
				if err != nil {
					if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
						// WebSocket error
						output <- map[string]interface{}{
							"error": WebSocketError{Message: err.Error()},
						}
					}
					break
				}

				if msgType == websocket.TextMessage {
					parameters, data := getHeadersAndData(message)
					path := parameters["Path"]
					if path == "turn.start" {
						downloadAudio = true
					} else if path == "turn.end" {
						output <- map[string]interface{}{
							"end": "",
						}
						downloadAudio = false
						break // End of audio data
					} else if path == "audio.metadata" {
						var metadata struct {
							Metadata []MetaDataEntry `json:"Metadata"`
						}
						err := json.Unmarshal(data, &metadata)
						if err != nil {
							msg := fmt.Sprintf("err=%s, data=%s", err.Error(), string(data))
							output <- map[string]interface{}{
								"error": UnknownResponse{Message: msg},
							}
							break
						}

						for _, metaObj := range metadata.Metadata {
							metaType := metaObj.Type
							if idx != prevIdx {
								shiftTime = sum(idx, finalUtterance)
								prevIdx = idx
							}
							if metaType == "WordBoundary" {
								finalUtterance[idx] = metaObj.Data.Offset + metaObj.Data.Duration + 8_750_000
								output <- map[string]interface{}{
									"type":     metaType,
									"offset":   metaObj.Data.Offset + shiftTime,
									"duration": metaObj.Data.Duration,
									"text":     metaObj.Data.Text,
								}
							} else if metaType == "SessionEnd" {
								continue
							} else {
								output <- map[string]interface{}{
									"error": UnknownResponse{Message: "Unknown metadata type: " + metaType},
								}
								break
							}
						}
					} else if path == "response" {
						// Do nothing
					} else {
						output <- map[string]interface{}{
							"error": UnknownResponse{Message: "The response from the service is not recognized.\n" + string(message)},
						}
						break
					}
				} else if msgType == websocket.BinaryMessage {
					if !downloadAudio {
						output <- map[string]interface{}{
							"error": UnknownResponse{"We received a binary message, but we are not expecting one."},
						}
					}

					if len(message) < 2 {
						output <- map[string]interface{}{
							"error": UnknownResponse{"We received a binary message, but it is missing the header length."},
						}
					}

					headerLength := int(binary.BigEndian.Uint16(message[:2]))
					if len(message) < headerLength+2 {
						output <- map[string]interface{}{
							"error": UnknownResponse{"We received a binary message, but it is missing the audio data."},
						}
					}

					audioData := message[headerLength+2:]
					output <- map[string]interface{}{
						"type": "audio",
						"data": AudioData{
							Data:  audioData,
							Index: idx,
						},
					}
					audioWasReceived = true
				} else {
					if message != nil {
						output <- map[string]interface{}{
							"error": WebSocketError{
								Message: string(message),
							},
						}
					} else {
						output <- map[string]interface{}{
							"error": WebSocketError{
								Message: "Unknown error",
							},
						}
					}
				}

			}

			if !audioWasReceived {
				output <- map[string]interface{}{
					"error": NoAudioReceived{Message: "No audio was received. Please verify that your parameters are correct."},
				}
			}
		}(idx)
	}
	c.op = output
	return output, nil
}

func sum(idx int, m map[int]int) int {
	sum := 0
	for i := 0; i < idx; i++ {
		sum += m[i]
	}
	return sum

}

func getHeadersAndData(data []byte) (map[string]string, []byte) {
	headers := make(map[string]string)

	headerEndIndex := bytes.Index(data, []byte("\r\n\r\n"))
	if headerEndIndex == -1 {
		panic("Invalid data format")
	}

	headerLines := bytes.Split(data[:headerEndIndex], []byte("\r\n"))
	for _, line := range headerLines {
		header := bytes.SplitN(line, []byte(":"), 2)
		if len(header) == 2 {
			key := string(bytes.TrimSpace(header[0]))
			value := string(bytes.TrimSpace(header[1]))
			headers[key] = value
		}
	}

	return headers, data[headerEndIndex+4:]
}

func removeIncompatibleCharacters(str string) string {
	chars := []rune(str)

	for i, char := range chars {
		code := int(char)
		if (0 <= code && code <= 8) || (11 <= code && code <= 12) || (14 <= code && code <= 31) {
			chars[i] = ' '
		}
	}

	return string(chars)
}

func generateConnectID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func splitTextByByteLength(text string, byteLength int) [][]byte {
	var result [][]byte

	textBytes := []byte(text)
	if byteLength <= 0 {
		return result
	}

	for len(textBytes) > byteLength {
		splitAt := bytes.LastIndexByte(textBytes[:byteLength], ' ')
		if splitAt == -1 {
			splitAt = byteLength
		} else {
			splitAt++
		}

		for bytes.Contains(textBytes[:splitAt], []byte("&")) && !bytes.Contains(textBytes[:splitAt], []byte(";")) {
			ampersandIndex := bytes.LastIndexByte(textBytes[:splitAt], '&')
			if semicolonIndex := bytes.IndexByte(textBytes[ampersandIndex:splitAt], ';'); semicolonIndex != -1 {
				break
			}

			splitAt = ampersandIndex - 1
			if splitAt < 0 {
				panic(errors.New("maximum byte length is too small or invalid text"))
			}
			if splitAt == 0 {
				break
			}
		}

		trimmedText := bytes.TrimSpace(textBytes[:splitAt])
		if len(trimmedText) > 0 {
			result = append(result, trimmedText)
		}
		if splitAt == 0 {
			splitAt = 1
		}
		textBytes = textBytes[splitAt:]
	}

	trimmedText := bytes.TrimSpace(textBytes)
	if len(trimmedText) > 0 {
		result = append(result, trimmedText)
	}

	return result
}

func makeSsml(text string, voice string, rate string, volume string) string {
	ssml := fmt.Sprintf(ssmlTemplate,
		voice,
		rate,
		volume,
		text)
	return ssml
}

func currentTimeInMST() string {
	// Use time.FixedZone to represent a fixed timezone offset of 0 (UTC)
	zone := time.FixedZone("UTC", 0)
	now := time.Now().In(zone)
	return now.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
}

func ssmlHeadersAppendExtraData(requestID string, timestamp string, ssml string) string {
	headers := fmt.Sprintf(ssmlHeaderTemplate,
		requestID,
		timestamp,
	)
	return headers + ssml
}

func calculateMaxMessageSize(voice string, rate string, volume string) int {
	websocketMaxSize := 1 << 16
	overheadPerMessage := len(ssmlHeadersAppendExtraData(generateConnectID(), currentTimeInMST(), makeSsml("", voice, rate, volume))) + 50
	return websocketMaxSize - overheadPerMessage
}

func escape(data string) string {
	// Must do ampersand first
	entities := make(map[string]string)
	data = html.EscapeString(data)
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
