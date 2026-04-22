# edgetts

[![Go Reference](https://pkg.go.dev/badge/github.com/lib-x/edgetts.svg)](https://pkg.go.dev/github.com/lib-x/edgetts)
[![Release](https://img.shields.io/github/v/release/lib-x/edgetts)](https://github.com/lib-x/edgetts/releases)
[![CI](https://github.com/lib-x/edgetts/actions/workflows/release.yml/badge.svg)](https://github.com/lib-x/edgetts/actions/workflows/release.yml)
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
        "你好，世界。",
        "hello.mp3",
        edgetts.WithVoice("zh-CN-XiaoxiaoNeural"),
    )
    if err != nil {
        panic(err)
    }
}
```

### Reuse a client

```go
client := edgetts.New(
    edgetts.WithVoice("zh-CN-XiaoxiaoNeural"),
    edgetts.WithRate("+10%"),
)

data, err := client.Bytes(context.Background(), "这是一段语音")
```

## Runnable demo

仓库内提供了一个可直接运行的 demo：

```bash
go run ./cmd/demo -text "你好，世界" -voice zh-CN-XiaoxiaoNeural -output hello.mp3
```

使用流式输出写文件：

```bash
go run ./cmd/demo -text "hello world" -voice en-US-GuyNeural -output hello.mp3 -stream
```

使用 SSML：

```bash
go run ./cmd/demo -type ssml -text '<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="en-US"><voice name="en-US-GuyNeural"><prosody rate="+10%">hello world</prosody></voice></speak>' -output hello.mp3
```

如果不传 `-output`，demo 会直接生成音频并打印字节数。

## Package-level convenience API

适合一次性调用。

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

适合复用默认配置、服务端场景和批量任务。

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
    {Name: "a.mp3", Request: edgetts.Text("你好", edgetts.WithVoice("zh-CN-XiaoxiaoNeural"))},
    {Name: "b.mp3", Request: edgetts.Text("hello", edgetts.WithVoice("en-US-GuyNeural"))},
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
    {Name: "a.mp3", Request: edgetts.Text("你好", edgetts.WithVoice("zh-CN-XiaoxiaoNeural"))},
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
    Locale: "zh-CN",
    Gender: "Female",
})
```

### Find first matching voice

```go
voice, err := client.FindVoice(ctx, edgetts.VoiceFilter{
    ShortName: "zh-CN-XiaoxiaoNeural",
})
```

## Options

- `WithVoice`
- `WithVoiceLangRegion`
- `WithPitch`
- `WithRate`
- `WithVolume`
- `WithHTTPProxy` / `WithHttpProxy`
- `WithSOCKS5Proxy` / `WithSocket5Proxy`
- `WithInsecureSkipVerify`

## Migration guide

旧 `Speech` API 仍可用，但已经不再推荐。迁移建议如下：

| 旧用法 | 新用法 |
| --- | --- |
| `NewSpeech(opts...)` | `client := edgetts.New(opts...)` |
| `speech.AddSingleTask(text, w); speech.StartTasks()` | `client.WriteTo(ctx, text, w)` |
| `speech.AddSingleTask(text, file); speech.StartTasks()` | `client.Save(ctx, text, path)` |
| `speech.GetVoiceList()` | `client.Voices(ctx)` |
| `AddPackTask(...)` | `client.SaveBatch(...)` 或 `client.WriteZIP(...)` |
| 文本任务 + 自定义入口 | `client.Do(edgetts.Text(...))` |
| SSML 高级场景 | `client.Do(edgetts.SSML(...))` / `client.StreamSSML(...)` |

### Migration example

旧写法：

```go
speech, err := edgetts.NewSpeech(edgetts.WithVoice("zh-CN-XiaoxiaoNeural"))
if err != nil {
    panic(err)
}

file, err := os.Create("hello.mp3")
if err != nil {
    panic(err)
}
defer file.Close()

if err := speech.AddSingleTask("你好，世界", file); err != nil {
    panic(err)
}
if err := speech.StartTasks(); err != nil {
    panic(err)
}
```

新写法：

```go
client := edgetts.New(edgetts.WithVoice("zh-CN-XiaoxiaoNeural"))
if err := client.Save(context.Background(), "你好，世界", "hello.mp3"); err != nil {
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
