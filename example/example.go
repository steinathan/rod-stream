package main

import (
	"log"
	"os"
	"time"

	"github.com/go-rod/rod"
	rodstream "github.com/navicstein/rod-stream"
)

func createBrowser() *rod.Browser {
	var l = rodstream.MustPrepareLauncher(rodstream.LauncherArgs{
		UserMode: false,
	}).
		Bin("/usr/bin/brave-browser").
		MustLaunch()

	browser := rod.New().ControlURL(l).
		NoDefaultDevice().
		MustConnect()

	return browser
}

func main() {
	url := "https://www.youtube.com/watch?v=Jl8iYAo90pE"
	browser := createBrowser()
	constraints := rodstream.StreamConstraints{
		Audio:              true,
		Video:              true,
		MimeType:           "video/webm;codecs=vp9,opus",
		AudioBitsPerSecond: 128000,
		VideoBitsPerSecond: 2500000,
		BitsPerSecond:      8000000,
		FrameSize:          1000,
	}

	page := browser.MustPage()
	page.MustNavigate(url).MustWaitRequestIdle()

	page.MustElement(".ytp-large-play-button").MustClick()

	pageInfo := rodstream.MustCreatePage(browser)
	streamCh := make(chan string, 1024)

	if err := rodstream.MustGetStream(pageInfo, constraints, streamCh); err != nil {
		log.Panicln(err)
	}

	// wait then close
	time.AfterFunc(time.Second*10, func() {
		rodstream.MustStopStream(pageInfo)
		browser.Close()
	})

	videoFile, err := os.Create("/tmp/video-test.webm")
	if err != nil {
		panic(err)
	}

	for x := range streamCh {
		b := rodstream.Parseb64(x)
		videoFile.Write(b)
	}
}
