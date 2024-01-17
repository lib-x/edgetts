package edgetts

import "testing"

func Test_GetVoiceList(t *testing.T) {
	voices, err := listVoices()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("voices: %+v", voices)
}
