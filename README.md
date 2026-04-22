# edgetts

[![Go Reference](https://pkg.go.dev/badge/github.com/lib-x/edgetts.svg)](https://pkg.go.dev/github.com/lib-x/edgetts)
[![Release](https://img.shields.io/github/v/release/lib-x/edgetts)](https://github.com/lib-x/edgetts/releases)
[![CI](https://github.com/lib-x/edgetts/actions/workflows/ci.yml/badge.svg)](https://github.com/lib-x/edgetts/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/lib-x/edgetts)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/lib-x/edgetts)](https://goreportcard.com/report/github.com/lib-x/edgetts)

English | [简体中文](README.zh-CN.md)

## Documentation

- English guide: `README.md`
- Chinese guide: [`README.zh-CN.md`](README.zh-CN.md)
- API reference: https://pkg.go.dev/github.com/lib-x/edgetts
- Releases: https://github.com/lib-x/edgetts/releases

A Go library for Microsoft Edge TTS with a simpler API for common use cases.

## Highlights

- Client-based API for reusable configuration.
- Package-level convenience functions for one-off calls.
- Text and SSML are first-class, symmetric inputs.
- Output to `[]byte`, file, `io.Writer`, stream, directory, and ZIP.
- Voice listing and filtering helpers.
- Legacy `Speech` API kept as a deprecated compatibility layer.

## Install

```bash
go get github.com/lib-x/edgetts
```

## Quick start

### Save text to mp3

```go
package main

import (
    "context"

    "github.com/lib-x/edgetts"
)

func main() {
    err := edgetts.Save(
        context.Background(),
        "Hello, world.",
        "hello.mp3",
        edgetts.WithVoice("en-US-GuyNeural"),
    )
    if err != nil {
        panic(err)
    }
}
```

### Reuse a client

```go
client := edgetts.New(
    edgetts.WithVoice("en-US-GuyNeural"),
    edgetts.WithRate("+10%"),
)

data, err := client.Bytes(context.Background(), "This is a reusable client example.")
```

## Runnable demo

A runnable demo is included in this repository:

```bash
go run ./cmd/demo -text "hello world" -voice en-US-GuyNeural -output hello.mp3
```

Write to a file through streaming output:

```bash
go run ./cmd/demo -text "hello world" -voice en-US-GuyNeural -output hello.mp3 -stream
```

Use SSML input:

```bash
go run ./cmd/demo -type ssml -text '<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="en-US"><voice name="en-US-GuyNeural"><prosody rate="+10%">hello world</prosody></voice></speak>' -output hello.mp3
```

If `-output` is omitted, the demo generates audio in memory and prints the byte size.

## Package-level convenience API

Best for one-off calls.

### Text to bytes

```go
data, err := edgetts.Bytes(
    ctx,
    "hello world",
    edgetts.WithVoice("en-US-GuyNeural"),
)
```

### SSML to bytes

```go
ssml := `<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="en-US"><voice name="en-US-GuyNeural"><prosody rate="+10%">hello world</prosody></voice></speak>`
data, err := edgetts.BytesSSML(ctx, ssml)
```

### Text directly to file

```go
err := edgetts.Save(ctx, "hello world", "hello.mp3", edgetts.WithVoice("en-US-GuyNeural"))
```

### SSML directly to file

```go
err := edgetts.SaveSSML(ctx, ssml, "hello.mp3")
```

## Client API

Best for reusable defaults, service-side usage, and batch workflows.

### Create a reusable client

```go
client := edgetts.New(
    edgetts.WithVoice("en-US-GuyNeural"),
    edgetts.WithRate("+15%"),
)
```

### Text / SSML with explicit request objects

```go
textReq := edgetts.Text("hello world", edgetts.WithVoice("en-US-GuyNeural"))
textData, err := client.Do(ctx, textReq)

ssmlReq := edgetts.SSML(`<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="en-US"><voice name="en-US-GuyNeural"><prosody pitch="+5Hz">hello world</prosody></voice></speak>`)
ssmlData, err := client.Do(ctx, ssmlReq)

_ = textData
_ = ssmlData
```

## Output shapes

### Write text to an `io.Writer`

```go
var buf bytes.Buffer
_, err := client.WriteTo(ctx, "hello world", &buf)
```

### Write SSML to an `io.Writer`

```go
var buf bytes.Buffer
_, err := client.WriteSSMLTo(ctx, ssml, &buf)
```

### Stream text audio

```go
stream, err := client.Stream(ctx, "hello world")
if err != nil {
    return err
}
defer stream.Close()

_, err = io.Copy(w, stream)
```

### Stream SSML audio

```go
stream, err := client.StreamSSML(ctx, ssml)
if err != nil {
    return err
}
defer stream.Close()

_, err = io.Copy(w, stream)
```

### Stream directly in an HTTP handler

```go
client := edgetts.New(edgetts.WithVoice("en-US-GuyNeural"))

http.HandleFunc("/tts", func(w http.ResponseWriter, r *http.Request) {
    stream, err := client.Stream(r.Context(), "hello from streaming tts")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer stream.Close()

    w.Header().Set("Content-Type", "audio/mpeg")
    _, _ = io.Copy(w, stream)
})
```

### Save SSML directly to file

```go
err := client.SaveSSML(ctx, ssml, "speech.mp3")
```

## Batch

### Save batch into a directory

```go
results, err := client.SaveBatch(ctx, "out", []edgetts.BatchItem{
    {Name: "a.mp3", Request: edgetts.Text("hello", edgetts.WithVoice("en-US-GuyNeural"))},
    {Name: "b.mp3", Request: edgetts.Text("welcome", edgetts.WithVoice("en-US-JennyNeural"))},
})
```

Each `BatchResult` contains:

- `Name`
- `Bytes`
- `N`
- `Err`

### Write batch into a zip file

```go
f, _ := os.Create("tts.zip")
defer f.Close()

err := client.WriteZIP(ctx, f, []edgetts.BatchItem{
    {Name: "a.mp3", Request: edgetts.Text("hello", edgetts.WithVoice("en-US-GuyNeural"))},
    {Name: "b.mp3", Request: edgetts.SSML(ssml)},
}, map[string]any{"source": "demo"})
```

## Voices

### List voices

```go
voices, err := client.Voices(ctx)
```

### Filter voices

```go
matches := edgetts.FilterVoices(voices, edgetts.VoiceFilter{
    Locale: "en-US",
    Gender: "Female",
})
```

### Find the first matching voice

```go
voice, err := client.FindVoice(ctx, edgetts.VoiceFilter{
    ShortName: "en-US-GuyNeural",
})
```

## Runnable demo flags

```bash
go run ./cmd/demo -h
```

Main flags:

- `-type` (`text` or `ssml`)
- `-text`
- `-output`
- `-voice`
- `-rate`
- `-pitch`
- `-volume`
- `-stream`

## Migration guide

The old `Speech` API still works, but it is no longer the recommended entry point.

| Old usage | New usage |
| --- | --- |
| `NewSpeech(opts...)` | `client := edgetts.New(opts...)` |
| `speech.AddSingleTask(text, w); speech.StartTasks()` | `client.WriteTo(ctx, text, w)` |
| `speech.AddSingleTask(text, file); speech.StartTasks()` | `client.Save(ctx, text, path)` |
| `speech.GetVoiceList()` | `client.Voices(ctx)` |
| `AddPackTask(...)` | `client.SaveBatch(...)` or `client.WriteZIP(...)` |
| Text tasks with per-call options | `client.Do(edgetts.Text(...))` |
| SSML advanced flows | `client.Do(edgetts.SSML(...))` or `client.StreamSSML(...)` |

### Migration example

Old:

```go
speech, err := edgetts.NewSpeech(edgetts.WithVoice("en-US-GuyNeural"))
if err != nil {
    panic(err)
}

file, err := os.Create("hello.mp3")
if err != nil {
    panic(err)
}
defer file.Close()

if err := speech.AddSingleTask("hello world", file); err != nil {
    panic(err)
}
if err := speech.StartTasks(); err != nil {
    panic(err)
}
```

New:

```go
client := edgetts.New(edgetts.WithVoice("en-US-GuyNeural"))
if err := client.Save(context.Background(), "hello world", "hello.mp3"); err != nil {
    panic(err)
}
```

## Legacy compatibility

The old `Speech` task API still exists as a compatibility wrapper, but new code should prefer `Client` and the package-level helpers.

## References

- https://github.com/rany2/edge-tts
- https://github.com/surfaceyu/edge-tts-go
- https://github.com/pp-group/edge-tts-go
- https://github.com/Migushthe2nd/MsEdgeTTS
- https://gist.github.com/czyt/a2d83de838c9b65ab14fc18136f53bc6
- https://learn.microsoft.com/en-us/azure/ai-services/speech-service/speech-synthesis-markup-voice

## Notes

- `Speech` is still available for compatibility, but new integrations should use `Client`.
- Real network synthesis depends on the upstream Edge TTS endpoint behavior.
