package communicate

import "encoding/xml"

// reference document at: https://learn.microsoft.com/en-us/azure/ai-services/speech-service/speech-synthesis-markup-structure

type Speak struct {
	XMLName         xml.Name        `xml:"speak"`
	Version         string          `xml:"version,attr"`
	Xmlns           string          `xml:"xmlns,attr"`
	Mstts           string          `xml:"mstts,attr"`
	Lang            string          `xml:"xml:lang,attr"`
	Backgroundaudio Backgroundaudio `xml:"mstts:backgroundaudio"`
	Voice           Voice           `xml:"voice"`
}

type Backgroundaudio struct {
	Src     string `xml:"src,attr"`
	Volume  string `xml:"volume,attr"`
	Fadein  string `xml:"fadein,attr"`
	Fadeout string `xml:"fadeout,attr"`
}

type Voice struct {
	Name     string   `xml:"name,attr"`
	Effect   string   `xml:"effect,attr"`
	Audio    Audio    `xml:"audio"`
	Bookmark string   `xml:"bookmark,omitempty"`
	Break    Break    `xml:"break,omitempty"`
	Emphasis Emphasis `xml:"emphasis,omitempty"`
	Lang     Lang     `xml:"lang"`
	Lexicon  Lexicon  `xml:"lexicon,omitempty"`
	Math     string   `xml:"math,omitempty"`
	Mstts    Mstts    `xml:"mstts,omitempty"`
	P        string   `xml:"p,omitempty"`
	Phoneme  Phoneme  `xml:"phoneme,omitempty"`
	Prosody  Prosody  `xml:"prosody"`
	SayAs    SayAs    `xml:"say-as,omitempty"`
	Sub      string   `xml:"sub,omitempty"`
}

type Audio struct {
	Src string `xml:"src"`
}

type Break struct {
	Strength string `xml:"strength,attr"`
	Time     string `xml:"time,attr"`
}

type Emphasis struct {
	Level string `xml:"level,attr"`
}

type Lang struct {
	XmlLang string `xml:"xml:lang,attr"`
}

type Lexicon struct {
	URI string `xml:"uri,attr"`
}

type Mstts struct {
	Backgroundaudio string       `xml:"backgroundaudio"`
	Ttsembedding    TtsEmbedding `xml:"ttsembedding"`
	ExpressAs       ExpressAs    `xml:"express-as"`
	Silence         Silence      `xml:"silence"`
	Viseme          Viseme       `xml:"viseme"`
	Audioduration   string       `xml:"audioduration"`
}

type TtsEmbedding struct {
	SpeakerProfileId string `xml:"speakerProfileId,attr"`
}

type ExpressAs struct {
	Style       string `xml:"style,attr"`
	Styledegree string `xml:"styledegree,attr"`
	Role        string `xml:"role,attr"`
}

type Silence struct {
	Type  string `xml:"type,attr"`
	Value string `xml:"value,attr"`
}

type Viseme struct {
	Type string `xml:"type,attr"`
}

type Phoneme struct {
	Alphabet string `xml:"alphabet,attr"`
	Ph       string `xml:"ph,attr"`
}

type Prosody struct {
	Pitch   string `xml:"pitch,attr"`
	Contour string `xml:"contour,attr,omitempty"`
	Range   string `xml:"range,attr,omitempty"`
	Rate    string `xml:"rate,attr"`
	Volume  string `xml:"volume,attr"`
}

type SayAs struct {
	InterpretAs string `xml:"interpret-as,attr"`
	Format      string `xml:"format,attr"`
	Detail      string `xml:"detail,attr"`
}
