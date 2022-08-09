package rodstream_test

import (
	"log"
	"net/http"
	"testing"

	"github.com/go-rod/rod"
	rodstream "github.com/navicstein/rod-stream"
)

func TestMustPrepareLauncher(t *testing.T) {
	var l = rodstream.MustPrepareLauncher()
	hasExtension := l.Flags["whitelisted-extension-id"]

	if len(hasExtension) == 0 {
		t.Error("whitelisted-extension-id is not set")
	}

	if hasExtension[0] != rodstream.ExtensionId {
		t.Errorf("Extension is invalid")
	}

}

func TestMustCreatePage(t *testing.T) {
	browser := createBrowser()
	extensionPage := rodstream.MustCreatePage(browser)
	if extensionPage.MustInfo().Title != "Video Streamer" {
		t.Errorf("Page title is invalid, got '%s'", extensionPage.MustInfo().Title)
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

	extensionPage := rodstream.MustCreatePage(browser)
	ch := make(chan string, 1024)

	if err := rodstream.MustGetStream(extensionPage, constraints, ch); err != nil {
		log.Panicln(err)
	}

	// get just one buff from channel
	buff := rodstream.Parseb64(<-ch)
	d := http.DetectContentType(buff)
	log.Println("Content-Type:", d)
	if d != "video/webm" {
		t.Errorf("Content type is invalid, got '%s'", d)
	}
}

func createBrowser() *rod.Browser {
	var l = rodstream.MustPrepareLauncher().
		Bin("/usr/bin/google-chrome").
		MustLaunch()

	browser := rod.New().ControlURL(l).
		NoDefaultDevice().
		MustConnect()

	return browser
}
