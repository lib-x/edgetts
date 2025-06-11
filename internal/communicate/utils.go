package communicate

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
	"unicode"
)

var (
	uuidReplacer   = strings.NewReplacer("-", "")
	escapeReplacer = strings.NewReplacer(">", "&gt;", "<", "&lt;")
)

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

func sumWithMap(idx int, m map[int]int) int {
	if idx <= 0 || m == nil {
		return 0
	}
	sumResult := 0
	for i := 0; i < idx; i++ {
		sumResult += m[i]
	}
	return sumResult
}

func metaDataContextFrom(data []byte) (*metaDataContext, error) {
	metadata := &metaDataContext{}
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
	return uuidReplacer.Replace(uuid.New().String())
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

func currentTimeInMST() string {
	// Use time.FixedZone to represent a fixed timezone offset of 0 (UTC)
	zone := time.FixedZone("UTC", 0)
	now := time.Now().In(zone)
	return now.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
}

func appendRequestContextToSsmlHeaders(requestID string, timestamp string, ssml string) string {
	headers := fmt.Sprintf(
		ssmlHeaderTemplate,
		requestID,
		timestamp,
	)
	return headers + ssml
}

func getMaxMessageSize(pitch, voice string, rate string, volume string) int {
	websocketMaxSize := 1 << 16
	overheadPerMessage := len(appendRequestContextToSsmlHeaders(generateConnectID(), currentTimeInMST(), makeSsml("", pitch, voice, rate, volume))) + 50
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
