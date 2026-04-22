package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/lib-x/edgetts"
)

func main() {
	var (
		inputType = flag.String("type", "text", "input type: text or ssml")
		text      = flag.String("text", "hello world", "text or ssml input")
		output    = flag.String("output", "", "output mp3 file path; if empty, print byte size only")
		voice     = flag.String("voice", "", "voice short name, e.g. zh-CN-XiaoxiaoNeural")
		rate      = flag.String("rate", "", "speech rate, e.g. +10%")
		pitch     = flag.String("pitch", "", "speech pitch, e.g. +5Hz")
		volume    = flag.String("volume", "", "speech volume, e.g. +10%")
		stream    = flag.Bool("stream", false, "use Stream/StreamSSML instead of Save/SaveSSML")
	)
	flag.Parse()

	opts := make([]edgetts.Option, 0, 4)
	if *voice != "" {
		opts = append(opts, edgetts.WithVoice(*voice))
	}
	if *rate != "" {
		opts = append(opts, edgetts.WithRate(*rate))
	}
	if *pitch != "" {
		opts = append(opts, edgetts.WithPitch(*pitch))
	}
	if *volume != "" {
		opts = append(opts, edgetts.WithVolume(*volume))
	}

	client := edgetts.New(opts...)
	ctx := context.Background()

	isSSML := strings.EqualFold(*inputType, "ssml")
	if *output == "" {
		var (
			data []byte
			err  error
		)
		if isSSML {
			data, err = client.BytesSSML(ctx, *text)
		} else {
			data, err = client.Bytes(ctx, *text)
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("generated %d bytes\n", len(data))
		return
	}

	if err := saveToFile(ctx, client, *output, *text, isSSML, *stream); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("saved audio to %s\n", *output)
}

func saveToFile(ctx context.Context, client *edgetts.Client, outputPath, input string, isSSML, stream bool) error {
	if !stream {
		if isSSML {
			return client.SaveSSML(ctx, input, outputPath)
		}
		return client.Save(ctx, input, outputPath)
	}

	var (
		reader io.ReadCloser
		err    error
	)
	if isSSML {
		reader, err = client.StreamSSML(ctx, input)
	} else {
		reader, err = client.Stream(ctx, input)
	}
	if err != nil {
		return err
	}
	defer reader.Close()

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if _, err := io.Copy(file, reader); err != nil {
		if removeErr := os.Remove(outputPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf("copy stream: %w (cleanup failed: %v)", err, removeErr)
		}
		return fmt.Errorf("copy stream: %w", err)
	}
	return nil
}
