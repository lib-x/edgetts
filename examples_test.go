package edgetts_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/lib-x/edgetts"
)

func ExampleSave() {
	err := edgetts.Save(
		context.Background(),
		"hello world",
		"hello.mp3",
		edgetts.WithVoice("en-US-GuyNeural"),
	)
	if err != nil {
		fmt.Println(err)
	}
}

func ExampleBytes() {
	data, err := edgetts.Bytes(
		context.Background(),
		"hello world",
		edgetts.WithVoice("en-US-GuyNeural"),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = data
}

func ExampleBytesSSML() {
	ssml := `<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="en-US"><voice name="en-US-GuyNeural"><prosody rate="+10%">hello world</prosody></voice></speak>`
	data, err := edgetts.BytesSSML(context.Background(), ssml)
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = data
}

func ExampleClient_WriteTo() {
	client := edgetts.New(edgetts.WithVoice("en-US-GuyNeural"))
	var buf bytes.Buffer

	_, err := client.WriteTo(context.Background(), "hello world", &buf)
	if err != nil {
		fmt.Println(err)
	}
}

func ExampleClient_WriteSSMLTo() {
	client := edgetts.New(edgetts.WithVoice("en-US-GuyNeural"))
	var buf bytes.Buffer
	ssml := `<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="en-US"><voice name="en-US-GuyNeural"><prosody pitch="+5Hz">hello world</prosody></voice></speak>`

	_, err := client.WriteSSMLTo(context.Background(), ssml, &buf)
	if err != nil {
		fmt.Println(err)
	}
}

func ExampleClient_Stream() {
	client := edgetts.New(edgetts.WithVoice("en-US-GuyNeural"))
	stream, err := client.Stream(context.Background(), "hello world")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stream.Close()

	_, _ = io.Copy(os.Stdout, stream)
}

func ExampleClient_StreamSSML() {
	client := edgetts.New()
	ssml := `<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="en-US"><voice name="en-US-GuyNeural"><prosody volume="+10%">hello world</prosody></voice></speak>`
	stream, err := client.StreamSSML(context.Background(), ssml)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stream.Close()

	_, _ = io.Copy(os.Stdout, stream)
}

func ExampleClient_Do_text() {
	client := edgetts.New(edgetts.WithVoice("en-US-GuyNeural"))
	req := edgetts.Text("hello world", edgetts.WithRate("+15%"))
	data, err := client.Do(context.Background(), req)
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = data
}

func ExampleClient_Do_ssml() {
	client := edgetts.New()
	ssml := `<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="en-US"><voice name="en-US-GuyNeural"><prosody rate="+20%">hello world</prosody></voice></speak>`
	data, err := client.Do(context.Background(), edgetts.SSML(ssml))
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = data
}

func ExampleClient_SaveBatch() {
	client := edgetts.New()
	results, err := client.SaveBatch(context.Background(), "out", []edgetts.BatchItem{
		{Name: "a.mp3", Request: edgetts.Text("你好", edgetts.WithVoice("zh-CN-XiaoxiaoNeural"))},
		{Name: "b.mp3", Request: edgetts.Text("hello", edgetts.WithVoice("en-US-GuyNeural"))},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = results
}

func ExampleClient_WriteZIP() {
	client := edgetts.New()
	f, err := os.Create("tts.zip")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	ssml := `<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="en-US"><voice name="en-US-GuyNeural"><prosody rate="+10%">hello world</prosody></voice></speak>`
	err = client.WriteZIP(context.Background(), f, []edgetts.BatchItem{
		{Name: "a.mp3", Request: edgetts.Text("你好", edgetts.WithVoice("zh-CN-XiaoxiaoNeural"))},
		{Name: "b.mp3", Request: edgetts.SSML(ssml)},
	}, map[string]any{"source": "demo"})
	if err != nil {
		fmt.Println(err)
	}
}

func ExampleClient_FindVoice() {
	client := edgetts.New()
	voice, err := client.FindVoice(context.Background(), edgetts.VoiceFilter{
		Locale: "zh-CN",
		Gender: "Female",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = voice
}

func ExampleClient_Stream_httpHandler() {
	client := edgetts.New(edgetts.WithVoice("en-US-GuyNeural"))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream, err := client.Stream(r.Context(), "hello from streaming tts")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer stream.Close()

		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = io.Copy(w, stream)
	})

	_ = handler
}

func ExampleText() {
	req := edgetts.Text("hello world", edgetts.WithVoice("en-US-GuyNeural"))
	fmt.Println(req.Type == edgetts.InputText)
	// Output: true
}

func ExampleSSML() {
	req := edgetts.SSML(`<speak version="1.0"></speak>`)
	fmt.Println(req.Type == edgetts.InputSSML)
	// Output: true
}
