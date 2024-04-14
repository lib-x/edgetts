package communicate

import "encoding/xml"

type Speak struct {
	XMLName xml.Name `xml:"speak"`
	Version string   `xml:"version,attr"`
	Xmlns   string   `xml:"xmlns,attr"`
	Lang    string   `xml:"xml:lang,attr"`
	Voice   []Voice  `xml:"voice"`
}

type Voice struct {
	Name    string  `xml:"name,attr"`
	Prosody Prosody `xml:"prosody"`
}

type Prosody struct {
	// Contour represents changes in pitch. These changes are represented as an array of targets at specified time
	//positions in the speech output. Sets of parameter pairs define each target. For example:
	//
	//<prosody contour="(0%,+20Hz) (10%,-2st) (40%,+10Hz)">
	//
	//The first value in each set of parameters specifies the location of the pitch change as a percentage of the
	//duration of the text. The second value specifies the amount to raise or lower the pitch by using a relative
	//value or an enumeration value for pitch (see pitch).
	Contour string `xml:"contour,attr,omitempty"`
	//Indicates the baseline pitch for the text. Pitch changes can be applied at the sentence level. The pitch changes
	//should be within 0.5 to 1.5 times the original audio. You can express the pitch as:
	//An absolute value:
	//Expressed as a number followed by "Hz" (Hertz). For example, <prosody pitch="600Hz">some text</prosody>.
	//A relative value:
	//	As a relative number: Expressed as a number preceded by "+" or "-" and followed by "Hz" or "st" that specifies
	//	an amount to change the pitch. For example:
	//	<prosody pitch="+80Hz">some text</prosody> or <prosody pitch="-2st">some text</prosody>.
	//	The "st" indicates the change unit is semitone, which is half of a tone (a half step) on the standard diatonic scale.
	//As a percentage: Expressed as a number preceded by "+" (optionally) or "-" and followed by "%", indicating the
	//relative change. For example: <prosody pitch="50%">some text</prosody> or <prosody pitch="-50%">some text</prosody>.
	// A constant value:
	//	x-low
	//	low
	//	medium
	//	high
	//	x-high
	//	default
	Pitch string `xml:"pitch,attr"`
	// Indicates the speaking rate of the text. Speaking rate can be applied at the word or sentence level. The rate changes
	//should be within 0.5 to 2 times the original audio. You can express rate as:
	//A relative value:
	//	As a relative number: Expressed as a number that acts as a multiplier of the default. For example, a value of 1 results
	//	in no change in the original rate. A value of 0.5 results in a halving of the original rate. A value of 2 results in
	//	twice the original rate.
	//	As a percentage: Expressed as a number preceded by "+" (optionally) or "-" and followed by "%", indicating the relative
	//	change. For example:
	//	<prosody rate="50%">some text</prosody> or <prosody rate="-50%">some text</prosody>.
	//	A constant value:
	//	x-slow
	//	slow
	//	medium
	//	fast
	//	x-fast
	//	default
	Rate string `xml:"rate,attr"`
	// Indicates the volume level of the speaking voice. Volume changes can be applied at the sentence level. You can express
	//the volume as:
	// An absolute value: Expressed as a number in the range of 0.0 to 100.0, from quietest to loudest, such as 75.
	//The default value is 100.0.
	// A relative value:
	// As a relative number: Expressed as a number preceded by "+" or "-" that specifies an amount to change the volume.
	//Examples are +10 or -5.5.
	// As a percentage: Expressed as a number preceded by "+" (optionally) or "-" and followed by "%", indicating the
	//relative change. For example:
	//	<prosody volume="50%">some text</prosody> or <prosody volume="+3%">some text</prosody>.
	//
	//	A constant value:
	//	silent
	//	x-soft
	//	soft
	//	medium
	//	loud
	//	x-loud
	//	default
	Volume string `xml:"volume,attr"`
	Text   string `xml:",chardata"`
}
