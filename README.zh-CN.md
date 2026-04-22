# edgetts

[![Go Reference](https://pkg.go.dev/badge/github.com/lib-x/edgetts.svg)](https://pkg.go.dev/github.com/lib-x/edgetts)
[![Release](https://img.shields.io/github/v/release/lib-x/edgetts)](https://github.com/lib-x/edgetts/releases)
[![CI](https://github.com/lib-x/edgetts/actions/workflows/ci.yml/badge.svg)](https://github.com/lib-x/edgetts/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/lib-x/edgetts)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/lib-x/edgetts)](https://goreportcard.com/report/github.com/lib-x/edgetts)

[English](README.md) | 简体中文

## 文档导航

- 英文文档：[`README.md`](README.md)
- 中文文档：`README.zh-CN.md`
- API 文档：https://pkg.go.dev/github.com/lib-x/edgetts
- Release 列表：https://github.com/lib-x/edgetts/releases

## 目录

- [特性](#特性)
- [安装](#安装)
- [快速开始](#快速开始)
- [可运行 demo](#可运行-demo)
- [包级便捷 API](#包级便捷-api)
- [Client API](#client-api)
- [输出方式](#输出方式)
- [批量处理](#批量处理)
- [Voices](#voices)
- [Demo 参数](#demo-参数)
- [迁移指南](#迁移指南)
- [兼容说明](#兼容说明)

一个更易用的 Microsoft Edge TTS Go 库，适合单次调用、服务端流式输出和批量生成等场景。

## 特性

- 基于 `Client` 的可复用 API。
- 提供包级便捷函数，适合一次性调用。
- `Text` 和 `SSML` 都是一等输入类型。
- 支持输出到 `[]byte`、文件、`io.Writer`、流、目录和 ZIP。
- 提供 voice 列表与筛选能力。
- 保留旧 `Speech` API 作为弃用兼容层。

## 安装

```bash
go get github.com/lib-x/edgetts
```

## 快速开始

### 保存文本到 mp3

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

### 复用一个 client

```go
client := edgetts.New(
    edgetts.WithVoice("zh-CN-XiaoxiaoNeural"),
    edgetts.WithRate("+10%"),
)

data, err := client.Bytes(context.Background(), "这是一段可复用 client 的示例。")
```

## 可运行 demo

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

如果不传 `-output`，demo 会在内存中生成音频并打印字节数。

## 包级便捷 API

适合一次性调用。

### 文本转 bytes

```go
data, err := edgetts.Bytes(
    ctx,
    "hello world",
    edgetts.WithVoice("en-US-GuyNeural"),
)
```

### SSML 转 bytes

```go
ssml := `<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="en-US"><voice name="en-US-GuyNeural"><prosody rate="+10%">hello world</prosody></voice></speak>`
data, err := edgetts.BytesSSML(ctx, ssml)
```

### 文本直接保存到文件

```go
err := edgetts.Save(ctx, "hello world", "hello.mp3", edgetts.WithVoice("en-US-GuyNeural"))
```

### SSML 直接保存到文件

```go
err := edgetts.SaveSSML(ctx, ssml, "hello.mp3")
```

## Client API

适合复用默认配置、服务端场景和批量任务。

### 创建一个可复用 client

```go
client := edgetts.New(
    edgetts.WithVoice("en-US-GuyNeural"),
    edgetts.WithRate("+15%"),
)
```

### 使用显式 Request 处理 Text / SSML

```go
textReq := edgetts.Text("hello world", edgetts.WithVoice("en-US-GuyNeural"))
textData, err := client.Do(ctx, textReq)

ssmlReq := edgetts.SSML(`<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="en-US"><voice name="en-US-GuyNeural"><prosody pitch="+5Hz">hello world</prosody></voice></speak>`)
ssmlData, err := client.Do(ctx, ssmlReq)

_ = textData
_ = ssmlData
```

## 输出方式

### 写入 `io.Writer`

```go
var buf bytes.Buffer
_, err := client.WriteTo(ctx, "hello world", &buf)
```

### 将 SSML 写入 `io.Writer`

```go
var buf bytes.Buffer
_, err := client.WriteSSMLTo(ctx, ssml, &buf)
```

### 流式输出文本音频

```go
stream, err := client.Stream(ctx, "hello world")
if err != nil {
    return err
}
defer stream.Close()

_, err = io.Copy(w, stream)
```

### 流式输出 SSML 音频

```go
stream, err := client.StreamSSML(ctx, ssml)
if err != nil {
    return err
}
defer stream.Close()

_, err = io.Copy(w, stream)
```

### 在 HTTP handler 中直接流式返回

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

### 直接保存 SSML 到文件

```go
err := client.SaveSSML(ctx, ssml, "speech.mp3")
```

## 批量处理

### 批量保存到目录

```go
results, err := client.SaveBatch(ctx, "out", []edgetts.BatchItem{
    {Name: "a.mp3", Request: edgetts.Text("你好", edgetts.WithVoice("zh-CN-XiaoxiaoNeural"))},
    {Name: "b.mp3", Request: edgetts.Text("hello", edgetts.WithVoice("en-US-GuyNeural"))},
})
```

每个 `BatchResult` 包含：

- `Name`
- `Bytes`
- `N`
- `Err`

### 批量写入 ZIP

```go
f, _ := os.Create("tts.zip")
defer f.Close()

err := client.WriteZIP(ctx, f, []edgetts.BatchItem{
    {Name: "a.mp3", Request: edgetts.Text("你好", edgetts.WithVoice("zh-CN-XiaoxiaoNeural"))},
    {Name: "b.mp3", Request: edgetts.SSML(ssml)},
}, map[string]any{"source": "demo"})
```

## Voices

### 获取 voice 列表

```go
voices, err := client.Voices(ctx)
```

### 筛选 voice

```go
matches := edgetts.FilterVoices(voices, edgetts.VoiceFilter{
    Locale: "zh-CN",
    Gender: "Female",
})
```

### 查找第一个匹配的 voice

```go
voice, err := client.FindVoice(ctx, edgetts.VoiceFilter{
    ShortName: "zh-CN-XiaoxiaoNeural",
})
```

## Demo 参数

```bash
go run ./cmd/demo -h
```

主要参数：

- `-type`（`text` 或 `ssml`）
- `-text`
- `-output`
- `-voice`
- `-rate`
- `-pitch`
- `-volume`
- `-stream`

## 迁移指南

旧 `Speech` API 仍然可用，但不再推荐作为新入口。

| 旧用法 | 新用法 |
| --- | --- |
| `NewSpeech(opts...)` | `client := edgetts.New(opts...)` |
| `speech.AddSingleTask(text, w); speech.StartTasks()` | `client.WriteTo(ctx, text, w)` |
| `speech.AddSingleTask(text, file); speech.StartTasks()` | `client.Save(ctx, text, path)` |
| `speech.GetVoiceList()` | `client.Voices(ctx)` |
| `AddPackTask(...)` | `client.SaveBatch(...)` 或 `client.WriteZIP(...)` |
| 文本任务 + 每次调用单独配置 | `client.Do(edgetts.Text(...))` |
| SSML 高级场景 | `client.Do(edgetts.SSML(...))` 或 `client.StreamSSML(...)` |

### 迁移示例

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

## 兼容说明

旧 `Speech` task API 仍作为兼容包装层存在，但新代码应优先使用 `Client` 和包级便捷函数。

## 参考

- https://github.com/rany2/edge-tts
- https://github.com/surfaceyu/edge-tts-go
- https://github.com/pp-group/edge-tts-go
- https://github.com/Migushthe2nd/MsEdgeTTS
- https://gist.github.com/czyt/a2d83de838c9b65ab14fc18136f53bc6
- https://learn.microsoft.com/en-us/azure/ai-services/speech-service/speech-synthesis-markup-voice

## 说明

- `Speech` 仍然保留用于兼容，但新集成建议使用 `Client`。
- 实际网络合成效果依赖上游 Edge TTS 服务行为。
