package edgetts

// InputType identifies the input payload type.
type InputType int

const (
	InputText InputType = iota
	InputSSML
)

// Request describes one synthesis request.
type Request struct {
	Input   string
	Type    InputType
	Options []Option
}

// Text creates a text synthesis request.
func Text(input string, opts ...Option) Request {
	return Request{Input: input, Type: InputText, Options: opts}
}

// SSML creates an SSML synthesis request.
func SSML(input string, opts ...Option) Request {
	return Request{Input: input, Type: InputSSML, Options: opts}
}

// BatchItem describes one named batch synthesis item.
type BatchItem struct {
	Name    string
	Request Request
}

// BatchResult contains one batch item result.
type BatchResult struct {
	Name  string
	Bytes []byte
	N     int64
	Err   error
}

// VoiceFilter filters voices.
type VoiceFilter struct {
	Name           string
	ShortName      string
	Locale         string
	Gender         string
	SuggestedCodec string
	Status         string
	Language       string
}
