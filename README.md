# Overview

An extension of [go-rod](https://github.com/go-rod/rod) to retrieve audio and/or video streams of a page

## Installation

```sh
$ go get github.com/navicstein/go-stream
```

## Usage

The first thing to do is to make sure that we have the lancher ready

```go
package main

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/stealth"

	rodstream "github.com/navicstein/rod-stream"
)

func main() {
	l := rodstream.
		MustPrepareLauncher().
        // .. other options
		Set("no-sandbox").
		Devtools(false).
		Headless(false).
		Bin("/usr/bin/brave-browser").
		MustLaunch()

	browser := rod.New().ControlURL(l).
		NoDefaultDevice().
		MustConnect()
}
```

optional if you don't want to do it yourself, override permissions for URLs

```go
	// Example URLS to grant permissions to
	urls := []string{"https://meet.google.com", "https://zoom.us"}
	if err := rodstream.GrantPermissions(urls, browser); err != nil {
		log.Panicln(err)
	}
```

Create a rod Page

```go

// ... browser
// if using stealth, important to escape bot detection
page := stealth.MustPage(browser)

// Other Proto calls
proto.PageSetAdBlockingEnabled{
	Enabled: true,
}.Call(page)

page.MustNavigate(url)

```

## Getting a stream 


```go
	go func() {
		// ⚠ Note: the page returned from `MustStreamPage` is not navigatable
		// so don't replace MustPage() with this
		extensionTarget := rodstream.MustCreatePage(browser)

		constraints := &rodstream.StreamConstraints{
			Audio:              true,
			Video:              true,
			MimeType:           "video/webm;codecs=vp9,opus",
			AudioBitsPerSecond: 128000,
			VideoBitsPerSecond: 2500000,
			BitsPerSecond:      8000000, // 1080p https://support.google.com/youtube/answer/1722171?hl=en#zippy=%2Cbitrate
			FrameSize:          1000,    // option passed to mediaRecorder.start(frameSize)
		}

		// Example: Saving to filesystem
		videoFile, err := os.Create("./videos/video.webm")
		if err != nil {
			log.Panicln(err)
		}

		// channel to receive the stream
		ch := make(chan string)
		rodstream.MustGetStream(extensionTarget, constraints, ch)

		for b64Str := range ch {
			//[important] remove base64 prefix
			buff := rodstream.Parseb64(b64Str)

			// write to File video
			videoFile.Write(buff)
		}

	}()
```


Piping to FFMPEG

```go

// Example: Piping the stream to ffmpeg
stdin, err := rodstream.GetStdInWriter("./videos/video.mp4")
if err != nil {
    log.Panicln(err)
}

defer func() {
    err := stdin.Close()
    log.Panicln(err)
}()

for b64Str := range ch {
    buff := rodstream.Parseb64(b64Str)
    _, err = stdin.Write(buff)
    log.Panicln(err)
}
```

## Closing a stream

```go
go func() {
	if err := rodstream.MustStopStream(extensionTarget); err != nil {
		log.Panicln(err)
	}
}()

```

## Running headfull in Docker
This plugin requires that you run in a headfull mode, so when using docker 
start rod with 

```Dockerfile
# https://helpmanual.io/help/xvfb-run/
RUN apt update && apt install -y xvfb
CMD xvfb-run --server-args="-screen 0 1024x768x24" ./cmd-name
```