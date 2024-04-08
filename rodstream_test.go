package rodstream_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	rodstream "github.com/navicstein/rod-stream"
)

func TestMustPrepareLauncher(t *testing.T) {
	var l = rodstream.MustPrepareLauncher(rodstream.LauncherArgs{
		UserMode: false,
	})

	var extensionId []string

	if value, ok := l.Flags["whitelisted-extension-id"]; ok {
		extensionId = value
	} else if value, ok := l.Flags["allowlisted-extension-id"]; ok {
		extensionId = value
	} else {
		t.Error("Neither whitelisted-extension-id nor allowlisted-extension-id is set")
	}

	if extensionId[0] != rodstream.ExtensionId {
		t.Errorf("Extension is invalid")
	}

}

func TestMustCreatePage(t *testing.T) {
	browser := createBrowser()
	pageInfo := rodstream.MustCreatePage(browser)
	if pageInfo.CapturePage.MustInfo().Title != "Video Streamer" {
		t.Errorf("Page title is invalid, got '%s'", pageInfo.CapturePage.MustInfo().Title)
	}

}

func TestMustGetStream(t *testing.T) {
	url := "https://www.youtube.com/watch?v=Jl8iYAo90pE"
	browser := createBrowser()
	constraints := &rodstream.StreamConstraints{
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

	if err := rodstream.MustGetStream(pageInfo, *constraints, streamCh); err != nil {
		log.Panicln(err)
	}

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

func createBrowser() *rod.Browser {
	path, _ := launcher.LookPath()

	var l = rodstream.MustPrepareLauncher(rodstream.LauncherArgs{
		UserMode: false,
	}).
		Bin(path).
		MustLaunch()

	browser := rod.New().ControlURL(l).
		NoDefaultDevice().
		MustConnect()

	return browser
}
