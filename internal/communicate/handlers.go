package communicate

import (
	"encoding/binary"
	"github.com/gorilla/websocket"
	"log"
)

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
		meta, err := metaDataContextFrom(data)
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
